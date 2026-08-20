package handler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sall-lah/store_notification/internal/domain"
	"github.com/sall-lah/store_notification/internal/idempotency"
)

// Router inspects incoming Kafka event envelopes, ensures idempotency, and delegates
// execution to the appropriate domain handlers.
type Router struct {
	idempotencyStore idempotency.Store
	authHandler      *AuthHandler
	orderHandler     *OrderHandler
	idempotencyTTL   time.Duration
}

// NewRouter constructs a Router instance with its handler dependencies.
func NewRouter(
	store idempotency.Store,
	authHandler *AuthHandler,
	orderHandler *OrderHandler,
	idempotencyTTL time.Duration,
) *Router {
	return &Router{
		idempotencyStore: store,
		authHandler:      authHandler,
		orderHandler:     orderHandler,
		idempotencyTTL:   idempotencyTTL,
	}
}

// Route processes a single event envelope through idempotency check and domain handler dispatch.
func (r *Router) Route(ctx context.Context, envelope *domain.EventEnvelope) error {
	if err := envelope.Validate(); err != nil {
		return fmt.Errorf("envelope validation failed: %w", err)
	}

	// 1. Idempotency Check: Prevent duplicate email dispatch on Kafka rebalances or retries
	acquired, err := r.idempotencyStore.Acquire(ctx, envelope.EventID, r.idempotencyTTL)
	if err != nil {
		return fmt.Errorf("idempotency check failed: %w", err)
	}
	if !acquired {
		slog.Info("Dropping duplicate event (already processed)", "event_id", envelope.EventID, "type", envelope.EventType)
		return nil
	}

	// 2. Dispatch to matching domain event handler
	var handleErr error
	switch envelope.EventType {
	case domain.EventTypeAuthRegistrationOTP:
		var data *domain.AuthOtpEventData
		if data, handleErr = envelope.ParseAuthOTP(); handleErr == nil {
			handleErr = r.authHandler.HandleRegistrationOTP(ctx, data)
		}

	case domain.EventTypeAuthPasswordResetOTP:
		var data *domain.AuthOtpEventData
		if data, handleErr = envelope.ParseAuthOTP(); handleErr == nil {
			handleErr = r.authHandler.HandlePasswordResetOTP(ctx, data)
		}

	case domain.EventTypeOrderCreated:
		var data *domain.OrderEventData
		if data, handleErr = envelope.ParseOrderEvent(); handleErr == nil {
			handleErr = r.orderHandler.HandleOrderCreated(ctx, data)
		}

	case domain.EventTypeOrderPaid:
		var data *domain.OrderEventData
		if data, handleErr = envelope.ParseOrderEvent(); handleErr == nil {
			handleErr = r.orderHandler.HandleOrderPaid(ctx, data)
		}

	case domain.EventTypeOrderCancelled:
		var data *domain.OrderEventData
		if data, handleErr = envelope.ParseOrderEvent(); handleErr == nil {
			handleErr = r.orderHandler.HandleOrderCancelled(ctx, data)
		}

	case domain.EventTypeOrderFulfilled:
		var data *domain.OrderEventData
		if data, handleErr = envelope.ParseOrderEvent(); handleErr == nil {
			handleErr = r.orderHandler.HandleOrderFulfilled(ctx, data)
		}

	default:
		slog.Warn("Unhandled event type received; acknowledging without action", "event_type", envelope.EventType, "event_id", envelope.EventID)
		return nil
	}

	// If handler failed, release lock to allow potential retry if upstream Kafka message is retried
	if handleErr != nil {
		_ = r.idempotencyStore.Release(ctx, envelope.EventID)
		return fmt.Errorf("handler failed for event %s (%s): %w", envelope.EventID, envelope.EventType, handleErr)
	}

	slog.Info("Successfully processed event", "event_id", envelope.EventID, "type", envelope.EventType)
	return nil
}
