package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/infra/logger"
)

const (
	permCacheTTL    = 5 * time.Minute
	permInvalidChan = "perm_invalidate"
)

func permCacheKey(userID uint64) string {
	return fmt.Sprintf("perm:%d", userID)
}

// PermissionService checks and caches user permissions (AD-9).
type PermissionService struct {
	users port.UserRepository
	rdb   *redis.Client
}

func NewPermissionService(users port.UserRepository, rdb *redis.Client) *PermissionService {
	ps := &PermissionService{users: users, rdb: rdb}
	go ps.subscribeInvalidations()
	return ps
}

// HasPermission returns true if the user holds resource:action.
func (ps *PermissionService) HasPermission(ctx context.Context, userID uint64, resource, action string) (bool, error) {
	perms, err := ps.loadPermissions(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, p := range perms {
		if p.Resource == resource && p.Action == action {
			return true, nil
		}
	}
	return false, nil
}

// InvalidateUser removes the cached permissions for a user and broadcasts the
// invalidation so other instances also drop their caches.
func (ps *PermissionService) InvalidateUser(ctx context.Context, userID uint64) {
	key := permCacheKey(userID)
	ps.rdb.Del(ctx, key)
	ps.rdb.Publish(ctx, permInvalidChan, fmt.Sprintf("%d", userID))
}

// loadPermissions fetches permissions from Redis cache or DB.
func (ps *PermissionService) loadPermissions(ctx context.Context, userID uint64) ([]entity.Permission, error) {
	key := permCacheKey(userID)

	cached, err := ps.rdb.Get(ctx, key).Bytes()
	if err == nil {
		var perms []entity.Permission
		if jsonErr := json.Unmarshal(cached, &perms); jsonErr == nil {
			return perms, nil
		}
	}

	perms, err := ps.users.GetPermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("loading permissions: %w", err)
	}

	if b, err := json.Marshal(perms); err == nil {
		ps.rdb.Set(ctx, key, b, permCacheTTL)
	}

	return perms, nil
}

// subscribeInvalidations listens on the Pub/Sub channel and drops local cache entries.
func (ps *PermissionService) subscribeInvalidations() {
	ctx := context.Background()
	sub := ps.rdb.Subscribe(ctx, permInvalidChan)
	defer sub.Close()

	log := logger.Named("perm")
	for msg := range sub.Channel() {
		var userID uint64
		if _, err := fmt.Sscanf(msg.Payload, "%d", &userID); err == nil {
			ps.rdb.Del(ctx, permCacheKey(userID))
			log.Sugar().Debugw("perm cache invalidated", "user_id", userID)
		}
	}
}
