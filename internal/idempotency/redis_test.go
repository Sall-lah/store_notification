package idempotency

import (
	"context"
	"testing"
	"time"
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
