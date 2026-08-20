package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sall-lah/store_notification/internal/config"
	"github.com/sall-lah/store_notification/internal/domain"
	"github.com/sall-lah/store_notification/internal/handler"
	"github.com/sall-lah/store_notification/internal/idempotency"
	"github.com/sall-lah/store_notification/internal/mailer"
	"github.com/sall-lah/store_notification/internal/template"
)

func main() {
	targetEmail := "hellofliqpy1@gmail.com"
	log.Printf("Preparing to send test emails to: %s\n", targetEmail)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set StoreAdminEmails to target test email so admin alerts also arrive at targetEmail
	cfg.StoreAdminEmails = []string{targetEmail}

	log.Printf("Using SMTP Host: %s:%d (From: %s)\n", cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFromEmail)

	renderer := template.NewRenderer()
	smtpMailer, err := mailer.NewSMTPMailer(cfg)
	if err != nil {
		log.Fatalf("Failed to create SMTP mailer: %v", err)
	}
	defer smtpMailer.Close()

	store := idempotency.NewMemoryStore()
	authHandler := handler.NewAuthHandler(renderer, smtpMailer, "Store Platform")
	orderHandler := handler.NewOrderHandler(renderer, smtpMailer, cfg.StoreAdminEmails, "Store Platform")
	router := handler.NewRouter(store, authHandler, orderHandler, 1*time.Hour)

	ctx := context.Background()

	testEvents := []struct {
		name     string
		envelope domain.EventEnvelope
	}{
		{
			name: "1. User Registration OTP",
			envelope: domain.EventEnvelope{
				EventID:   uuid.NewString(),
				EventType: domain.EventTypeAuthRegistrationOTP,
				Timestamp: time.Now().UTC(),
				Producer:  "store_auth",
				Data: json.RawMessage(fmt.Sprintf(`{
					"email": %q,
					"code": "847291",
					"name": "Fliqpy",
					"type": "REGISTRATION"
				}`, targetEmail)),
			},
		},
		{
			name: "2. Password Reset OTP",
			envelope: domain.EventEnvelope{
				EventID:   uuid.NewString(),
				EventType: domain.EventTypeAuthPasswordResetOTP,
				Timestamp: time.Now().UTC(),
				Producer:  "store_auth",
				Data: json.RawMessage(fmt.Sprintf(`{
					"email": %q,
					"code": "391054",
					"name": "Fliqpy",
					"type": "PASSWORD_RESET"
				}`, targetEmail)),
			},
		},
		{
			name: "3. Order Created (Pending Payment)",
			envelope: domain.EventEnvelope{
				EventID:   uuid.NewString(),
				EventType: domain.EventTypeOrderCreated,
				Timestamp: time.Now().UTC(),
				Producer:  "store_order",
				Data: json.RawMessage(fmt.Sprintf(`{
					"id": %q,
					"orderNumber": "ORD-2026-TEST-001",
					"userId": "usr_test_123",
					"userEmail": %q,
					"status": "PENDING_PAYMENT",
					"totalAmount": 150.00,
					"shippingFee": 10.00,
					"shippingAddress": "Jl. Sudirman No. 45, Jakarta Pusat, DKI Jakarta",
					"paymentType": "midtrans_snap",
					"snapRedirectUrl": "https://app.sandbox.midtrans.com/snap/v2/vtweb/demo-token",
					"items": [
						{
							"productName": "Ergonomic Mechanical Keyboard",
							"variantName": "Wireless / Gateron Brown",
							"sku": "KB-MECH-BRN",
							"price": 140.00,
							"quantity": 1,
							"subtotal": 140.00
						}
					]
				}`, uuid.NewString(), targetEmail)),
			},
		},
		{
			name: "4. Order Paid (Receipt to Customer + Alert to Admin)",
			envelope: domain.EventEnvelope{
				EventID:   uuid.NewString(),
				EventType: domain.EventTypeOrderPaid,
				Timestamp: time.Now().UTC(),
				Producer:  "store_order",
				Data: json.RawMessage(fmt.Sprintf(`{
					"id": %q,
					"orderNumber": "ORD-2026-TEST-002",
					"userId": "usr_test_123",
					"userEmail": %q,
					"status": "PAID",
					"totalAmount": 285.50,
					"shippingFee": 15.50,
					"shippingAddress": "Jl. Sudirman No. 45, Jakarta Pusat, DKI Jakarta",
					"paymentType": "credit_card",
					"items": [
						{
							"productName": "Ultra-light Wireless Gaming Mouse",
							"variantName": "Matte White",
							"sku": "MS-WL-WHT",
							"price": 85.00,
							"quantity": 2,
							"subtotal": 170.00
						},
						{
							"productName": "Desk Mat XL",
							"variantName": "900x400mm Topo Dark",
							"sku": "MAT-XL-TOPO",
							"price": 30.00,
							"quantity": 1,
							"subtotal": 30.00
						}
					]
				}`, uuid.NewString(), targetEmail)),
			},
		},
		{
			name: "5. Order Cancelled",
			envelope: domain.EventEnvelope{
				EventID:   uuid.NewString(),
				EventType: domain.EventTypeOrderCancelled,
				Timestamp: time.Now().UTC(),
				Producer:  "store_order",
				Data: json.RawMessage(fmt.Sprintf(`{
					"id": %q,
					"orderNumber": "ORD-2026-TEST-003",
					"userId": "usr_test_123",
					"userEmail": %q,
					"status": "CANCELLED",
					"totalAmount": 99.00,
					"reason": "Payment window expired after 24 hours"
				}`, uuid.NewString(), targetEmail)),
			},
		},
		{
			name: "6. Order Fulfilled (Dispatched / Shipping)",
			envelope: domain.EventEnvelope{
				EventID:   uuid.NewString(),
				EventType: domain.EventTypeOrderFulfilled,
				Timestamp: time.Now().UTC(),
				Producer:  "store_order",
				Data: json.RawMessage(fmt.Sprintf(`{
					"id": %q,
					"orderNumber": "ORD-2026-TEST-004",
					"userId": "usr_test_123",
					"userEmail": %q,
					"status": "FULFILLED",
					"totalAmount": 285.50,
					"shippingAddress": "Jl. Sudirman No. 45, Jakarta Pusat, DKI Jakarta"
				}`, uuid.NewString(), targetEmail)),
			},
		},
	}

	successCount := 0
	for _, item := range testEvents {
		fmt.Printf("\n--> Dispatching [%s]...\n", item.name)
		if err := router.Route(ctx, &item.envelope); err != nil {
			log.Printf("❌ Failed to send [%s]: %v\n", item.name, err)
		} else {
			fmt.Printf("✅ Sent [%s] successfully!\n", item.name)
			successCount++
		}
		// Brief pause between emails to avoid hitting rapid rate limits
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("\n========================================\n")
	fmt.Printf("Dispatch complete: %d/%d test scenarios processed.\n", successCount, len(testEvents))
	fmt.Printf("========================================\n")
}
