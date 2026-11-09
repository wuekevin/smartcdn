package repo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// CDNAccountRepo ───────────────────────────────────────────────────────────

type CDNAccountRepo struct{ db *gorm.DB }

func NewCDNAccountRepo(db *gorm.DB) *CDNAccountRepo { return &CDNAccountRepo{db: db} }

func (r *CDNAccountRepo) Create(ctx context.Context, acc *entity.CDNAccount) error {
	return r.db.WithContext(ctx).Create(acc).Error
}

func (r *CDNAccountRepo) FindByID(ctx context.Context, id uint64) (*entity.CDNAccount, error) {
	var acc entity.CDNAccount
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&acc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("CDNAccountRepo.FindByID: %w", err)
	}
	return &acc, nil
}

func (r *CDNAccountRepo) List(ctx context.Context, afterID uint64, limit int) ([]entity.CDNAccount, error) {
	if limit <= 0 {
		limit = 20
	}
	var accs []entity.CDNAccount
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	return accs, q.Order("id ASC").Limit(limit).Find(&accs).Error
}

func (r *CDNAccountRepo) Update(ctx context.Context, acc *entity.CDNAccount) error {
	return r.db.WithContext(ctx).Save(acc).Error
}

func (r *CDNAccountRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.CDNAccount{}).Error
}

// CDNDomainRepo ────────────────────────────────────────────────────────────

type CDNDomainRepo struct{ db *gorm.DB }

func NewCDNDomainRepo(db *gorm.DB) *CDNDomainRepo { return &CDNDomainRepo{db: db} }

func (r *CDNDomainRepo) Create(ctx context.Context, d *entity.CDNDomain) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *CDNDomainRepo) FindByID(ctx context.Context, id uint64) (*entity.CDNDomain, error) {
	var d entity.CDNDomain
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&d).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("CDNDomainRepo.FindByID: %w", err)
	}
	return &d, nil
}

func (r *CDNDomainRepo) FindByDomainName(ctx context.Context, name string) (*entity.CDNDomain, error) {
	var d entity.CDNDomain
	if err := r.db.WithContext(ctx).Where("domain_name = ? AND deleted_at IS NULL", name).First(&d).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("CDNDomainRepo.FindByDomainName: %w", err)
	}
	return &d, nil
}

func (r *CDNDomainRepo) List(ctx context.Context, afterID uint64, limit int) ([]entity.CDNDomain, error) {
	if limit <= 0 {
		limit = 20
	}
	var ds []entity.CDNDomain
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	return ds, q.Order("id ASC").Limit(limit).Find(&ds).Error
}

func (r *CDNDomainRepo) Update(ctx context.Context, d *entity.CDNDomain) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *CDNDomainRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.CDNDomain{}).Error
}

// CDNCustomHostnameRepo ────────────────────────────────────────────────────

type CDNCustomHostnameRepo struct{ db *gorm.DB }

func NewCDNCustomHostnameRepo(db *gorm.DB) *CDNCustomHostnameRepo {
	return &CDNCustomHostnameRepo{db: db}
}

func (r *CDNCustomHostnameRepo) Create(ctx context.Context, h *entity.CDNCustomHostname) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *CDNCustomHostnameRepo) FindByID(ctx context.Context, id uint64) (*entity.CDNCustomHostname, error) {
	var h entity.CDNCustomHostname
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&h).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("CDNCustomHostnameRepo.FindByID: %w", err)
	}
	return &h, nil
}

func (r *CDNCustomHostnameRepo) ListByCDNDomain(ctx context.Context, cdnDomainID uint64, afterID uint64, limit int) ([]entity.CDNCustomHostname, error) {
	if limit <= 0 {
		limit = 20
	}
	var hs []entity.CDNCustomHostname
	q := r.db.WithContext(ctx).Where("cdn_domain_id = ?", cdnDomainID)
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	return hs, q.Order("id ASC").Limit(limit).Find(&hs).Error
}

func (r *CDNCustomHostnameRepo) Update(ctx context.Context, h *entity.CDNCustomHostname) error {
	return r.db.WithContext(ctx).Save(h).Error
}

func (r *CDNCustomHostnameRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.CDNCustomHostname{}).Error
}
