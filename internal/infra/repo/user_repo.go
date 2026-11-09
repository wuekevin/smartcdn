package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

var ErrNotFound = errors.New("record not found")

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	var u entity.User
	if err := r.db.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", username).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("FindByUsername: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uint64) (*entity.User, error) {
	var u entity.User
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("FindByID: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) UpdateLoginFailed(ctx context.Context, userID uint64, count int, lockedUntil interface{}) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"login_failed_count": count,
			"locked_until":       lockedUntil,
		}).Error
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, userID uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at":      now,
			"login_failed_count": 0,
			"locked_until":       nil,
		}).Error
}

func (r *UserRepo) GetPermissions(ctx context.Context, userID uint64) ([]entity.Permission, error) {
	var perms []entity.Permission
	err := r.db.WithContext(ctx).Raw(`
		SELECT rp.role_id, rp.resource, rp.action
		FROM role_permissions rp
		INNER JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = ?
	`, userID).Scan(&perms).Error
	return perms, err
}
