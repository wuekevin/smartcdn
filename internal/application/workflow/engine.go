package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/infra/repo"
)

var (
	ErrTemplateNotFound = errors.New("workflow template not found")
	ErrRunNotFound      = errors.New("workflow run not found")
	ErrCyclicDAG        = errors.New("workflow graph contains a cycle")
	ErrUnknownNodeType  = errors.New("unknown workflow node type")
)

// WorkflowNodeDef is a node in the DAG template graph.
type WorkflowNodeDef struct {
	ID       string                 `json:"id"`       // unique within template
	Type     entity.WorkflowNodeType `json:"type"`
	Params   map[string]interface{} `json:"params,omitempty"`
	DependsOn []string              `json:"depends_on,omitempty"` // node IDs that must complete before this
}

// NodeExecutor is called by the engine to execute a single node.
type NodeExecutor interface {
	Execute(ctx context.Context, run *entity.WorkflowRun, node WorkflowNodeDef, input map[string]interface{}) (output map[string]interface{}, err error)
}

// Engine is the DAG workflow execution engine.
type Engine struct {
	templates port.WorkflowTemplateRepository
	runs      port.WorkflowRunRepository
	nodeRuns  port.WorkflowNodeRunRepository
	executors map[entity.WorkflowNodeType]NodeExecutor
	mu        sync.RWMutex
}

func NewEngine(
	templates port.WorkflowTemplateRepository,
	runs port.WorkflowRunRepository,
	nodeRuns port.WorkflowNodeRunRepository,
) *Engine {
	return &Engine{
		templates: templates,
		runs:      runs,
		nodeRuns:  nodeRuns,
		executors: make(map[entity.WorkflowNodeType]NodeExecutor),
	}
}

// RegisterExecutor registers a handler for a given node type.
func (e *Engine) RegisterExecutor(nodeType entity.WorkflowNodeType, exec NodeExecutor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.executors[nodeType] = exec
}

// --- Template management ---

type TemplateView struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	GraphJSON   string `json:"graph_json"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

func (e *Engine) CreateTemplate(ctx context.Context, name, description, graphJSON string) (*TemplateView, error) {
	if err := validateGraphJSON(graphJSON); err != nil {
		return nil, fmt.Errorf("workflow: invalid graph: %w", err)
	}
	t := &entity.WorkflowTemplate{
		Name:        name,
		Description: description,
		GraphJSON:   graphJSON,
		IsActive:    true,
	}
	if err := e.templates.Create(ctx, t); err != nil {
		return nil, err
	}
	return toTemplateView(t), nil
}

func (e *Engine) GetTemplate(ctx context.Context, id uint64) (*TemplateView, error) {
	t, err := e.templates.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}
	return toTemplateView(t), nil
}

func (e *Engine) ListTemplates(ctx context.Context, afterID uint64, limit int) ([]TemplateView, error) {
	ts, err := e.templates.List(ctx, afterID, limit)
	if err != nil {
		return nil, err
	}
	views := make([]TemplateView, len(ts))
	for i, t := range ts {
		views[i] = *toTemplateView(&t)
	}
	return views, nil
}

func (e *Engine) UpdateTemplate(ctx context.Context, id uint64, description, graphJSON string) (*TemplateView, error) {
	t, err := e.templates.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}
	if graphJSON != "" {
		if err := validateGraphJSON(graphJSON); err != nil {
			return nil, fmt.Errorf("workflow: invalid graph: %w", err)
		}
		t.GraphJSON = graphJSON
	}
	if description != "" {
		t.Description = description
	}
	if err := e.templates.Update(ctx, t); err != nil {
		return nil, err
	}
	return toTemplateView(t), nil
}

func (e *Engine) DeleteTemplate(ctx context.Context, id uint64) error {
	if _, err := e.templates.FindByID(ctx, id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrTemplateNotFound
		}
		return err
	}
	return e.templates.Delete(ctx, id)
}

// --- Run management ---

type RunView struct {
	ID           uint64                `json:"id"`
	TemplateName string                `json:"template_name"`
	TriggerType  string                `json:"trigger_type"`
	TriggerRef   string                `json:"trigger_ref,omitempty"`
	Status       entity.WorkflowStatus  `json:"status"`
	StartedAt    *time.Time            `json:"started_at,omitempty"`
	FinishedAt   *time.Time            `json:"finished_at,omitempty"`
	LastError    string                `json:"last_error,omitempty"`
	RetryCount   int                   `json:"retry_count"`
	CreatedAt    time.Time             `json:"created_at"`
}

type NodeRunView struct {
	ID         uint64                     `json:"id"`
	NodeID     string                     `json:"node_id"`
	NodeType   entity.WorkflowNodeType    `json:"node_type"`
	Status     entity.WorkflowNodeStatus  `json:"status"`
	StartedAt  *time.Time                 `json:"started_at,omitempty"`
	FinishedAt *time.Time                 `json:"finished_at,omitempty"`
	Error      string                     `json:"error,omitempty"`
	RetryCount int                        `json:"retry_count"`
}

// TriggerRun starts a workflow run for a given template.
func (e *Engine) TriggerRun(ctx context.Context, templateID uint64, triggerType, triggerRef string, ctxVars map[string]interface{}) (*RunView, error) {
	t, err := e.templates.FindByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}
	ctxJSON, _ := json.Marshal(ctxVars)
	run := &entity.WorkflowRun{
		TemplateID:   t.ID,
		TemplateName: t.Name,
		TriggerType:  triggerType,
		TriggerRef:   triggerRef,
		Status:       entity.WFStatusPending,
		ContextJSON:  string(ctxJSON),
	}
	if err := e.runs.Create(ctx, run); err != nil {
		return nil, fmt.Errorf("workflow: create run: %w", err)
	}
	// Execute asynchronously.
	go e.executeRun(context.Background(), run.ID, t.GraphJSON)
	return toRunView(run), nil
}

func (e *Engine) GetRun(ctx context.Context, id uint64) (*RunView, error) {
	r, err := e.runs.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrRunNotFound
		}
		return nil, err
	}
	return toRunView(r), nil
}

func (e *Engine) ListRuns(ctx context.Context, templateID uint64, afterID uint64, limit int) ([]RunView, error) {
	rs, err := e.runs.ListByTemplate(ctx, templateID, afterID, limit)
	if err != nil {
		return nil, err
	}
	views := make([]RunView, len(rs))
	for i, r := range rs {
		views[i] = *toRunView(&r)
	}
	return views, nil
}

func (e *Engine) GetNodeRuns(ctx context.Context, runID uint64) ([]NodeRunView, error) {
	ns, err := e.nodeRuns.ListByRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	views := make([]NodeRunView, len(ns))
	for i, n := range ns {
		views[i] = toNodeRunView(&n)
	}
	return views, nil
}

func (e *Engine) CancelRun(ctx context.Context, id uint64) error {
	run, err := e.runs.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrRunNotFound
		}
		return err
	}
	if run.Status != entity.WFStatusPending && run.Status != entity.WFStatusRunning {
		return fmt.Errorf("workflow: run %d is already %s", id, run.Status)
	}
	run.Status = entity.WFStatusCancelled
	now := time.Now()
	run.FinishedAt = &now
	return e.runs.Update(ctx, run)
}

// --- DAG execution ---

func (e *Engine) executeRun(ctx context.Context, runID uint64, graphJSON string) {
	run, err := e.runs.FindByID(ctx, runID)
	if err != nil {
		return
	}
	run.Status = entity.WFStatusRunning
	now := time.Now()
	run.StartedAt = &now
	_ = e.runs.Update(ctx, run)

	nodes, err := parseGraph(graphJSON)
	if err != nil {
		e.failRun(ctx, run, err.Error())
		return
	}

	// Topological sort to get execution order.
	order, err := topoSort(nodes)
	if err != nil {
		e.failRun(ctx, run, err.Error())
		return
	}

	// Build a map of node results for dependency injection.
	nodeOutputs := map[string]map[string]interface{}{}
	var runCtx map[string]interface{}
	_ = json.Unmarshal([]byte(run.ContextJSON), &runCtx)
	if runCtx == nil {
		runCtx = map[string]interface{}{}
	}

	for _, node := range order {
		// Check if run was cancelled.
		fresh, _ := e.runs.FindByID(ctx, runID)
		if fresh != nil && fresh.Status == entity.WFStatusCancelled {
			return
		}

		// Build input from run context + dependency outputs.
		input := map[string]interface{}{}
		for k, v := range runCtx {
			input[k] = v
		}
		for _, dep := range node.DependsOn {
			if out, ok := nodeOutputs[dep]; ok {
				for k, v := range out {
					input[dep+"."+k] = v
				}
			}
		}

		nr := &entity.WorkflowNodeRun{
			RunID:    runID,
			NodeID:   node.ID,
			NodeType: node.Type,
			Status:   entity.WFNodeRunning,
		}
		inJ, _ := json.Marshal(input)
		nr.InputJSON = string(inJ)
		startTime := time.Now()
		nr.StartedAt = &startTime
		_ = e.nodeRuns.Create(ctx, nr)

		e.mu.RLock()
		exec, ok := e.executors[node.Type]
		e.mu.RUnlock()

		var output map[string]interface{}
		var execErr error

		if !ok {
			execErr = fmt.Errorf("%w: %s", ErrUnknownNodeType, node.Type)
		} else {
			output, execErr = exec.Execute(ctx, run, node, input)
		}

		finTime := time.Now()
		nr.FinishedAt = &finTime
		if execErr != nil {
			nr.Status = entity.WFNodeFailed
			nr.Error = execErr.Error()
			_ = e.nodeRuns.Update(ctx, nr)
			e.failRun(ctx, run, fmt.Sprintf("node %s failed: %s", node.ID, execErr.Error()))
			return
		}
		outJ, _ := json.Marshal(output)
		nr.OutputJSON = string(outJ)
		nr.Status = entity.WFNodeSucceeded
		_ = e.nodeRuns.Update(ctx, nr)
		nodeOutputs[node.ID] = output
	}

	// All nodes succeeded.
	run.Status = entity.WFStatusSucceeded
	done := time.Now()
	run.FinishedAt = &done
	_ = e.runs.Update(ctx, run)
}

func (e *Engine) failRun(ctx context.Context, run *entity.WorkflowRun, reason string) {
	run.Status = entity.WFStatusFailed
	run.LastError = reason
	now := time.Now()
	run.FinishedAt = &now
	_ = e.runs.Update(ctx, run)
}

// --- DAG helpers ---

func parseGraph(graphJSON string) ([]WorkflowNodeDef, error) {
	var nodes []WorkflowNodeDef
	if err := json.Unmarshal([]byte(graphJSON), &nodes); err != nil {
		return nil, fmt.Errorf("parse graph JSON: %w", err)
	}
	return nodes, nil
}

// topoSort returns nodes in topological order (dependencies first).
// Returns ErrCyclicDAG if a cycle is detected.
func topoSort(nodes []WorkflowNodeDef) ([]WorkflowNodeDef, error) {
	idx := map[string]WorkflowNodeDef{}
	for _, n := range nodes {
		idx[n.ID] = n
	}
	visited := map[string]int{} // 0=unvisited, 1=in-progress, 2=done
	var order []WorkflowNodeDef

	var visit func(id string) error
	visit = func(id string) error {
		switch visited[id] {
		case 2:
			return nil
		case 1:
			return fmt.Errorf("%w: detected at node %q", ErrCyclicDAG, id)
		}
		visited[id] = 1
		node, ok := idx[id]
		if !ok {
			return fmt.Errorf("workflow: undefined dependency %q", id)
		}
		for _, dep := range node.DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visited[id] = 2
		order = append(order, node)
		return nil
	}

	for _, n := range nodes {
		if err := visit(n.ID); err != nil {
			return nil, err
		}
	}
	return order, nil
}

func validateGraphJSON(graphJSON string) error {
	nodes, err := parseGraph(graphJSON)
	if err != nil {
		return err
	}
	_, err = topoSort(nodes)
	return err
}

// --- helpers ---

func toTemplateView(t *entity.WorkflowTemplate) *TemplateView {
	return &TemplateView{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		GraphJSON:   t.GraphJSON,
		IsActive:    t.IsActive,
		CreatedAt:   t.CreatedAt,
	}
}

func toRunView(r *entity.WorkflowRun) *RunView {
	return &RunView{
		ID:           r.ID,
		TemplateName: r.TemplateName,
		TriggerType:  r.TriggerType,
		TriggerRef:   r.TriggerRef,
		Status:       r.Status,
		StartedAt:    r.StartedAt,
		FinishedAt:   r.FinishedAt,
		LastError:    r.LastError,
		RetryCount:   r.RetryCount,
		CreatedAt:    r.CreatedAt,
	}
}

func toNodeRunView(n *entity.WorkflowNodeRun) NodeRunView {
	return NodeRunView{
		ID:         n.ID,
		NodeID:     n.NodeID,
		NodeType:   n.NodeType,
		Status:     n.Status,
		StartedAt:  n.StartedAt,
		FinishedAt: n.FinishedAt,
		Error:      n.Error,
		RetryCount: n.RetryCount,
	}
}
