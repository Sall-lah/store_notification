package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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
	emailFlag := flag.String("email", "hellofliqy1@gmail.com", "Target email address for test notifications")
	scenarioFlag := flag.String("scenario", "all", "Specific scenario to test (e.g. cancelled, expired, paid, created, all)")
	flag.Parse()

	targetEmail := strings.TrimSpace(*emailFlag)
	if targetEmail == "" {
		targetEmail = "hellofliqy1@gmail.com"
	}

	log.Printf("Preparing to send test emails to: %s (Scenario filter: %s)\n", targetEmail, *scenarioFlag)

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
		id       string
		name     string
		envelope domain.EventEnvelope
	}{
		{
			id:   "reg_otp",
			name: "1. User Registration OTP",
			envelope: domain.EventEnvelope{
				EventID:   uuid.NewString(),
				EventType: domain.EventTypeAuthRegistrationOTP,
				Timestamp: time.Now().UTC(),
				Producer:  "store_auth",
				Data: json.RawMessage(fmt.Sprintf(`{
					"email": %q,
					"code": "847291",
					"name": "Customer",
					"type": "REGISTRATION"
				}`, targetEmail)),
			},
		},
		{
			id:   "reset_otp",
			name: "2. Password Reset OTP",
			envelope: domain.EventEnvelope{
				EventID:   uuid.NewString(),
				EventType: domain.EventTypeAuthPasswordResetOTP,
				Timestamp: time.Now().UTC(),
				Producer:  "store_auth",
				Data: json.RawMessage(fmt.Sprintf(`{
					"email": %q,
					"code": "391054",
					"name": "Customer",
					"type": "PASSWORD_RESET"
				}`, targetEmail)),
			},
		},
		{
			id:   "created",
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
			id:   "paid",
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
			id:   "cancelled",
			name: "5. Order Cancelled (by User)",
			envelope: domain.EventEnvelope{
				EventID:   uuid.NewString(),
				EventType: domain.EventTypeOrderCancelled,
				Timestamp: time.Now().UTC(),
				Producer:  "store_order",
				Data: json.RawMessage(fmt.Sprintf(`{
					"id": %q,
					"orderNumber": %q,
					"userId": "usr_test_123",
					"userEmail": %q,
					"status": "CANCELLED",
					"totalAmount": 129.50,
					"reason": "Cancelled by customer before payment"
				}`, uuid.NewString(), fmt.Sprintf("ORD-%s-CANCEL-%s", time.Now().Format("150405"), strings.ToUpper(uuid.NewString()[:4])), targetEmail)),
			},
		},
		{
			id:   "expired",
			name: "6. Order Expired (Payment Window Expired)",
			envelope: domain.EventEnvelope{
				EventID:   uuid.NewString(),
				EventType: domain.EventTypeOrderExpired,
				Timestamp: time.Now().UTC(),
				Producer:  "store_order",
				Data: json.RawMessage(fmt.Sprintf(`{
					"id": %q,
					"orderNumber": %q,
					"userId": "usr_test_123",
					"userEmail": %q,
					"status": "EXPIRED",
					"totalAmount": 175.00,
					"reason": "Payment window expired after 24 hours (items returned to stock)"
				}`, uuid.NewString(), fmt.Sprintf("ORD-%s-EXPIRE-%s", time.Now().Format("150405"), strings.ToUpper(uuid.NewString()[:4])), targetEmail)),
			},
		},
		{
			id:   "fulfilled",
			name: "7. Order Fulfilled (Dispatched / Shipping)",
			envelope: domain.EventEnvelope{
				EventID:   uuid.NewString(),
				EventType: domain.EventTypeOrderFulfilled,
				Timestamp: time.Now().UTC(),
				Producer:  "store_order",
				Data: json.RawMessage(fmt.Sprintf(`{
					"id": %q,
					"orderNumber": "ORD-2026-TEST-005",
					"userId": "usr_test_123",
					"userEmail": %q,
					"status": "FULFILLED",
					"totalAmount": 285.50,
					"shippingAddress": "Jl. Sudirman No. 45, Jakarta Pusat, DKI Jakarta"
				}`, uuid.NewString(), targetEmail)),
			},
		},
	}

	selectedScenario := strings.ToLower(*scenarioFlag)
	var executedCount int
	var successCount int

	for _, item := range testEvents {
		if selectedScenario != "all" && item.id != selectedScenario && !strings.Contains(strings.ToLower(item.name), selectedScenario) {
			continue
		}

		executedCount++
		fmt.Printf("\n--> Dispatching [%s] to %s...\n", item.name, targetEmail)
		if err := router.Route(ctx, &item.envelope); err != nil {
			log.Printf("❌ Failed to send [%s]: %v\n", item.name, err)
		} else {
			fmt.Printf("✅ Sent [%s] successfully!\n", item.name)
			successCount++
		}
		// Brief pause between emails to avoid hitting rapid rate limits
		time.Sleep(600 * time.Millisecond)
	}

	fmt.Printf("\n========================================\n")
	fmt.Printf("Dispatch complete: %d/%d selected scenarios processed successfully to %s.\n", successCount, executedCount, targetEmail)
	fmt.Printf("========================================\n")

	if successCount < executedCount {
		os.Exit(1)
	}
}

