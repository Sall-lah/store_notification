package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config represents the consolidated application runtime configuration.
// It centralizes environment variables to ensure fail-fast behavior on boot.
type Config struct {
	AppEnv         string
	Port           string
	KafkaBrokers   []string
	KafkaGroupID   string
	KafkaTopics    []string
	RedisURL       string
	RedisPassword  string
	RedisDB        int
	IdempotencyTTL time.Duration

	SMTPHost               string
	SMTPPort               int
	SMTPUser               string
	SMTPPassword           string
	SMTPFromEmail          string
	SMTPFromName           string
	SMTPRequireTLS         bool
	SMTPInsecureSkipVerify bool

	StoreAdminEmails []string
}

// Load reads configuration from the environment and optional .env files.
// It provides deterministic defaults for local development while requiring
// critical variables in production environments.
func Load() (*Config, error) {
	// Attempt to load .env file from current or parent directories (useful when running sub-package tests).
	for _, path := range []string{".env", "../.env", "../../.env"} {
		if err := godotenv.Load(path); err == nil {
			break
		}
	}

	smtpHost := getEnv("SMTP_HOST", "localhost")
	port := getEnvInt("SMTP_PORT", 1025)
	if strings.Contains(smtpHost, "gmail.com") && (port == 1025 || port == 0) {
		port = 587
	}

	cfg := &Config{
		AppEnv:                 getEnv("APP_ENV", "development"),
		Port:                   getEnv("PORT", "8070"),
		KafkaBrokers:           getEnvSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		KafkaGroupID:           getEnv("KAFKA_GROUP_ID", "store-notification-group"),
		KafkaTopics:            getEnvSlice("KAFKA_TOPICS", []string{"auth.events", "order.events"}),
		RedisURL:               getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword:          getEnv("REDIS_PASSWORD", ""),
		RedisDB:                getEnvInt("REDIS_DB", 0),
		IdempotencyTTL:         time.Duration(getEnvInt("IDEMPOTENCY_TTL_HOURS", 24)) * time.Hour,
		SMTPHost:               smtpHost,
		SMTPPort:               port,
		SMTPUser:               getEnv("SMTP_USER", ""),
		SMTPPassword:           getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail:          getEnv("SMTP_FROM_EMAIL", "noreply@store.example.com"),
		SMTPFromName:           getEnv("SMTP_FROM_NAME", "Store Notifications"),
		SMTPRequireTLS:         getEnvBool("SMTP_REQUIRE_TLS", false),
		SMTPInsecureSkipVerify: getEnvBool("SMTP_INSECURE_SKIP_VERIFY", false),
		StoreAdminEmails:       getEnvSlice("STORE_ADMIN_EMAILS", []string{"admin@store.example.com"}),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate verifies that critical connection parameters are provided.
func (c *Config) Validate() error {
	if len(c.KafkaBrokers) == 0 || c.KafkaBrokers[0] == "" {
		return fmt.Errorf("KAFKA_BROKERS must specify at least one valid broker host")
	}
	if c.RedisURL == "" {
		return fmt.Errorf("REDIS_URL must not be empty")
	}
	if c.SMTPHost == "" {
		return fmt.Errorf("SMTP_HOST must not be empty")
	}
	if c.SMTPFromEmail == "" {
		return fmt.Errorf("SMTP_FROM_EMAIL must not be empty")
	}
	return nil
}

// getEnv retrieves an environment variable or falls back to a default value.
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return strings.TrimSpace(val)
	}
	return fallback
}

// getEnvInt parses an integer environment variable with fallback handling.
func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
			return parsed
		}
	}
	return fallback
}

// getEnvBool parses a boolean environment variable with fallback handling.
func getEnvBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		parsed, err := strconv.ParseBool(strings.TrimSpace(val))
		if err == nil {
			return parsed
		}
	}
	return fallback
}

// getEnvSlice parses comma-separated environment variables into a trimmed slice.
func getEnvSlice(key string, fallback []string) []string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}
