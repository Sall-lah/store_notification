package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEventEnvelopeValidation(t *testing.T) {
	envelope := EventEnvelope{
		EventID:   "evt-123",
		EventType: EventTypeAuthRegistrationOTP,
		Timestamp: time.Now(),
		Producer:  "store_auth",
		Data:      json.RawMessage(`{"email":"user@test.com","code":"123456"}`),
	}

	if err := envelope.Validate(); err != nil {
		t.Fatalf("expected valid envelope, got error: %v", err)
	}

	authData, err := envelope.ParseAuthOTP()
	if err != nil {
		t.Fatalf("failed to parse auth OTP: %v", err)
	}

	if authData.Email != "user@test.com" || authData.Code != "123456" {
		t.Errorf("unexpected auth payload: %+v", authData)
	}
}

func TestOrderEventParsing(t *testing.T) {
	orderJSON := `{
		"id": "ord-999",
		"orderNumber": "ORD-2026-0001",
		"userId": "usr-1",
		"userEmail": "customer@test.com",
		"status": "PAID",
		"totalAmount": 150.50,
		"items": [
			{"productName": "Sneakers", "quantity": 1, "price": 150.50, "subtotal": 150.50}
		]
	}`

	envelope := EventEnvelope{
		EventID:   "evt-456",
		EventType: EventTypeOrderPaid,
		Timestamp: time.Now(),
		Producer:  "store_order",
		Data:      json.RawMessage(orderJSON),
	}

	orderData, err := envelope.ParseOrderEvent()
	if err != nil {
		t.Fatalf("failed to parse order event: %v", err)
	}

	if orderData.OrderNumber != "ORD-2026-0001" || orderData.TotalAmount != 150.50 {
		t.Errorf("unexpected order data: %+v", orderData)
	}
	if len(orderData.Items) != 1 || orderData.Items[0].ProductName != "Sneakers" {
		t.Errorf("unexpected order items: %+v", orderData.Items)
	}
}
