package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sall-lah/store_notification/internal/config"
	"github.com/sall-lah/store_notification/internal/domain"
	"github.com/sall-lah/store_notification/internal/handler"
	"github.com/segmentio/kafka-go"
)

// Consumer manages per-topic Kafka readers, event ingestion, and graceful shutdown.
type Consumer struct {
	readers []*kafka.Reader
	router  *handler.Router
	running atomic.Bool
	wg      sync.WaitGroup
}

// NewConsumer constructs dedicated Kafka readers with isolated group IDs per topic.
func NewConsumer(cfg *config.Config, router *handler.Router) *Consumer {
	dialer := &kafka.Dialer{
		Timeout: 10 * time.Second,
		DialFunc: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err == nil {
				// Try DNS lookup first (works inside Podman/Docker networks)
				if _, lookupErr := net.LookupHost(host); lookupErr != nil {
					// Fallback for host development outside container network
					if host == "kafka" || host == "broker" {
						addr = net.JoinHostPort("127.0.0.1", port)
					}
				}
			}
			var d net.Dialer
			return d.DialContext(ctx, network, addr)
		},
	}

	readers := make([]*kafka.Reader, 0, len(cfg.KafkaTopics))
	for _, topic := range cfg.KafkaTopics {
		topicClean := strings.ReplaceAll(topic, ".", "-")
		topicGroupID := fmt.Sprintf("%s-%s", cfg.KafkaGroupID, topicClean)

		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cfg.KafkaBrokers,
			GroupID:        topicGroupID,
			Topic:          topic,
			Dialer:         dialer,
			MinBytes:       1,
			MaxBytes:       10e6, // 10MB
			CommitInterval: 1 * time.Second,
			StartOffset:    kafka.FirstOffset,
			ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
				slog.Error(fmt.Sprintf(msg, args...))
			}),
		})
		readers = append(readers, r)
	}

	return &Consumer{
		readers: readers,
		router:  router,
	}
}

// Start spawns background worker loops consuming from each topic reader until canceled.
func (c *Consumer) Start(ctx context.Context) {
	c.running.Store(true)

	for _, reader := range c.readers {
		c.wg.Add(1)
		go c.consumeTopic(ctx, reader)
	}
}

// consumeTopic runs the message loop for a specific topic reader.
func (c *Consumer) consumeTopic(ctx context.Context, reader *kafka.Reader) {
	defer c.wg.Done()

	topicName := reader.Config().Topic
	slog.Info("Starting Kafka consumer worker for topic", "topic", topicName, "group", reader.Config().GroupID)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping Kafka consumer worker", "topic", topicName)
			return
		default:
		}

		// ReadMessage automatically fetches, handles consumer group assignment, and commits
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			slog.Error("Failed to read message from Kafka", "topic", topicName, "error", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Process and route message
		c.processMessage(ctx, msg)
	}
}

// processMessage deserializes the raw Kafka value into an EventEnvelope and routes it.
func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) {
	var envelope domain.EventEnvelope
	if err := json.Unmarshal(msg.Value, &envelope); err != nil {
		slog.Error("Failed to deserialize Kafka message into EventEnvelope",
			"topic", msg.Topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
			"raw_value", string(msg.Value),
			"error", err,
		)
		return
	}

	if err := c.router.Route(ctx, &envelope); err != nil {
		slog.Error("Failed to process event envelope",
			"topic", msg.Topic,
			"event_id", envelope.EventID,
			"event_type", envelope.EventType,
			"error", err,
		)
	}
}

// IsRunning indicates if the consumer worker loop is actively running.
func (c *Consumer) IsRunning() bool {
	return c.running.Load()
}

// Close gracefully closes all Kafka readers and waits for workers to finish.
func (c *Consumer) Close() error {
	c.running.Store(false)
	var firstErr error
	for _, r := range c.readers {
		if err := r.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	c.wg.Wait()
	return firstErr
}
