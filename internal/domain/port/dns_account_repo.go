package port

import (
	"context"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// DNSAccountRepository abstracts persistence for DNS provider accounts.
type DNSAccountRepository interface {
	Create(ctx context.Context, acc *entity.DNSAccount) error
	FindByID(ctx context.Context, id uint64) (*entity.DNSAccount, error)
	List(ctx context.Context, afterID uint64, limit int) ([]entity.DNSAccount, error)
	UpdateCredentials(ctx context.Context, id uint64, credentialsEnc string) error
	UpdateCapabilities(ctx context.Context, id uint64, capabilitiesJSON string) error
	Delete(ctx context.Context, id uint64) error
}
