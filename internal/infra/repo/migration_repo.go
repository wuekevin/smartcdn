package repo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// MigrationJobRepo ─────────────────────────────────────────────────────────

type MigrationJobRepo struct{ db *gorm.DB }

func NewMigrationJobRepo(db *gorm.DB) *MigrationJobRepo { return &MigrationJobRepo{db: db} }

func (r *MigrationJobRepo) Create(ctx context.Context, job *entity.MigrationJob) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *MigrationJobRepo) FindByID(ctx context.Context, id uint64) (*entity.MigrationJob, error) {
	var job entity.MigrationJob
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("MigrationJobRepo.FindByID: %w", err)
	}
	return &job, nil
}

func (r *MigrationJobRepo) Update(ctx context.Context, job *entity.MigrationJob) error {
	return r.db.WithContext(ctx).Save(job).Error
}

func (r *MigrationJobRepo) ListByDomain(ctx context.Context, domainID uint64, afterID uint64, limit int) ([]entity.MigrationJob, error) {
	if limit <= 0 {
		limit = 20
	}
	var jobs []entity.MigrationJob
	q := r.db.WithContext(ctx).Where("domain_id = ?", domainID)
	if afterID > 0 {
		q = q.Where("id < ?", afterID)
	}
	return jobs, q.Order("id DESC").Limit(limit).Find(&jobs).Error
}

// MultiDNSConfigRepo ───────────────────────────────────────────────────────

type MultiDNSConfigRepo struct{ db *gorm.DB }

func NewMultiDNSConfigRepo(db *gorm.DB) *MultiDNSConfigRepo {
	return &MultiDNSConfigRepo{db: db}
}

func (r *MultiDNSConfigRepo) Create(ctx context.Context, cfg *entity.MultiDNSConfig) error {
	return r.db.WithContext(ctx).Create(cfg).Error
}

func (r *MultiDNSConfigRepo) FindByDomainID(ctx context.Context, domainID uint64) (*entity.MultiDNSConfig, error) {
	var cfg entity.MultiDNSConfig
	if err := r.db.WithContext(ctx).Where("domain_id = ?", domainID).First(&cfg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("MultiDNSConfigRepo.FindByDomainID: %w", err)
	}
	return &cfg, nil
}

func (r *MultiDNSConfigRepo) Update(ctx context.Context, cfg *entity.MultiDNSConfig) error {
	return r.db.WithContext(ctx).Save(cfg).Error
}

func (r *MultiDNSConfigRepo) Delete(ctx context.Context, domainID uint64) error {
	return r.db.WithContext(ctx).Where("domain_id = ?", domainID).Delete(&entity.MultiDNSConfig{}).Error
}
