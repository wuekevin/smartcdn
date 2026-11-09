package repo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

type DomainRepo struct{ db *gorm.DB }

func NewDomainRepo(db *gorm.DB) *DomainRepo { return &DomainRepo{db: db} }

func (r *DomainRepo) Create(ctx context.Context, d *entity.Domain) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *DomainRepo) FindByID(ctx context.Context, id uint64) (*entity.Domain, error) {
	var d entity.Domain
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&d).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("DomainRepo.FindByID: %w", err)
	}
	return &d, nil
}

func (r *DomainRepo) FindByName(ctx context.Context, name string) (*entity.Domain, error) {
	var d entity.Domain
	if err := r.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", name).First(&d).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("DomainRepo.FindByName: %w", err)
	}
	return &d, nil
}

func (r *DomainRepo) List(ctx context.Context, afterID uint64, limit int) ([]entity.Domain, error) {
	if limit <= 0 {
		limit = 20
	}
	var ds []entity.Domain
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	return ds, q.Order("id ASC").Limit(limit).Find(&ds).Error
}

func (r *DomainRepo) Update(ctx context.Context, d *entity.Domain) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *DomainRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Domain{}).Error
}

// --- DNS Record Repo ---

type DNSRecordRepo struct{ db *gorm.DB }

func NewDNSRecordRepo(db *gorm.DB) *DNSRecordRepo { return &DNSRecordRepo{db: db} }

func (r *DNSRecordRepo) Create(ctx context.Context, rec *entity.DNSRecord) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *DNSRecordRepo) FindByID(ctx context.Context, id uint64) (*entity.DNSRecord, error) {
	var rec entity.DNSRecord
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("DNSRecordRepo.FindByID: %w", err)
	}
	return &rec, nil
}

func (r *DNSRecordRepo) ListByDomain(ctx context.Context, domainID uint64, afterID uint64, limit int) ([]entity.DNSRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	var records []entity.DNSRecord
	q := r.db.WithContext(ctx).Where("domain_id = ? AND deleted_at IS NULL", domainID)
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	return records, q.Order("id ASC").Limit(limit).Find(&records).Error
}

func (r *DNSRecordRepo) Update(ctx context.Context, rec *entity.DNSRecord) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

func (r *DNSRecordRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.DNSRecord{}).Error
}

func (r *DNSRecordRepo) DeleteByDomain(ctx context.Context, domainID uint64) error {
	return r.db.WithContext(ctx).Where("domain_id = ?", domainID).Delete(&entity.DNSRecord{}).Error
}

func (r *DNSRecordRepo) FindByProviderID(ctx context.Context, domainID uint64, providerID string) (*entity.DNSRecord, error) {
	var rec entity.DNSRecord
	if err := r.db.WithContext(ctx).Where("domain_id = ? AND provider_id = ? AND deleted_at IS NULL", domainID, providerID).First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("DNSRecordRepo.FindByProviderID: %w", err)
	}
	return &rec, nil
}

// --- DNS Snapshot Repo ---

type DNSSnapshotRepo struct{ db *gorm.DB }

func NewDNSSnapshotRepo(db *gorm.DB) *DNSSnapshotRepo { return &DNSSnapshotRepo{db: db} }

func (r *DNSSnapshotRepo) Create(ctx context.Context, s *entity.DNSSnapshot) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *DNSSnapshotRepo) FindByID(ctx context.Context, id uint64) (*entity.DNSSnapshot, error) {
	var s entity.DNSSnapshot
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("DNSSnapshotRepo.FindByID: %w", err)
	}
	return &s, nil
}

func (r *DNSSnapshotRepo) ListByDomain(ctx context.Context, domainID uint64, afterID uint64, limit int) ([]entity.DNSSnapshot, error) {
	if limit <= 0 {
		limit = 20
	}
	var ss []entity.DNSSnapshot
	q := r.db.WithContext(ctx).Where("domain_id = ?", domainID)
	if afterID > 0 {
		q = q.Where("id < ?", afterID) // snapshots are listed newest-first
	}
	return ss, q.Order("id DESC").Limit(limit).Find(&ss).Error
}
