package port

import (
	"context"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// CDNAccountRepository abstracts persistence for CDN provider accounts.
type CDNAccountRepository interface {
	Create(ctx context.Context, acc *entity.CDNAccount) error
	FindByID(ctx context.Context, id uint64) (*entity.CDNAccount, error)
	List(ctx context.Context, afterID uint64, limit int) ([]entity.CDNAccount, error)
	Update(ctx context.Context, acc *entity.CDNAccount) error
	Delete(ctx context.Context, id uint64) error
}

// CDNDomainRepository abstracts persistence for CDN domains.
type CDNDomainRepository interface {
	Create(ctx context.Context, d *entity.CDNDomain) error
	FindByID(ctx context.Context, id uint64) (*entity.CDNDomain, error)
	FindByDomainName(ctx context.Context, domainName string) (*entity.CDNDomain, error)
	List(ctx context.Context, afterID uint64, limit int) ([]entity.CDNDomain, error)
	Update(ctx context.Context, d *entity.CDNDomain) error
	Delete(ctx context.Context, id uint64) error
}

// CDNCustomHostnameRepository abstracts persistence for SaaS custom hostnames.
type CDNCustomHostnameRepository interface {
	Create(ctx context.Context, h *entity.CDNCustomHostname) error
	FindByID(ctx context.Context, id uint64) (*entity.CDNCustomHostname, error)
	ListByCDNDomain(ctx context.Context, cdnDomainID uint64, afterID uint64, limit int) ([]entity.CDNCustomHostname, error)
	Update(ctx context.Context, h *entity.CDNCustomHostname) error
	Delete(ctx context.Context, id uint64) error
}
