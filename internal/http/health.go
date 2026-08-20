package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sall-lah/store_notification/internal/consumer"
	"github.com/sall-lah/store_notification/internal/idempotency"
	"github.com/sall-lah/store_notification/internal/mailer"
)

// HealthHandler handles container liveness and subsystem readiness health probes.
type HealthHandler struct {
	store    idempotency.Store
	mailer   mailer.Mailer
	consumer *consumer.Consumer
}

// NewHealthHandler constructs a HealthHandler instance.
func NewHealthHandler(store idempotency.Store, mailer mailer.Mailer, consumer *consumer.Consumer) *HealthHandler {
	return &HealthHandler{
		store:    store,
		mailer:   mailer,
		consumer: consumer,
	}
}

// LivenessProbe responds with HTTP 200 to confirm the application process is running.
func (h *HealthHandler) LivenessProbe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":    "UP",
		"service":   "store_notification",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ReadinessProbe checks connectivity to Redis, SMTP, and Kafka consumer status.
func (h *HealthHandler) ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	subsystems := map[string]string{
		"redis":          "UP",
		"smtp":           "UP",
		"kafka_consumer": "RUNNING",
	}

	allHealthy := true

	if err := h.store.Ping(ctx); err != nil {
		subsystems["redis"] = "DOWN: " + err.Error()
		allHealthy = false
	}

	if err := h.mailer.Ping(ctx); err != nil {
		// Log but note SMTP might be lazy or local
		subsystems["smtp"] = "DOWN: " + err.Error()
		allHealthy = false
	}

	if h.consumer != nil && !h.consumer.IsRunning() {
		subsystems["kafka_consumer"] = "STOPPED"
		allHealthy = false
	}

	w.Header().Set("Content-Type", "application/json")
	if allHealthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	statusStr := "READY"
	if !allHealthy {
		statusStr = "DEGRADED"
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":     statusStr,
		"service":    "store_notification",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"subsystems": subsystems,
	})
}
