package consumer

import (
	"context"
	"testing"
	"time"

	"github.com/sall-lah/store_notification/internal/handler"
	"github.com/sall-lah/store_notification/internal/idempotency"
	"github.com/sall-lah/store_notification/internal/mailer"
	"github.com/sall-lah/store_notification/internal/template"
	"github.com/segmentio/kafka-go"
)

func TestProcessMessage(t *testing.T) {
	renderer := template.NewRenderer()
	mockMailer := mailer.NewMockMailer()
	store := idempotency.NewMemoryStore()

	authHandler := handler.NewAuthHandler(renderer, mockMailer, "Store Test")
	orderHandler := handler.NewOrderHandler(renderer, mockMailer, []string{"admin@test.com"}, "Store Test")
	router := handler.NewRouter(store, authHandler, orderHandler, 1*time.Hour)

	c := &Consumer{
		router: router,
	}

	validMessage := kafka.Message{
		Topic: "auth.events",
		Value: []byte(`{
			"event_id": "kafka-test-01",
			"event_type": "auth.registration_otp",
			"timestamp": "2026-08-20T12:00:00Z",
			"producer": "store_auth",
			"data": {
				"email": "kafka_user@test.com",
				"code": "654321",
				"name": "Kafka User"
			}
		}`),
	}

	ctx := context.Background()
	c.processMessage(ctx, validMessage)

	if mockMailer.GetSentCount() != 1 {
		t.Fatalf("expected 1 email dispatched from kafka message, got %d", mockMailer.GetSentCount())
	}

	last := mockMailer.GetLastMessage()
	if last.To[0] != "kafka_user@test.com" {
		t.Errorf("expected recipient kafka_user@test.com, got %s", last.To[0])
	}
}
