package lock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrLockConflict = errors.New("lock already held by another process")

const defaultTTL = 30 * time.Second

// DistributedLock is a Redis-backed mutex for protecting DNS publish jobs.
type DistributedLock struct {
	rdb   *redis.Client
	key   string
	token string
}

// Acquire tries to obtain the lock. Returns ErrLockConflict if already held.
func Acquire(ctx context.Context, rdb *redis.Client, resource string, token string) (*DistributedLock, error) {
	key := "lock:" + resource
	ok, err := rdb.SetNX(ctx, key, token, defaultTTL).Result()
	if err != nil {
		return nil, fmt.Errorf("lock acquire: %w", err)
	}
	if !ok {
		return nil, ErrLockConflict
	}
	return &DistributedLock{rdb: rdb, key: key, token: token}, nil
}

// Release removes the lock only if the token matches (Lua script for atomicity).
func (l *DistributedLock) Release(ctx context.Context) error {
	const script = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
else
  return 0
end`
	_, err := l.rdb.Eval(ctx, script, []string{l.key}, l.token).Result()
	return err
}

// Extend refreshes the lock TTL to prevent expiry during long operations.
func (l *DistributedLock) Extend(ctx context.Context) error {
	const script = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("PEXPIRE", KEYS[1], ARGV[2])
else
  return 0
end`
	ms := fmt.Sprintf("%d", defaultTTL.Milliseconds())
	_, err := l.rdb.Eval(ctx, script, []string{l.key}, l.token, ms).Result()
	return err
}
