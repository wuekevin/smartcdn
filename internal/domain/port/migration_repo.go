package port

import (
	"context"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// MigrationJobRepository persists DNS migration jobs.
type MigrationJobRepository interface {
	Create(ctx context.Context, job *entity.MigrationJob) error
	FindByID(ctx context.Context, id uint64) (*entity.MigrationJob, error)
	Update(ctx context.Context, job *entity.MigrationJob) error
	ListByDomain(ctx context.Context, domainID uint64, afterID uint64, limit int) ([]entity.MigrationJob, error)
}

// MultiDNSConfigRepository persists dual-authority configuration.
type MultiDNSConfigRepository interface {
	Create(ctx context.Context, cfg *entity.MultiDNSConfig) error
	FindByDomainID(ctx context.Context, domainID uint64) (*entity.MultiDNSConfig, error)
	Update(ctx context.Context, cfg *entity.MultiDNSConfig) error
	Delete(ctx context.Context, domainID uint64) error
}
