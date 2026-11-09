package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// CertificateRepo ─────────────────────────────────────────────────────────

type CertificateRepo struct{ db *gorm.DB }

func NewCertificateRepo(db *gorm.DB) *CertificateRepo { return &CertificateRepo{db: db} }

func (r *CertificateRepo) Create(ctx context.Context, c *entity.Certificate) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CertificateRepo) FindByID(ctx context.Context, id uint64) (*entity.Certificate, error) {
	var c entity.Certificate
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("CertificateRepo.FindByID: %w", err)
	}
	return &c, nil
}

func (r *CertificateRepo) FindByDomainID(ctx context.Context, domainID uint64) ([]entity.Certificate, error) {
	var cs []entity.Certificate
	return cs, r.db.WithContext(ctx).
		Where("domain_id = ? AND deleted_at IS NULL", domainID).
		Order("id DESC").Find(&cs).Error
}

func (r *CertificateRepo) ListDueForRenewal(ctx context.Context, daysThreshold int, limit int) ([]entity.Certificate, error) {
	if limit <= 0 {
		limit = 50
	}
	cutoff := time.Now().AddDate(0, 0, daysThreshold)
	var cs []entity.Certificate
	return cs, r.db.WithContext(ctx).
		Where("status = 'active' AND auto_renew = 1 AND not_after <= ? AND deleted_at IS NULL", cutoff).
		Limit(limit).Find(&cs).Error
}

func (r *CertificateRepo) Update(ctx context.Context, c *entity.Certificate) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *CertificateRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Certificate{}).Error
}

// CertDeployTargetRepo ─────────────────────────────────────────────────────

type CertDeployTargetRepo struct{ db *gorm.DB }

func NewCertDeployTargetRepo(db *gorm.DB) *CertDeployTargetRepo {
	return &CertDeployTargetRepo{db: db}
}

func (r *CertDeployTargetRepo) Create(ctx context.Context, t *entity.CertDeployTarget) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *CertDeployTargetRepo) FindByID(ctx context.Context, id uint64) (*entity.CertDeployTarget, error) {
	var t entity.CertDeployTarget
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("CertDeployTargetRepo.FindByID: %w", err)
	}
	return &t, nil
}

func (r *CertDeployTargetRepo) ListByCert(ctx context.Context, certID uint64) ([]entity.CertDeployTarget, error) {
	var ts []entity.CertDeployTarget
	return ts, r.db.WithContext(ctx).Where("certificate_id = ?", certID).Find(&ts).Error
}

func (r *CertDeployTargetRepo) Update(ctx context.Context, t *entity.CertDeployTarget) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *CertDeployTargetRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.CertDeployTarget{}).Error
}

// ACMEAccountRepo ─────────────────────────────────────────────────────────

type ACMEAccountRepo struct{ db *gorm.DB }

func NewACMEAccountRepo(db *gorm.DB) *ACMEAccountRepo { return &ACMEAccountRepo{db: db} }

func (r *ACMEAccountRepo) Create(ctx context.Context, a *entity.ACMEAccount) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *ACMEAccountRepo) FindByEmail(ctx context.Context, email string) (*entity.ACMEAccount, error) {
	var a entity.ACMEAccount
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("ACMEAccountRepo.FindByEmail: %w", err)
	}
	return &a, nil
}

func (r *ACMEAccountRepo) FindByID(ctx context.Context, id uint64) (*entity.ACMEAccount, error) {
	var a entity.ACMEAccount
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("ACMEAccountRepo.FindByID: %w", err)
	}
	return &a, nil
}

func (r *ACMEAccountRepo) Update(ctx context.Context, a *entity.ACMEAccount) error {
	return r.db.WithContext(ctx).Save(a).Error
}
