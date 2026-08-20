package idempotency

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store provides an atomic deduplication lock interface to ensure exactly-once
// notification processing across Kafka rebalances and retries.
type Store interface {
	// Acquire attempts to atomically lock the given event ID.
	// Returns true if the event has NOT been processed yet and the lock is acquired.
	// Returns false if the event was already processed or is currently being processed.
	Acquire(ctx context.Context, eventID string, ttl time.Duration) (bool, error)

	// Release removes the lock for the given event ID (e.g. on unrecoverable transient error).
	Release(ctx context.Context, eventID string) error

	// Ping checks connectivity to the underlying store.
	Ping(ctx context.Context) error

	// Close terminates any active store connections.
	Close() error
}

// RedisStore implements Store using Redis atomic SETNX with TTL.
type RedisStore struct {
	client *redis.Client
	prefix string
}

// NewRedisStore creates and returns a connected RedisStore instance.
func NewRedisStore(redisURL, password string, db int) (*RedisStore, error) {
	opts := &redis.Options{
		Addr:     redisURL,
		Password: password,
		DB:       db,
	}

	client := redis.NewClient(opts)

	// Verify connectivity with a short timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", redisURL, err)
	}

	return &RedisStore{
		client: client,
		prefix: "notif:evt:",
	}, nil
}

// NewFromClient wraps an existing Redis client into an idempotency Store.
func NewFromClient(client *redis.Client) *RedisStore {
	return &RedisStore{
		client: client,
		prefix: "notif:evt:",
	}
}

// Acquire executes SET key value NX EX ttl atomically.
func (r *RedisStore) Acquire(ctx context.Context, eventID string, ttl time.Duration) (bool, error) {
	key := r.prefix + eventID
	now := time.Now().UTC().Format(time.RFC3339)

	ok, err := r.client.SetNX(ctx, key, now, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis idempotency setnx error: %w", err)
	}
	return ok, nil
}

// Release deletes the idempotency key, allowing retry if an upstream error was transient.
func (r *RedisStore) Release(ctx context.Context, eventID string) error {
	key := r.prefix + eventID
	return r.client.Del(ctx, key).Err()
}

// Ping checks Redis health.
func (r *RedisStore) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the Redis connection pool.
func (r *RedisStore) Close() error {
	return r.client.Close()
}

// MemoryStore provides a thread-safe in-memory implementation of Store for local unit testing.
type MemoryStore struct {
	mu   sync.Mutex
	keys map[string]time.Time
}

// NewMemoryStore constructs a new MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		keys: make(map[string]time.Time),
	}
}

// Acquire locks the eventID in memory with expiration.
func (m *MemoryStore) Acquire(ctx context.Context, eventID string, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if exp, exists := m.keys[eventID]; exists && exp.After(now) {
		return false, nil
	}

	m.keys[eventID] = now.Add(ttl)
	return true, nil
}

// Release removes the eventID from memory.
func (m *MemoryStore) Release(ctx context.Context, eventID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.keys, eventID)
	return nil
}

// Ping always succeeds for in-memory store.
func (m *MemoryStore) Ping(ctx context.Context) error {
	return nil
}

// Close clears the in-memory map.
func (m *MemoryStore) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keys = make(map[string]time.Time)
	return nil
}
