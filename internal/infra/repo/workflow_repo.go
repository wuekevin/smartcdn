package repo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// WorkflowTemplateRepo ─────────────────────────────────────────────────────

type WorkflowTemplateRepo struct{ db *gorm.DB }

func NewWorkflowTemplateRepo(db *gorm.DB) *WorkflowTemplateRepo {
	return &WorkflowTemplateRepo{db: db}
}

func (r *WorkflowTemplateRepo) Create(ctx context.Context, t *entity.WorkflowTemplate) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *WorkflowTemplateRepo) FindByID(ctx context.Context, id uint64) (*entity.WorkflowTemplate, error) {
	var t entity.WorkflowTemplate
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("WorkflowTemplateRepo.FindByID: %w", err)
	}
	return &t, nil
}

func (r *WorkflowTemplateRepo) FindByName(ctx context.Context, name string) (*entity.WorkflowTemplate, error) {
	var t entity.WorkflowTemplate
	if err := r.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", name).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("WorkflowTemplateRepo.FindByName: %w", err)
	}
	return &t, nil
}

func (r *WorkflowTemplateRepo) List(ctx context.Context, afterID uint64, limit int) ([]entity.WorkflowTemplate, error) {
	if limit <= 0 {
		limit = 20
	}
	var ts []entity.WorkflowTemplate
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	return ts, q.Order("id ASC").Limit(limit).Find(&ts).Error
}

func (r *WorkflowTemplateRepo) Update(ctx context.Context, t *entity.WorkflowTemplate) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *WorkflowTemplateRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.WorkflowTemplate{}).Error
}

// WorkflowRunRepo ──────────────────────────────────────────────────────────

type WorkflowRunRepo struct{ db *gorm.DB }

func NewWorkflowRunRepo(db *gorm.DB) *WorkflowRunRepo { return &WorkflowRunRepo{db: db} }

func (r *WorkflowRunRepo) Create(ctx context.Context, run *entity.WorkflowRun) error {
	return r.db.WithContext(ctx).Create(run).Error
}

func (r *WorkflowRunRepo) FindByID(ctx context.Context, id uint64) (*entity.WorkflowRun, error) {
	var run entity.WorkflowRun
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("WorkflowRunRepo.FindByID: %w", err)
	}
	return &run, nil
}

func (r *WorkflowRunRepo) ListByTemplate(ctx context.Context, templateID uint64, afterID uint64, limit int) ([]entity.WorkflowRun, error) {
	if limit <= 0 {
		limit = 20
	}
	var runs []entity.WorkflowRun
	q := r.db.WithContext(ctx).Where("template_id = ?", templateID)
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	return runs, q.Order("id DESC").Limit(limit).Find(&runs).Error
}

func (r *WorkflowRunRepo) ListPending(ctx context.Context, limit int) ([]entity.WorkflowRun, error) {
	if limit <= 0 {
		limit = 20
	}
	var runs []entity.WorkflowRun
	return runs, r.db.WithContext(ctx).
		Where("status IN ('pending','running')").
		Order("id ASC").Limit(limit).Find(&runs).Error
}

func (r *WorkflowRunRepo) Update(ctx context.Context, run *entity.WorkflowRun) error {
	return r.db.WithContext(ctx).Save(run).Error
}

// WorkflowNodeRunRepo ──────────────────────────────────────────────────────

type WorkflowNodeRunRepo struct{ db *gorm.DB }

func NewWorkflowNodeRunRepo(db *gorm.DB) *WorkflowNodeRunRepo {
	return &WorkflowNodeRunRepo{db: db}
}

func (r *WorkflowNodeRunRepo) Create(ctx context.Context, n *entity.WorkflowNodeRun) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *WorkflowNodeRunRepo) FindByID(ctx context.Context, id uint64) (*entity.WorkflowNodeRun, error) {
	var n entity.WorkflowNodeRun
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&n).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("WorkflowNodeRunRepo.FindByID: %w", err)
	}
	return &n, nil
}

func (r *WorkflowNodeRunRepo) FindByRunAndNode(ctx context.Context, runID uint64, nodeID string) (*entity.WorkflowNodeRun, error) {
	var n entity.WorkflowNodeRun
	if err := r.db.WithContext(ctx).Where("run_id = ? AND node_id = ?", runID, nodeID).First(&n).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("WorkflowNodeRunRepo.FindByRunAndNode: %w", err)
	}
	return &n, nil
}

func (r *WorkflowNodeRunRepo) ListByRun(ctx context.Context, runID uint64) ([]entity.WorkflowNodeRun, error) {
	var ns []entity.WorkflowNodeRun
	return ns, r.db.WithContext(ctx).Where("run_id = ?", runID).Order("id ASC").Find(&ns).Error
}

func (r *WorkflowNodeRunRepo) Update(ctx context.Context, n *entity.WorkflowNodeRun) error {
	return r.db.WithContext(ctx).Save(n).Error
}
