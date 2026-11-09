package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/wuekevin/smartcdn/internal/infra/logger"
)

type Config struct {
	Addr     string
	Password string
	DB       int
}

// New creates and pings a Redis client.
func New(cfg Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("connecting to Redis: %w", err)
	}

	logger.Named("redis").Info("Redis connected")
	return rdb, nil
}
