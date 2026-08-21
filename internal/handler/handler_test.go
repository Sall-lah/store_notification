package handler

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/sall-lah/store_notification/internal/domain"
	"github.com/sall-lah/store_notification/internal/idempotency"
	"github.com/sall-lah/store_notification/internal/mailer"
	"github.com/sall-lah/store_notification/internal/template"
)

func setupTestRouter() (*Router, *mailer.MockMailer) {
	renderer := template.NewRenderer()
	mockMailer := mailer.NewMockMailer()
	store := idempotency.NewMemoryStore()

	authHandler := NewAuthHandler(renderer, mockMailer, "Store Test")
	orderHandler := NewOrderHandler(renderer, mockMailer, []string{"admin@test.com"}, "Store Test")

	r := NewRouter(store, authHandler, orderHandler, 1*time.Hour)
	return r, mockMailer
}

func TestRouterRegistrationOTP(t *testing.T) {
	router, mockMailer := setupTestRouter()
	ctx := context.Background()

	envelope := &domain.EventEnvelope{
		EventID:   "evt-auth-1",
		EventType: domain.EventTypeAuthRegistrationOTP,
		Timestamp: time.Now(),
		Producer:  "store_auth",
		Data:      json.RawMessage(`{"email":"user@test.com","code":"778899","name":"Alice"}`),
	}

	if err := router.Route(ctx, envelope); err != nil {
		t.Fatalf("unexpected route error: %v", err)
	}

	if mockMailer.GetSentCount() != 1 {
		t.Fatalf("expected 1 email sent, got %d", mockMailer.GetSentCount())
	}

	last := mockMailer.GetLastMessage()
	if !strings.Contains(last.Subject, "Verify your email address") {
		t.Errorf("unexpected subject: %s", last.Subject)
	}
	if !strings.Contains(last.HTMLBody, "778899") {
		t.Errorf("expected OTP code in HTML body")
	}

	// Test idempotency: second identical event should not dispatch email
	if err := router.Route(ctx, envelope); err != nil {
		t.Fatalf("unexpected route error on duplicate: %v", err)
	}
	if mockMailer.GetSentCount() != 1 {
		t.Errorf("duplicate event triggered another email send! count: %d", mockMailer.GetSentCount())
	}
}

func TestRouterOrderPaid(t *testing.T) {
	router, mockMailer := setupTestRouter()
	ctx := context.Background()

	envelope := &domain.EventEnvelope{
		EventID:   "evt-order-paid-1",
		EventType: domain.EventTypeOrderPaid,
		Timestamp: time.Now(),
		Producer:  "store_order",
		Data: json.RawMessage(`{
			"orderNumber": "ORD-9999",
			"userEmail": "customer@test.com",
			"totalAmount": 120.00,
			"shippingAddress": "123 Main St",
			"items": [
				{"productName": "Hoodie", "price": 120.00, "quantity": 1, "subtotal": 120.00}
			]
		}`),
	}

	if err := router.Route(ctx, envelope); err != nil {
		t.Fatalf("unexpected route error: %v", err)
	}

	// Expect 2 emails: 1 for customer, 1 for store admin
	if mockMailer.GetSentCount() != 2 {
		t.Fatalf("expected 2 emails sent (customer + admin), got %d", mockMailer.GetSentCount())
	}

	customerEmail := mockMailer.SentMessages[0]
	if !strings.Contains(customerEmail.Subject, "Payment Confirmed") {
		t.Errorf("unexpected customer subject: %s", customerEmail.Subject)
	}

	adminEmail := mockMailer.SentMessages[1]
	if !strings.Contains(adminEmail.Subject, "Admin Alert") {
		t.Errorf("unexpected admin subject: %s", adminEmail.Subject)
	}
}

func TestRouterOrderCancelled(t *testing.T) {
	router, mockMailer := setupTestRouter()
	ctx := context.Background()

	targetEmail := "hellofliqy1@gmail.com"
	envelope := &domain.EventEnvelope{
		EventID:   "evt-order-cancelled-1",
		EventType: domain.EventTypeOrderCancelled,
		Timestamp: time.Now(),
		Producer:  "store_order",
		Data: json.RawMessage(`{
			"orderNumber": "ORD-2026-CANCEL-001",
			"userEmail": "hellofliqy1@gmail.com",
			"totalAmount": 149.99,
			"reason": "Cancelled by customer before payment"
		}`),
	}

	if err := router.Route(ctx, envelope); err != nil {
		t.Fatalf("unexpected route error: %v", err)
	}

	if mockMailer.GetSentCount() != 1 {
		t.Fatalf("expected 1 email sent, got %d", mockMailer.GetSentCount())
	}

	last := mockMailer.GetLastMessage()
	if len(last.To) != 1 || last.To[0] != targetEmail {
		t.Errorf("expected recipient %s, got %v", targetEmail, last.To)
	}
	if !strings.Contains(last.Subject, "Cancelled") || !strings.Contains(last.Subject, "ORD-2026-CANCEL-001") {
		t.Errorf("unexpected subject: %s", last.Subject)
	}
	if !strings.Contains(last.HTMLBody, "Cancelled") || !strings.Contains(last.HTMLBody, "Cancelled by customer before payment") {
		t.Errorf("expected cancellation details in HTML body: %s", last.HTMLBody)
	}
}

func TestRouterOrderExpired(t *testing.T) {
	router, mockMailer := setupTestRouter()
	ctx := context.Background()

	targetEmail := "hellofliqy1@gmail.com"
	envelope := &domain.EventEnvelope{
		EventID:   "evt-order-expired-1",
		EventType: domain.EventTypeOrderExpired,
		Timestamp: time.Now(),
		Producer:  "store_order",
		Data: json.RawMessage(`{
			"orderNumber": "ORD-2026-EXPIRE-001",
			"userEmail": "hellofliqy1@gmail.com",
			"totalAmount": 299.50,
			"reason": "Payment window expired after 24 hours"
		}`),
	}

	if err := router.Route(ctx, envelope); err != nil {
		t.Fatalf("unexpected route error: %v", err)
	}

	if mockMailer.GetSentCount() != 1 {
		t.Fatalf("expected 1 email sent, got %d", mockMailer.GetSentCount())
	}

	last := mockMailer.GetLastMessage()
	if len(last.To) != 1 || last.To[0] != targetEmail {
		t.Errorf("expected recipient %s, got %v", targetEmail, last.To)
	}
	if !strings.Contains(last.Subject, "Expired") || !strings.Contains(last.Subject, "ORD-2026-EXPIRE-001") {
		t.Errorf("unexpected subject: %s", last.Subject)
	}
	if !strings.Contains(last.HTMLBody, "Payment Window Expired") || !strings.Contains(last.HTMLBody, "Payment window expired after 24 hours") {
		t.Errorf("expected expiration details in HTML body: %s", last.HTMLBody)
	}
}

