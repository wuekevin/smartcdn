package repo

import (
	"context"

	"gorm.io/gorm"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

type AuditRepo struct {
	db *gorm.DB
}

func NewAuditRepo(db *gorm.DB) *AuditRepo {
	return &AuditRepo{db: db}
}

// Create inserts an audit log entry. No delete method is exposed (AD-15).
func (r *AuditRepo) Create(ctx context.Context, log *entity.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// List returns audit log entries with cursor-based pagination (AD-7).
func (r *AuditRepo) List(ctx context.Context, filter AuditFilter) ([]entity.AuditLog, error) {
	q := r.db.WithContext(ctx).Model(&entity.AuditLog{}).Order("id DESC")

	if filter.OperatorID != 0 {
		q = q.Where("operator_id = ?", filter.OperatorID)
	}
	if !filter.StartTime.IsZero() {
		q = q.Where("created_at >= ?", filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		q = q.Where("created_at <= ?", filter.EndTime)
	}
	if filter.AfterID != 0 {
		q = q.Where("id < ?", filter.AfterID)
	}

	limit := filter.Limit
	if limit == 0 || limit > 100 {
		limit = 20
	}
	q = q.Limit(limit)

	var logs []entity.AuditLog
	return logs, q.Find(&logs).Error
}
