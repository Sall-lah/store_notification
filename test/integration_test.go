package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sall-lah/store_notification/internal/domain"
	"github.com/sall-lah/store_notification/internal/handler"
	httpInternal "github.com/sall-lah/store_notification/internal/http"
	"github.com/sall-lah/store_notification/internal/idempotency"
	"github.com/sall-lah/store_notification/internal/mailer"
	"github.com/sall-lah/store_notification/internal/template"
)

func TestEndToEndNotificationFlow(t *testing.T) {
	ctx := context.Background()

	// 1. Setup core components with in-memory stores and mock transport
	store := idempotency.NewMemoryStore()
	mockMailer := mailer.NewMockMailer()
	renderer := template.NewRenderer()

	adminEmails := []string{"admin@store.example.com", "manager@store.example.com"}
	authHandler := handler.NewAuthHandler(renderer, mockMailer, "Store Test E2E")
	orderHandler := handler.NewOrderHandler(renderer, mockMailer, adminEmails, "Store Test E2E")

	eventRouter := handler.NewRouter(store, authHandler, orderHandler, 24*time.Hour)

	// 2. Test Auth Registration Event
	regEvent := &domain.EventEnvelope{
		EventID:   "e2e-reg-001",
		EventType: domain.EventTypeAuthRegistrationOTP,
		Timestamp: time.Now().UTC(),
		Producer:  "store_auth",
		Data:      json.RawMessage(`{"email":"bob@example.com","code":"554433","name":"Bob Smith"}`),
	}

	if err := eventRouter.Route(ctx, regEvent); err != nil {
		t.Fatalf("failed to route registration event: %v", err)
	}

	if mockMailer.GetSentCount() != 1 {
		t.Fatalf("expected 1 email after registration event, got %d", mockMailer.GetSentCount())
	}
	lastMsg := mockMailer.GetLastMessage()
	if !strings.Contains(lastMsg.Subject, "Verify your email address") {
		t.Errorf("unexpected registration subject: %s", lastMsg.Subject)
	}
	if !strings.Contains(lastMsg.HTMLBody, "554433") {
		t.Errorf("expected HTML body to contain OTP code 554433")
	}

	// 3. Test Idempotency: Replaying the same registration event must not send duplicate email
	if err := eventRouter.Route(ctx, regEvent); err != nil {
		t.Fatalf("failed to re-route duplicate registration event: %v", err)
	}
	if mockMailer.GetSentCount() != 1 {
		t.Fatalf("idempotency check failed! duplicate email was sent, count=%d", mockMailer.GetSentCount())
	}

	// 4. Test Order Paid Event (Customer receipt + Admin alert)
	paidEvent := &domain.EventEnvelope{
		EventID:   "e2e-order-paid-001",
		EventType: domain.EventTypeOrderPaid,
		Timestamp: time.Now().UTC(),
		Producer:  "store_order",
		Data: json.RawMessage(`{
			"orderNumber": "ORD-E2E-2026",
			"userEmail": "bob@example.com",
			"status": "PAID",
			"totalAmount": 199.99,
			"shippingAddress": "456 Market St",
			"items": [
				{"productName": "Wireless Headphones", "price": 199.99, "quantity": 1, "subtotal": 199.99}
			]
		}`),
	}

	if err := eventRouter.Route(ctx, paidEvent); err != nil {
		t.Fatalf("failed to route order paid event: %v", err)
	}

	// Total emails should now be 1 (reg) + 2 (order paid: 1 customer + 1 admin) = 3
	if mockMailer.GetSentCount() != 3 {
		t.Fatalf("expected 3 total emails sent, got %d", mockMailer.GetSentCount())
	}

	// 5. Test HTTP Documentation & Health Endpoints
	healthHandler := httpInternal.NewHealthHandler(store, mockMailer, nil)
	docsHandler := httpInternal.NewDocsHandler()
	httpRouter := httpInternal.NewRouter(&httpInternal.ServerConfig{
		HealthHandler: healthHandler,
		DocsHandler:   docsHandler,
	})

	// Check /health
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	httpRouter.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected /health status 200, got %d", rec.Code)
	}

	// Check /docs/notifications/openapi.yaml
	req = httptest.NewRequest(http.MethodGet, "/docs/notifications/openapi.yaml", nil)
	rec = httptest.NewRecorder()
	httpRouter.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected /docs/notifications/openapi.yaml status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Store Notification Microservice API") {
		t.Errorf("expected OpenAPI spec title in response body")
	}
}
