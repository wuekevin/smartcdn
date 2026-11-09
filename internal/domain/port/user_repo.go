package port

import (
	"context"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// UserRepository abstracts persistence for users.
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	FindByID(ctx context.Context, id uint64) (*entity.User, error)
	UpdateLoginFailed(ctx context.Context, userID uint64, count int, lockedUntil interface{}) error
	UpdateLastLogin(ctx context.Context, userID uint64) error
	GetPermissions(ctx context.Context, userID uint64) ([]entity.Permission, error)
}
