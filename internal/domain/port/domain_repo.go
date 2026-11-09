package port

import (
	"context"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// DomainRepository abstracts persistence for managed DNS zones.
type DomainRepository interface {
	Create(ctx context.Context, d *entity.Domain) error
	FindByID(ctx context.Context, id uint64) (*entity.Domain, error)
	FindByName(ctx context.Context, name string) (*entity.Domain, error)
	List(ctx context.Context, afterID uint64, limit int) ([]entity.Domain, error)
	Update(ctx context.Context, d *entity.Domain) error
	Delete(ctx context.Context, id uint64) error
}

// DNSRecordRepository abstracts persistence for DNS records.
type DNSRecordRepository interface {
	Create(ctx context.Context, r *entity.DNSRecord) error
	FindByID(ctx context.Context, id uint64) (*entity.DNSRecord, error)
	ListByDomain(ctx context.Context, domainID uint64, afterID uint64, limit int) ([]entity.DNSRecord, error)
	Update(ctx context.Context, r *entity.DNSRecord) error
	Delete(ctx context.Context, id uint64) error
	DeleteByDomain(ctx context.Context, domainID uint64) error
	FindByProviderID(ctx context.Context, domainID uint64, providerID string) (*entity.DNSRecord, error)
}

// DNSSnapshotRepository abstracts persistence for DNS snapshots.
type DNSSnapshotRepository interface {
	Create(ctx context.Context, s *entity.DNSSnapshot) error
	FindByID(ctx context.Context, id uint64) (*entity.DNSSnapshot, error)
	ListByDomain(ctx context.Context, domainID uint64, afterID uint64, limit int) ([]entity.DNSSnapshot, error)
}
