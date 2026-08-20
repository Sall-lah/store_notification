package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
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
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	targetEmail := "hellofliqpy1@gmail.com"
	log.Printf("Connecting to Kafka broker(s) at: %v\n", cfg.KafkaBrokers)

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

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. Live Registration OTP Event on auth.events
	authPayload, _ := json.Marshal(domain.AuthOtpEventData{
		Email: targetEmail,
		Code:  "991122",
		Name:  "Kafka Live Tester",
		Type:  "REGISTRATION",
	})

	authEnvelope, _ := json.Marshal(domain.EventEnvelope{
		EventID:   uuid.NewString(),
		EventType: domain.EventTypeAuthRegistrationOTP,
		Timestamp: time.Now().UTC(),
		Producer:  "store_auth",
		Data:      authPayload,
	})

	authMsg := kafka.Message{
		Topic: "auth.events",
		Key:   []byte(targetEmail),
		Value: authEnvelope,
	}

	log.Printf("Publishing live event to Kafka topic [auth.events]...")
	if err := writer.WriteMessages(ctx, authMsg); err != nil {
		log.Fatalf("❌ Failed to write to Kafka topic auth.events: %v", err)
	}
	log.Printf("✅ Published registration OTP event to topic [auth.events] successfully!")

	// 2. Live Order Paid Event on order.events
	orderPayload, _ := json.Marshal(domain.OrderEventData{
		ID:          uuid.NewString(),
		OrderNumber: "ORD-KAFKA-LIVE-888",
		UserID:      "usr_kafka_99",
		UserEmail:   targetEmail,
		Status:      "PAID",
		TotalAmount: 349.00,
		ShippingFee: 0.00,
		ShippingAddress: "Jl. Jend. Sudirman Kav. 52-53, SCBD, Jakarta",
		PaymentType: "qris",
		Items: []domain.OrderItemData{
			{
				ProductName: "4K Gaming Monitor 144Hz",
				VariantName: "27 inch / IPS Panel",
				SKU:         "MON-4K-27-IPS",
				Price:       349.00,
				Quantity:    1,
				Subtotal:    349.00,
			},
		},
	})

	orderEnvelope, _ := json.Marshal(domain.EventEnvelope{
		EventID:   uuid.NewString(),
		EventType: domain.EventTypeOrderPaid,
		Timestamp: time.Now().UTC(),
		Producer:  "store_order",
		Data:      orderPayload,
	})

	orderMsg := kafka.Message{
		Topic: "order.events",
		Key:   []byte("ORD-KAFKA-LIVE-888"),
		Value: orderEnvelope,
	}

	log.Printf("Publishing live event to Kafka topic [order.events]...")
	if err := writer.WriteMessages(ctx, orderMsg); err != nil {
		log.Fatalf("❌ Failed to write to Kafka topic order.events: %v", err)
	}
	log.Printf("✅ Published order.paid event to topic [order.events] successfully!")

	fmt.Println("\n=======================================================")
	fmt.Println("Events published to Kafka! Check store_notification logs.")
	fmt.Println("=======================================================")
}
