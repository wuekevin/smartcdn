package port

import (
	"context"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// CertificateRepository abstracts persistence for SSL certificates.
type CertificateRepository interface {
	Create(ctx context.Context, c *entity.Certificate) error
	FindByID(ctx context.Context, id uint64) (*entity.Certificate, error)
	FindByDomainID(ctx context.Context, domainID uint64) ([]entity.Certificate, error)
	ListDueForRenewal(ctx context.Context, daysThreshold int, limit int) ([]entity.Certificate, error)
	Update(ctx context.Context, c *entity.Certificate) error
	Delete(ctx context.Context, id uint64) error
}

// CertDeployTargetRepository abstracts deploy target persistence.
type CertDeployTargetRepository interface {
	Create(ctx context.Context, t *entity.CertDeployTarget) error
	FindByID(ctx context.Context, id uint64) (*entity.CertDeployTarget, error)
	ListByCert(ctx context.Context, certID uint64) ([]entity.CertDeployTarget, error)
	Update(ctx context.Context, t *entity.CertDeployTarget) error
	Delete(ctx context.Context, id uint64) error
}

// ACMEAccountRepository abstracts ACME account persistence.
type ACMEAccountRepository interface {
	Create(ctx context.Context, a *entity.ACMEAccount) error
	FindByEmail(ctx context.Context, email string) (*entity.ACMEAccount, error)
	FindByID(ctx context.Context, id uint64) (*entity.ACMEAccount, error)
	Update(ctx context.Context, a *entity.ACMEAccount) error
}
