package ratelimit

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/wuekevin/smartcdn/internal/infra/logger"
)

// TokenBucket is a simple per-provider token bucket rate limiter.
// When tokens are exhausted, Wait blocks until a token becomes available
// or the context is cancelled.
type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	label      string
	log        *zap.SugaredLogger
}

// NewTokenBucket creates a bucket allowing ratePerMinute calls per minute.
// label is a human-readable identifier used in log output.
func NewTokenBucket(label string, ratePerMinute int) *TokenBucket {
	rate := float64(ratePerMinute) / 60.0
	return &TokenBucket{
		tokens:     float64(ratePerMinute), // start full
		maxTokens:  float64(ratePerMinute),
		refillRate: rate,
		lastRefill: time.Now(),
		label:      label,
		log:        logger.Named("ratelimit").Sugar(),
	}
}

// Wait blocks until a token is available or ctx is cancelled.
// On success returns nil; returns ctx.Err() on cancellation.
func (b *TokenBucket) Wait(ctx context.Context) error {
	for {
		wait := b.tryAcquire()
		if wait == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
			// retry
		}
	}
}

// tryAcquire refills the bucket and either consumes a token (returning 0)
// or returns the duration to wait before retrying.
func (b *TokenBucket) tryAcquire() time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.lastRefill = now

	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return 0
	}

	// Need to wait for 1 token.
	waitSeconds := (1.0 - b.tokens) / b.refillRate
	waitDur := time.Duration(waitSeconds * float64(time.Second))
	b.log.Debugw("rate limited", "module", b.label, "event", "rate_limited", "waited_ms", waitDur.Milliseconds())
	return waitDur
}
