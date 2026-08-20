package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigLoad(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("KAFKA_BROKERS", "broker1:9092,broker2:9092")
	os.Setenv("STORE_ADMIN_EMAILS", "admin1@test.com, admin2@test.com")
	os.Setenv("IDEMPOTENCY_TTL_HOURS", "48")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("KAFKA_BROKERS")
		os.Unsetenv("STORE_ADMIN_EMAILS")
		os.Unsetenv("IDEMPOTENCY_TTL_HOURS")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("expected Port 9090, got %s", cfg.Port)
	}
	if len(cfg.KafkaBrokers) != 2 || cfg.KafkaBrokers[0] != "broker1:9092" {
		t.Errorf("expected 2 Kafka brokers, got %v", cfg.KafkaBrokers)
	}
	if len(cfg.StoreAdminEmails) != 2 || cfg.StoreAdminEmails[1] != "admin2@test.com" {
		t.Errorf("expected 2 Admin emails, got %v", cfg.StoreAdminEmails)
	}
	if cfg.IdempotencyTTL != 48*time.Hour {
		t.Errorf("expected 48h TTL, got %v", cfg.IdempotencyTTL)
	}
}
