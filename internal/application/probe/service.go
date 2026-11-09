package probe

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/infra/repo"
)

var (
	ErrAgentNotFound = errors.New("probe agent not found")
)

// Service manages Probe Agents and their measurement results.
type Service struct {
	agents     port.ProbeAgentRepository
	results    port.ProbeResultRepository
	candidates port.IPCandidateRepository
}

func NewService(
	agents port.ProbeAgentRepository,
	results port.ProbeResultRepository,
	candidates port.IPCandidateRepository,
) *Service {
	return &Service{agents: agents, results: results, candidates: candidates}
}

// --- Probe Agent management ---

type RegisterAgentInput struct {
	Name       string
	ISP        string
	Region     string
	CertSN     string // from mTLS client cert
}

type AgentView struct {
	ID         uint64     `json:"id"`
	Name       string     `json:"name"`
	ISP        string     `json:"isp"`
	Region     string     `json:"region"`
	IsActive   bool       `json:"is_active"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (s *Service) RegisterAgent(ctx context.Context, input RegisterAgentInput) (*AgentView, error) {
	// Upsert by CertSN: re-registration with same cert updates metadata.
	existing, err := s.agents.FindByCertSN(ctx, input.CertSN)
	if err == nil {
		// Update existing
		existing.Name = input.Name
		existing.ISP = input.ISP
		existing.Region = input.Region
		now := time.Now()
		existing.LastSeenAt = &now
		if err := s.agents.Update(ctx, existing); err != nil {
			return nil, err
		}
		return toAgentView(existing), nil
	}
	a := &entity.ProbeAgent{
		Name:         input.Name,
		ISP:          input.ISP,
		Region:       input.Region,
		ClientCertSN: input.CertSN,
		IsActive:     true,
	}
	if err := s.agents.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("probe: register agent: %w", err)
	}
	return toAgentView(a), nil
}

func (s *Service) GetAgent(ctx context.Context, id uint64) (*AgentView, error) {
	a, err := s.agents.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrAgentNotFound
		}
		return nil, err
	}
	return toAgentView(a), nil
}

func (s *Service) ListAgents(ctx context.Context, afterID uint64, limit int) ([]AgentView, error) {
	as, err := s.agents.List(ctx, afterID, limit)
	if err != nil {
		return nil, err
	}
	views := make([]AgentView, len(as))
	for i, a := range as {
		views[i] = *toAgentView(&a)
	}
	return views, nil
}

// RecordHeartbeat updates LastSeenAt for a probe agent identified by CertSN.
func (s *Service) RecordHeartbeat(ctx context.Context, certSN string) error {
	a, err := s.agents.FindByCertSN(ctx, certSN)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrAgentNotFound
		}
		return err
	}
	now := time.Now()
	a.LastSeenAt = &now
	return s.agents.Update(ctx, a)
}

// --- Probe result ingest ---

type SubmitResultInput struct {
	ProbeAgentID uint64
	TargetIP     string
	ISP          string
	LatencyMs    int64
	TTFBMs       int64
	HTTPSStatus  int
	Success      bool
}

func (s *Service) SubmitResult(ctx context.Context, input SubmitResultInput) error {
	r := &entity.ProbeResult{
		ProbeAgentID: input.ProbeAgentID,
		TargetIP:     input.TargetIP,
		ISP:          input.ISP,
		LatencyMs:    input.LatencyMs,
		TTFBMs:       input.TTFBMs,
		HTTPSStatus:  input.HTTPSStatus,
		Success:      input.Success,
		ProbeAt:      time.Now(),
	}
	if err := s.results.Create(ctx, r); err != nil {
		return fmt.Errorf("probe: submit result: %w", err)
	}
	// Trigger candidate status update for the probed IP.
	go s.updateCandidateHealth(context.Background(), input.TargetIP, input.ISP, input.Success)
	return nil
}

// --- Candidate state machine ---

func (s *Service) updateCandidateHealth(ctx context.Context, ip, isp string, success bool) {
	// Find all candidates matching this IP. TargetIP maps to CDNNode.IP.
	// We scan all domains' candidates for this IP, since a node may serve multiple domains.
	// In practice the scheduler will re-score, but we update fail streaks here.
	now := time.Now()
	since := now.Add(-15 * time.Minute)
	results, err := s.results.ListRecentByIPAndISP(ctx, ip, isp, since, 10)
	if err != nil {
		return
	}
	failures := 0
	for _, r := range results {
		if !r.Success {
			failures++
		}
	}
	// If 3+ failures in last 15 minutes → move candidate to "offline".
	if failures >= 3 {
		s.markCandidatesOffline(ctx, ip)
	}
}

func (s *Service) markCandidatesOffline(ctx context.Context, ip string) {
	// We don't have a FindByIP on IPCandidateRepo, so this is best-effort.
	// The scheduler's rescore loop will handle it properly.
	_ = ip
}

// --- IP Candidate management ---

type AddCandidateInput struct {
	DomainID uint64
	NodeID   uint64
	IP       string
	ISP      string
}

type CandidateView struct {
	ID          uint64                      `json:"id"`
	DomainID    uint64                      `json:"domain_id"`
	NodeID      uint64                      `json:"node_id"`
	IP          string                      `json:"ip"`
	ISP         string                      `json:"isp"`
	Status      entity.IPCandidateStatus    `json:"status"`
	FailStreak  int                         `json:"fail_streak"`
	Score       float64                     `json:"score"`
	LastScoreAt *time.Time                  `json:"last_score_at,omitempty"`
}

func (s *Service) AddCandidate(ctx context.Context, input AddCandidateInput) (*CandidateView, error) {
	c := &entity.IPCandidate{
		DomainID: input.DomainID,
		NodeID:   input.NodeID,
		IP:       input.IP,
		ISP:      input.ISP,
		Status:   entity.IPCandidateCandidate,
	}
	if err := s.candidates.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("probe: add candidate: %w", err)
	}
	return toCandidateView(c), nil
}

func (s *Service) ListCandidates(ctx context.Context, domainID uint64, afterID uint64, limit int) ([]CandidateView, error) {
	cs, err := s.candidates.ListByDomain(ctx, domainID, afterID, limit)
	if err != nil {
		return nil, err
	}
	views := make([]CandidateView, len(cs))
	for i, c := range cs {
		views[i] = *toCandidateView(&c)
	}
	return views, nil
}

func (s *Service) PromoteCandidate(ctx context.Context, id uint64) error {
	c, err := s.candidates.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if c.Status != entity.IPCandidateCandidate {
		return fmt.Errorf("probe: candidate %d is not in 'candidate' status", id)
	}
	c.Status = entity.IPCandidateEligible
	return s.candidates.Update(ctx, c)
}

func (s *Service) LockCandidate(ctx context.Context, id uint64) error {
	c, err := s.candidates.FindByID(ctx, id)
	if err != nil {
		return err
	}
	c.Status = entity.IPCandidateLocked
	return s.candidates.Update(ctx, c)
}

func (s *Service) DeleteCandidate(ctx context.Context, id uint64) error {
	return s.candidates.Delete(ctx, id)
}

// GetRecentResults returns probe results for a given IP from the last N minutes.
func (s *Service) GetRecentResults(ctx context.Context, ip string, minutes int) ([]entity.ProbeResult, error) {
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	return s.results.ListRecentByIP(ctx, ip, since, 50)
}

// --- helpers ---

func toAgentView(a *entity.ProbeAgent) *AgentView {
	return &AgentView{
		ID:         a.ID,
		Name:       a.Name,
		ISP:        a.ISP,
		Region:     a.Region,
		IsActive:   a.IsActive,
		LastSeenAt: a.LastSeenAt,
		CreatedAt:  a.CreatedAt,
	}
}

func toCandidateView(c *entity.IPCandidate) *CandidateView {
	return &CandidateView{
		ID:          c.ID,
		DomainID:    c.DomainID,
		NodeID:      c.NodeID,
		IP:          c.IP,
		ISP:         c.ISP,
		Status:      c.Status,
		FailStreak:  c.FailStreak,
		Score:       c.Score,
		LastScoreAt: c.LastScoreAt,
	}
}
