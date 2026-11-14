package repo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

type PublishJobRepo struct{ db *gorm.DB }

func NewPublishJobRepo(db *gorm.DB) *PublishJobRepo { return &PublishJobRepo{db: db} }

func (r *PublishJobRepo) Create(ctx context.Context, job *entity.PublishJob) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *PublishJobRepo) FindByID(ctx context.Context, id uint64) (*entity.PublishJob, error) {
	var job entity.PublishJob
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("PublishJobRepo.FindByID: %w", err)
	}
	return &job, nil
}

func (r *PublishJobRepo) Update(ctx context.Context, job *entity.PublishJob) error {
	return r.db.WithContext(ctx).Save(job).Error
}

func (r *PublishJobRepo) List(ctx context.Context, afterID uint64, limit int) ([]entity.PublishJob, error) {
	if limit <= 0 {
		limit = 20
	}
	var jobs []entity.PublishJob
	q := r.db.WithContext(ctx)
	if afterID > 0 {
		q = q.Where("id < ?", afterID)
	}
	return jobs, q.Order("id DESC").Limit(limit).Find(&jobs).Error
}

func (r *PublishJobRepo) ListByDomain(ctx context.Context, domainID uint64, afterID uint64, limit int) ([]entity.PublishJob, error) {
	if limit <= 0 {
		limit = 20
	}
	var jobs []entity.PublishJob
	q := r.db.WithContext(ctx).Where("domain_id = ?", domainID)
	if afterID > 0 {
		q = q.Where("id < ?", afterID)
	}
	return jobs, q.Order("id DESC").Limit(limit).Find(&jobs).Error
}
