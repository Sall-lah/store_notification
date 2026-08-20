package idempotency

import (
	"context"
	"testing"
	"time"

	"github.com/sall-lah/store_notification/internal/config"
)

func TestMemoryStoreIdempotency(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	eventID := "evt-unique-001"
	ttl := 1 * time.Second

	// First attempt should succeed
	acquired, err := store.Acquire(ctx, eventID, ttl)
	if err != nil {
		t.Fatalf("unexpected error on first acquire: %v", err)
	}
	if !acquired {
		t.Errorf("expected first acquire to succeed")
	}

	// Second attempt with same eventID should be rejected (idempotent duplicate)
	acquiredAgain, err := store.Acquire(ctx, eventID, ttl)
	if err != nil {
		t.Fatalf("unexpected error on second acquire: %v", err)
	}
	if acquiredAgain {
		t.Errorf("expected duplicate acquire to be rejected")
	}

	// Release and retry should succeed
	if err := store.Release(ctx, eventID); err != nil {
		t.Fatalf("failed to release event key: %v", err)
	}

	acquiredAfterRelease, err := store.Acquire(ctx, eventID, ttl)
	if err != nil {
		t.Fatalf("unexpected error after release: %v", err)
	}
	if !acquiredAfterRelease {
		t.Errorf("expected acquire to succeed after release")
	}
}

// TestRedisStoreLive tests the actual Redis instance connection and authentication
// when a Redis server is reachable according to environment variables.
func TestRedisStoreLive(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Skip("skipping live Redis test: configuration load error")
	}

	// Attempt connecting to the configured Redis instance
	store, err := NewRedisStore(cfg.RedisURL, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		t.Fatalf("failed to connect/authenticate to live Redis at %s: %v", cfg.RedisURL, err)
	}
	defer store.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Verify Ping
	if err := store.Ping(ctx); err != nil {
		t.Fatalf("failed to ping redis: %v", err)
	}

	eventID := "live-test-evt-" + time.Now().Format("20060102150405.000000")
	ttl := 5 * time.Second

	// Acquire lock
	ok, err := store.Acquire(ctx, eventID, ttl)
	if err != nil || !ok {
		t.Fatalf("failed to acquire live redis lock: %v, ok=%v", err, ok)
	}

	// Duplicate acquire must fail
	dup, err := store.Acquire(ctx, eventID, ttl)
	if err != nil || dup {
		t.Fatalf("expected duplicate lock to be rejected, got dup=%v, err=%v", dup, err)
	}

	// Release lock
	if err := store.Release(ctx, eventID); err != nil {
		t.Fatalf("failed to release live redis lock: %v", err)
	}
}

