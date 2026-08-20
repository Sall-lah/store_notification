package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sall-lah/store_notification/internal/config"
	"github.com/sall-lah/store_notification/internal/consumer"
	"github.com/sall-lah/store_notification/internal/handler"
	httpInternal "github.com/sall-lah/store_notification/internal/http"
	"github.com/sall-lah/store_notification/internal/idempotency"
	"github.com/sall-lah/store_notification/internal/mailer"
	"github.com/sall-lah/store_notification/internal/template"
)

func main() {
	// Initialize structured JSON logging for production-grade observability
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Bootstrapping Store Notification Microservice")

	// 1. Load application configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// 2. Initialize Redis Idempotency Store
	var idempotencyStore idempotency.Store
	redisStore, err := idempotency.NewRedisStore(cfg.RedisURL, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		slog.Warn("Could not connect to Redis at startup; falling back to in-memory idempotency store for local operation", "redis_url", cfg.RedisURL, "error", err)
		idempotencyStore = idempotency.NewMemoryStore()
	} else {
		idempotencyStore = redisStore
		slog.Info("Connected to Redis idempotency cache", "redis_url", cfg.RedisURL)
	}
	defer idempotencyStore.Close()

	// 3. Initialize Template Renderer
	templateRenderer := template.NewRenderer()

	// 4. Initialize SMTP Mailer
	smtpMailer, err := mailer.NewSMTPMailer(cfg)
	if err != nil {
		slog.Error("Failed to initialize SMTP mailer client", "error", err)
		os.Exit(1)
	}
	defer smtpMailer.Close()

	// 5. Initialize Domain Event Handlers
	authHandler := handler.NewAuthHandler(templateRenderer, smtpMailer, "Store Platform")
	orderHandler := handler.NewOrderHandler(templateRenderer, smtpMailer, cfg.StoreAdminEmails, "Store Platform")

	// 6. Initialize Central Event Router
	eventRouter := handler.NewRouter(idempotencyStore, authHandler, orderHandler, cfg.IdempotencyTTL)

	// 7. Initialize Kafka Consumer Group
	kafkaConsumer := consumer.NewConsumer(cfg, eventRouter)

	// Root cancellation context for background workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Kafka consumer worker loop
	kafkaConsumer.Start(ctx)
	defer kafkaConsumer.Close()

	// 8. Initialize HTTP Server for Health Probes and Gateway Documentation
	healthHandler := httpInternal.NewHealthHandler(idempotencyStore, smtpMailer, kafkaConsumer)
	docsHandler := httpInternal.NewDocsHandler()
	httpRouter := httpInternal.NewRouter(&httpInternal.ServerConfig{
		HealthHandler: healthHandler,
		DocsHandler:   docsHandler,
	})

	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      httpRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP listener in background goroutine
	go func() {
		slog.Info(fmt.Sprintf("HTTP Server listening on http://0.0.0.0:%s", cfg.Port),
			"health", fmt.Sprintf("http://localhost:%s/health", cfg.Port),
			"docs_scalar", fmt.Sprintf("http://localhost:%s/docs/notifications/scalar", cfg.Port),
		)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	// 9. Signal Handling & Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	slog.Info("Shutdown signal received; initiating graceful termination", "signal", sig.String())

	// Stop background Kafka consumers first
	cancel()

	// Gracefully shutdown HTTP server with a 5-second deadline
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Forced HTTP server shutdown", "error", err)
	}

	slog.Info("Store Notification Service terminated cleanly")
}
