package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Standard Event Type Constants used across the store microservices ecosystem.
const (
	EventTypeAuthRegistrationOTP  = "auth.registration_otp"
	EventTypeAuthPasswordResetOTP = "auth.password_reset_otp"

	EventTypeOrderCreated   = "order.created"
	EventTypeOrderPaid      = "order.paid"
	EventTypeOrderCancelled = "order.cancelled"
	EventTypeOrderExpired   = "order.expired"
	EventTypeOrderFulfilled = "order.fulfilled"
)

// EventEnvelope represents the standard JSON message envelope received from Kafka topics.
// It decouples message transport metadata from domain-specific payloads.
type EventEnvelope struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Timestamp time.Time       `json:"timestamp"`
	Producer  string          `json:"producer"`
	Data      json.RawMessage `json:"data"`
}

// Validate ensures that the event envelope has required structural metadata.
func (e *EventEnvelope) Validate() error {
	if e.EventID == "" {
		return errors.New("missing event_id in event envelope")
	}
	if e.EventType == "" {
		return errors.New("missing event_type in event envelope")
	}
	if len(e.Data) == 0 {
		return errors.New("empty data payload in event envelope")
	}
	return nil
}

// AuthOtpEventData represents the payload emitted by store_auth for OTP dispatch.
type AuthOtpEventData struct {
	Email string `json:"email"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Type  string `json:"type"` // REGISTRATION or PASSWORD_RESET
}

// Validate ensures that email and OTP verification code are present.
func (a *AuthOtpEventData) Validate() error {
	if a.Email == "" {
		return errors.New("missing email in auth OTP event")
	}
	if a.Code == "" {
		return errors.New("missing code in auth OTP event")
	}
	return nil
}

// OrderItemData represents individual line item details within an order event.
type OrderItemData struct {
	ID          string  `json:"id,omitempty"`
	ProductID   string  `json:"productId,omitempty"`
	VariantID   string  `json:"variantId,omitempty"`
	ProductName string  `json:"productName"`
	VariantName string  `json:"variantName,omitempty"`
	SKU         string  `json:"sku,omitempty"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}

// OrderEventData represents the order lifecycle payload emitted by store_order outbox.
type OrderEventData struct {
	ID              string          `json:"id"`
	OrderNumber     string          `json:"orderNumber"`
	UserID          string          `json:"userId"`
	UserEmail       string          `json:"userEmail"`
	Status          string          `json:"status"`
	TotalAmount     float64         `json:"totalAmount"`
	ShippingFee     float64         `json:"shippingFee,omitempty"`
	ShippingAddress string          `json:"shippingAddress,omitempty"`
	PaymentType     string          `json:"paymentType,omitempty"`
	SnapRedirectURL string          `json:"snapRedirectUrl,omitempty"`
	PaidAt          *time.Time      `json:"paidAt,omitempty"`
	CreatedAt       time.Time       `json:"createdAt,omitempty"`
	Items           []OrderItemData `json:"items,omitempty"`
	Reason          string          `json:"reason,omitempty"`
}

// Validate verifies that the order event contains identifiable order information.
func (o *OrderEventData) Validate() error {
	if o.OrderNumber == "" && o.ID == "" {
		return errors.New("order event must have either orderNumber or id")
	}
	if o.UserEmail == "" {
		return errors.New("missing userEmail in order event")
	}
	return nil
}

// ParseAuthOTP extracts AuthOtpEventData from raw JSON data.
func (e *EventEnvelope) ParseAuthOTP() (*AuthOtpEventData, error) {
	var data AuthOtpEventData
	if err := json.Unmarshal(e.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth OTP data: %w", err)
	}
	if err := data.Validate(); err != nil {
		return nil, err
	}
	return &data, nil
}

// ParseOrderEvent extracts OrderEventData from raw JSON data.
func (e *EventEnvelope) ParseOrderEvent() (*OrderEventData, error) {
	var data OrderEventData
	if err := json.Unmarshal(e.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order event data: %w", err)
	}
	if err := data.Validate(); err != nil {
		return nil, err
	}
	return &data, nil
}
