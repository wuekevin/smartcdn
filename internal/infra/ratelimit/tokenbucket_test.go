package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucket_AllowsUnderLimit(t *testing.T) {
	b := NewTokenBucket("test", 300)
	for i := 0; i < 10; i++ {
		start := time.Now()
		if err := b.Wait(context.Background()); err != nil {
			t.Fatal(err)
		}
		if time.Since(start) > 50*time.Millisecond {
			t.Error("first 10 requests should not block (bucket starts full)")
		}
	}
}

func TestTokenBucket_BlocksWhenEmpty(t *testing.T) {
	// 1 token per second rate
	b := NewTokenBucket("test-slow", 1) // 1/min → ~0.0167/sec → very slow
	// Drain the bucket (starts with 1 token for 1/min)
	_ = b.Wait(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := b.Wait(ctx)
	if err == nil {
		t.Error("expected context deadline exceeded for exhausted bucket")
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Error("should have returned quickly on context cancel")
	}
}

func TestTokenBucket_ContextCancel(t *testing.T) {
	b := NewTokenBucket("test-cancel", 1)
	_ = b.Wait(context.Background()) // drain

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	err := b.Wait(ctx)
	if err == nil {
		t.Error("expected error on cancelled context")
	}
}
