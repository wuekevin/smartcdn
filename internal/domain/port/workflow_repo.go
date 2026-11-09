package port

import (
	"context"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// WorkflowTemplateRepository abstracts persistence for DAG templates.
type WorkflowTemplateRepository interface {
	Create(ctx context.Context, t *entity.WorkflowTemplate) error
	FindByID(ctx context.Context, id uint64) (*entity.WorkflowTemplate, error)
	FindByName(ctx context.Context, name string) (*entity.WorkflowTemplate, error)
	List(ctx context.Context, afterID uint64, limit int) ([]entity.WorkflowTemplate, error)
	Update(ctx context.Context, t *entity.WorkflowTemplate) error
	Delete(ctx context.Context, id uint64) error
}

// WorkflowRunRepository abstracts persistence for workflow executions.
type WorkflowRunRepository interface {
	Create(ctx context.Context, r *entity.WorkflowRun) error
	FindByID(ctx context.Context, id uint64) (*entity.WorkflowRun, error)
	ListByTemplate(ctx context.Context, templateID uint64, afterID uint64, limit int) ([]entity.WorkflowRun, error)
	ListPending(ctx context.Context, limit int) ([]entity.WorkflowRun, error)
	Update(ctx context.Context, r *entity.WorkflowRun) error
}

// WorkflowNodeRunRepository abstracts persistence for node executions.
type WorkflowNodeRunRepository interface {
	Create(ctx context.Context, n *entity.WorkflowNodeRun) error
	FindByID(ctx context.Context, id uint64) (*entity.WorkflowNodeRun, error)
	FindByRunAndNode(ctx context.Context, runID uint64, nodeID string) (*entity.WorkflowNodeRun, error)
	ListByRun(ctx context.Context, runID uint64) ([]entity.WorkflowNodeRun, error)
	Update(ctx context.Context, n *entity.WorkflowNodeRun) error
}
