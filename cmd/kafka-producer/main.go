package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sall-lah/store_notification/internal/config"
	"github.com/sall-lah/store_notification/internal/domain"
	"github.com/segmentio/kafka-go"
)

// smartDialer creates a Kafka dialer that intercepts Docker hostnames like "kafka" and routes them to 127.0.0.1.
func newSmartDialer() *kafka.Dialer {
	return &kafka.Dialer{
		Timeout: 10 * time.Second,
		DialFunc: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err == nil && (host == "kafka" || host == "broker") {
				addr = net.JoinHostPort("127.0.0.1", port)
			}
			var d net.Dialer
			return d.DialContext(ctx, network, addr)
		},
	}
}

func main() {
	emailFlag := flag.String("email", "cann27089@gmail.com", "Target email address for order notifications")
	orderFlag := flag.String("order", "", "Custom order number (optional)")
	flag.Parse()

	targetEmail := strings.TrimSpace(*emailFlag)
	if targetEmail == "" {
		targetEmail = "cann27089@gmail.com"
	}

	orderNum := strings.TrimSpace(*orderFlag)
	if orderNum == "" {
		orderNum = fmt.Sprintf("ORD-%s-%s", time.Now().Format("20060102-150405"), strings.ToUpper(uuid.NewString()[:4]))
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Connecting to Kafka broker(s) at: %v\n", cfg.KafkaBrokers)
	log.Printf("Testing full order flow for: %s (Order #%s)\n", targetEmail, orderNum)

	dialer := newSmartDialer()

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.KafkaBrokers...),
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireOne,
		AllowAutoTopicCreation: true,
		Transport: &kafka.Transport{
			Dial: dialer.DialFunc,
		},
	}
	defer writer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sampleItems := []domain.OrderItemData{
		{
			ProductName: "Keychron Q1 Pro Wireless Custom Mechanical Keyboard",
			VariantName: "Carbon Black / Red Switch",
			SKU:         "KB-Q1PRO-RED",
			Price:       199.00,
			Quantity:    1,
			Subtotal:    199.00,
		},
		{
			ProductName: "Logitech MX Master 3S",
			VariantName: "Space Gray",
			SKU:         "MS-MXM3S-GRY",
			Price:       99.00,
			Quantity:    1,
			Subtotal:    99.00,
		},
	}

	now := time.Now().UTC()

	// 1. Live Order Created Event on order.events
	createdPayload, _ := json.Marshal(domain.OrderEventData{
		ID:              uuid.NewString(),
		OrderNumber:     orderNum,
		UserID:          "usr_live_test_101",
		UserEmail:       targetEmail,
		Status:          "PENDING_PAYMENT",
		TotalAmount:     308.00,
		ShippingFee:     10.00,
		ShippingAddress: "Sudirman Central Business District (SCBD) Lot 28, Senayan, Kebayoran Baru, Jakarta Selatan, 12190",
		PaymentType:     "midtrans_snap",
		SnapRedirectURL: "https://app.sandbox.midtrans.com/snap/v2/vtweb/demo-payment-token",
		Items:           sampleItems,
		CreatedAt:       now,
	})

	createdEnvelope, _ := json.Marshal(domain.EventEnvelope{
		EventID:   uuid.NewString(),
		EventType: domain.EventTypeOrderCreated,
		Timestamp: now,
		Producer:  "store_order",
		Data:      createdPayload,
	})

	log.Printf("==> [1/3] Publishing [order.created] to topic [order.events]...")
	if err := writer.WriteMessages(ctx, kafka.Message{
		Topic: "order.events",
		Key:   []byte(orderNum),
		Value: createdEnvelope,
	}); err != nil {
		log.Fatalf("❌ Failed to write order.created: %v", err)
	}
	log.Printf("✅ Published order.created event successfully!")

	time.Sleep(2 * time.Second)

	// 2. Live Order Paid Event on order.events
	paidAt := time.Now().UTC()
	paidPayload, _ := json.Marshal(domain.OrderEventData{
		ID:              uuid.NewString(),
		OrderNumber:     orderNum,
		UserID:          "usr_live_test_101",
		UserEmail:       targetEmail,
		Status:          "PAID",
		TotalAmount:     308.00,
		ShippingFee:     10.00,
		ShippingAddress: "Sudirman Central Business District (SCBD) Lot 28, Senayan, Kebayoran Baru, Jakarta Selatan, 12190",
		PaymentType:     "gopay",
		PaidAt:          &paidAt,
		Items:           sampleItems,
	})

	paidEnvelope, _ := json.Marshal(domain.EventEnvelope{
		EventID:   uuid.NewString(),
		EventType: domain.EventTypeOrderPaid,
		Timestamp: time.Now().UTC(),
		Producer:  "store_order",
		Data:      paidPayload,
	})

	log.Printf("==> [2/3] Publishing [order.paid] to topic [order.events]...")
	if err := writer.WriteMessages(ctx, kafka.Message{
		Topic: "order.events",
		Key:   []byte(orderNum),
		Value: paidEnvelope,
	}); err != nil {
		log.Fatalf("❌ Failed to write order.paid: %v", err)
	}
	log.Printf("✅ Published order.paid event successfully!")

	time.Sleep(2 * time.Second)

	// 3. Live Order Fulfilled Event on order.events
	fulfilledPayload, _ := json.Marshal(domain.OrderEventData{
		ID:              uuid.NewString(),
		OrderNumber:     orderNum,
		UserID:          "usr_live_test_101",
		UserEmail:       targetEmail,
		Status:          "FULFILLED",
		TotalAmount:     308.00,
		ShippingFee:     10.00,
		ShippingAddress: "Sudirman Central Business District (SCBD) Lot 28, Senayan, Kebayoran Baru, Jakarta Selatan, 12190",
		PaymentType:     "gopay",
		Items:           sampleItems,
	})

	fulfilledEnvelope, _ := json.Marshal(domain.EventEnvelope{
		EventID:   uuid.NewString(),
		EventType: domain.EventTypeOrderFulfilled,
		Timestamp: time.Now().UTC(),
		Producer:  "store_order",
		Data:      fulfilledPayload,
	})

	log.Printf("==> [3/3] Publishing [order.fulfilled] to topic [order.events]...")
	if err := writer.WriteMessages(ctx, kafka.Message{
		Topic: "order.events",
		Key:   []byte(orderNum),
		Value: fulfilledEnvelope,
	}); err != nil {
		log.Fatalf("❌ Failed to write order.fulfilled: %v", err)
	}
	log.Printf("✅ Published order.fulfilled event successfully!")

	fmt.Println("\n=======================================================")
	fmt.Printf("All lifecycle events (Created -> Paid -> Fulfilled) published to Kafka for %s!\n", targetEmail)
	fmt.Println("Check store_notification_app container logs and email inbox.")
	fmt.Println("=======================================================")
}

