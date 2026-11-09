package node

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/infra/dnscheck"
	"github.com/wuekevin/smartcdn/internal/infra/httpclient"
	"github.com/wuekevin/smartcdn/internal/infra/repo"
)

var (
	ErrNodeNotFound  = errors.New("node not found")
	ErrNodeIPTaken   = errors.New("node IP already registered")
	ErrPoolNotFound  = errors.New("pool not found")
	ErrTaskNotFound  = errors.New("task not found")
)

// allowedTaskTypes is the whitelist of tasks Node Agent may execute.
var allowedTaskTypes = map[string]bool{
	"reload_nginx":    true,
	"restart_nginx":   true,
	"clear_cache":     true,
	"check_https":     true,
	"collect_metrics": true,
}

// Service manages self-hosted CDN nodes, pools, and agent tasks.
type Service struct {
	nodes  port.CDNNodeRepository
	agents port.NodeAgentRepository
	tasks  port.NodeTaskRepository
	pools  port.NodePoolRepository
}

func NewService(
	nodes port.CDNNodeRepository,
	agents port.NodeAgentRepository,
	tasks port.NodeTaskRepository,
	pools port.NodePoolRepository,
) *Service {
	return &Service{nodes: nodes, agents: agents, tasks: tasks, pools: pools}
}

// --- Node CRUD ---

type CreateNodeInput struct {
	Name           string
	IP             string
	ISP            string
	Region         string
	OriginAddr     string
	OriginPort     int
	AuthConfirmed  bool // operator must check "I am authorized to use this node"
}

type NodeView struct {
	ID             uint64     `json:"id"`
	Name           string     `json:"name"`
	IP             string     `json:"ip"`
	ISP            string     `json:"isp"`
	Region         string     `json:"region"`
	OriginAddr     string     `json:"origin_addr"`
	OriginPort     int        `json:"origin_port"`
	IsActive       bool       `json:"is_active"`
	AuthToken      string     `json:"auth_token,omitempty"` // shown only at creation
	AuthConfirmed  bool       `json:"auth_confirmed"`
	HTTPSStatus    string     `json:"https_status"`
	HTTPSCode      int        `json:"https_code,omitempty"`
	TLSDomain      string     `json:"tls_domain,omitempty"`
	HTTPSLatencyMs int64      `json:"https_latency_ms,omitempty"`
	HTTPSCheckedAt *time.Time `json:"https_checked_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (s *Service) CreateNode(ctx context.Context, input CreateNodeInput) (*NodeView, error) {
	if input.IP == "" {
		return nil, fmt.Errorf("node IP is required")
	}
	if err := httpclient.CheckIP(input.IP); err != nil {
		return nil, fmt.Errorf("node: SSRF guard rejected IP %q: %w", input.IP, err)
	}
	if _, err := s.nodes.FindByIP(ctx, input.IP); err == nil {
		return nil, ErrNodeIPTaken
	}
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("node: generate auth token: %w", err)
	}
	port_ := input.OriginPort
	if port_ == 0 {
		port_ = 80
	}
	n := &entity.CDNNode{
		Name:          input.Name,
		IP:            input.IP,
		ISP:           input.ISP,
		Region:        input.Region,
		OriginAddr:    input.OriginAddr,
		OriginPort:    port_,
		IsActive:      true,
		AuthToken:     token,
		AuthConfirmed: input.AuthConfirmed,
		HTTPSStatus:   "pending",
	}
	if err := s.nodes.Create(ctx, n); err != nil {
		return nil, fmt.Errorf("node: create: %w", err)
	}
	// Trigger HTTPS check asynchronously (non-blocking).
	go s.runHTTPSCheck(context.Background(), n.ID, input.IP)

	v := toNodeView(n)
	v.AuthToken = token // reveal only at creation
	return v, nil
}

func (s *Service) GetNode(ctx context.Context, id uint64) (*NodeView, error) {
	n, err := s.nodes.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNodeNotFound
		}
		return nil, err
	}
	return toNodeView(n), nil
}

func (s *Service) ListNodes(ctx context.Context, afterID uint64, limit int) ([]NodeView, error) {
	ns, err := s.nodes.List(ctx, afterID, limit)
	if err != nil {
		return nil, err
	}
	views := make([]NodeView, len(ns))
	for i, n := range ns {
		views[i] = *toNodeView(&n)
	}
	return views, nil
}

func (s *Service) DeleteNode(ctx context.Context, id uint64) error {
	if _, err := s.nodes.FindByID(ctx, id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrNodeNotFound
		}
		return err
	}
	return s.nodes.Delete(ctx, id)
}

func (s *Service) RetriggerHTTPSCheck(ctx context.Context, id uint64) error {
	n, err := s.nodes.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrNodeNotFound
		}
		return err
	}
	go s.runHTTPSCheck(context.Background(), n.ID, n.IP)
	return nil
}

// InstallCommand returns the one-liner install script for a node agent.
func (s *Service) InstallCommand(ctx context.Context, id uint64, managementAddr string) (string, error) {
	n, err := s.nodes.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return "", ErrNodeNotFound
		}
		return "", err
	}
	return fmt.Sprintf(
		`curl -fsSL %s/install-agent.sh | bash -s -- --token %s --server %s`,
		managementAddr, n.AuthToken, managementAddr,
	), nil
}

// --- Agent heartbeat ---

type HeartbeatInput struct {
	Token       string
	CPUPercent  float64
	MemPercent  float64
	HTTPSStatus string
	CertSN      string // TLS client cert serial (from mTLS context)
}

func (s *Service) RecordHeartbeat(ctx context.Context, input HeartbeatInput) error {
	n, err := s.nodes.FindByToken(ctx, input.Token)
	if err != nil {
		return fmt.Errorf("node agent: unknown token")
	}
	now := time.Now()
	a := &entity.NodeAgent{
		NodeID:       n.ID,
		ClientCertSN: input.CertSN,
		LastSeenAt:   &now,
		CPUPercent:   input.CPUPercent,
		MemPercent:   input.MemPercent,
		HTTPSStatus:  input.HTTPSStatus,
	}
	// Try to find existing record so Upsert updates in place.
	existing, err := s.agents.FindByNodeID(ctx, n.ID)
	if err == nil {
		a.ID = existing.ID
	}
	return s.agents.Upsert(ctx, a)
}

// --- Task dispatch ---

type DispatchTaskInput struct {
	NodeID  uint64
	Type    string
	Payload string
}

type TaskView struct {
	ID         uint64                 `json:"id"`
	NodeID     uint64                 `json:"node_id"`
	Type       string                 `json:"type"`
	Status     entity.NodeTaskStatus  `json:"status"`
	ResultMsg  string                 `json:"result_msg,omitempty"`
	FinishedAt *time.Time             `json:"finished_at,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

func (s *Service) DispatchTask(ctx context.Context, input DispatchTaskInput) (*TaskView, error) {
	if !allowedTaskTypes[input.Type] {
		return nil, fmt.Errorf("node: task type %q is not in whitelist", input.Type)
	}
	if _, err := s.nodes.FindByID(ctx, input.NodeID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNodeNotFound
		}
		return nil, err
	}
	t := &entity.NodeTask{
		NodeID:       input.NodeID,
		Type:         input.Type,
		Payload:      input.Payload,
		Status:       entity.NodeTaskPending,
		DispatchedAt: time.Now(),
	}
	if err := s.tasks.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("node: dispatch task: %w", err)
	}
	return toTaskView(t), nil
}

// PullTasks is called by Node Agent to fetch pending tasks.
func (s *Service) PullTasks(ctx context.Context, token string) ([]TaskView, error) {
	n, err := s.nodes.FindByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("node: unknown token")
	}
	tasks, err := s.tasks.ListPendingForNode(ctx, n.ID, 10)
	if err != nil {
		return nil, err
	}
	views := make([]TaskView, len(tasks))
	for i, t := range tasks {
		views[i] = *toTaskView(&t)
	}
	return views, nil
}

// AckTask is called by Node Agent to report task completion.
func (s *Service) AckTask(ctx context.Context, token string, taskID uint64, status entity.NodeTaskStatus, resultMsg string) error {
	n, err := s.nodes.FindByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("node: unknown token")
	}
	t, err := s.tasks.FindByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrTaskNotFound
		}
		return err
	}
	if t.NodeID != n.ID {
		return fmt.Errorf("node: task %d does not belong to this node", taskID)
	}
	// Reject forbidden task type at ack time (defence in depth).
	if !allowedTaskTypes[t.Type] {
		t.Status = entity.NodeTaskForbidden
		t.ResultMsg = "FORBIDDEN_TASK_TYPE: " + t.Type
	} else {
		now := time.Now()
		t.Status = status
		t.ResultMsg = resultMsg
		t.FinishedAt = &now
	}
	return s.tasks.Update(ctx, t)
}

// --- Pool management ---

type CreatePoolInput struct {
	Name    string
	NodeIDs []uint64
}

type PoolView struct {
	ID       uint64              `json:"id"`
	Name     string              `json:"name"`
	IsActive bool                `json:"is_active"`
	Members  []PoolMemberView    `json:"members,omitempty"`
}

type PoolMemberView struct {
	NodeID   uint64 `json:"node_id"`
	Priority int    `json:"priority"`
}

func (s *Service) CreatePool(ctx context.Context, input CreatePoolInput) (*PoolView, error) {
	p := &entity.NodePool{Name: input.Name, IsActive: true}
	if err := s.pools.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("node: create pool: %w", err)
	}
	for _, nid := range input.NodeIDs {
		if err := s.pools.AddMember(ctx, &entity.NodePoolMember{PoolID: p.ID, NodeID: nid}); err != nil {
			return nil, fmt.Errorf("node: add pool member %d: %w", nid, err)
		}
	}
	members, _ := s.pools.ListMembers(ctx, p.ID)
	return toPoolView(p, members), nil
}

func (s *Service) GetPool(ctx context.Context, id uint64) (*PoolView, error) {
	p, err := s.pools.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrPoolNotFound
		}
		return nil, err
	}
	members, _ := s.pools.ListMembers(ctx, p.ID)
	return toPoolView(p, members), nil
}

func (s *Service) ListPools(ctx context.Context, afterID uint64, limit int) ([]PoolView, error) {
	ps, err := s.pools.List(ctx, afterID, limit)
	if err != nil {
		return nil, err
	}
	views := make([]PoolView, len(ps))
	for i, p := range ps {
		views[i] = *toPoolView(&p, nil)
	}
	return views, nil
}

func (s *Service) AddPoolMember(ctx context.Context, poolID, nodeID uint64) error {
	if _, err := s.pools.FindByID(ctx, poolID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrPoolNotFound
		}
		return err
	}
	return s.pools.AddMember(ctx, &entity.NodePoolMember{PoolID: poolID, NodeID: nodeID})
}

func (s *Service) RemovePoolMember(ctx context.Context, poolID, nodeID uint64) error {
	return s.pools.RemoveMember(ctx, poolID, nodeID)
}

func (s *Service) DeletePool(ctx context.Context, id uint64) error {
	if _, err := s.pools.FindByID(ctx, id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrPoolNotFound
		}
		return err
	}
	return s.pools.Delete(ctx, id)
}

// --- internal helpers ---

func (s *Service) runHTTPSCheck(ctx context.Context, nodeID uint64, ip string) {
	n, err := s.nodes.FindByID(ctx, nodeID)
	if err != nil {
		return
	}
	res, err := dnscheck.CheckHTTPS(ctx, ip, ip)
	now := time.Now()
	n.HTTPSCheckedAt = &now
	if err != nil || res == nil {
		n.HTTPSStatus = "failed"
		n.HTTPSCode = 0
	} else {
		n.HTTPSStatus = "ok"
		n.HTTPSCode = res.StatusCode
		n.HTTPSLatencyMs = res.Latency.Milliseconds()
	}
	_ = s.nodes.Update(ctx, n)
}

func generateToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func toNodeView(n *entity.CDNNode) *NodeView {
	return &NodeView{
		ID:             n.ID,
		Name:           n.Name,
		IP:             n.IP,
		ISP:            n.ISP,
		Region:         n.Region,
		OriginAddr:     n.OriginAddr,
		OriginPort:     n.OriginPort,
		IsActive:       n.IsActive,
		AuthConfirmed:  n.AuthConfirmed,
		HTTPSStatus:    n.HTTPSStatus,
		HTTPSCode:      n.HTTPSCode,
		TLSDomain:      n.TLSDomain,
		HTTPSLatencyMs: n.HTTPSLatencyMs,
		HTTPSCheckedAt: n.HTTPSCheckedAt,
		CreatedAt:      n.CreatedAt,
	}
}

func toTaskView(t *entity.NodeTask) *TaskView {
	return &TaskView{
		ID:         t.ID,
		NodeID:     t.NodeID,
		Type:       t.Type,
		Status:     t.Status,
		ResultMsg:  t.ResultMsg,
		FinishedAt: t.FinishedAt,
		CreatedAt:  t.CreatedAt,
	}
}

func toPoolView(p *entity.NodePool, members []entity.NodePoolMember) *PoolView {
	v := &PoolView{ID: p.ID, Name: p.Name, IsActive: p.IsActive}
	for _, m := range members {
		v.Members = append(v.Members, PoolMemberView{NodeID: m.NodeID, Priority: m.Priority})
	}
	return v
}
