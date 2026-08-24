## 1. Project Scaffolding & Configuration

- [x] 1.1 Initialize Go module (`go.mod`) with dependencies: `segmentio/kafka-go`, `redis/go-redis/v9`, `go-chi/chi/v5`, `go-mail/mail/v2` (or standard `net/smtp`).
- [x] 1.2 Implement configuration loader (`internal/config/config.go`) reading SMTP credentials, Kafka brokers/topics, Redis URL, server port, and `STORE_ADMIN_EMAILS` from environment.
- [x] 1.3 Create `.env.example` documenting all configuration keys.

## 2. Domain Models & Event Envelopes

- [x] 2.1 Define strongly typed `EventEnvelope` and event payload structs in `internal/domain/event.go` (`AuthOtpEventData`, `OrderEventData`, `OrderItemData`).
- [x] 2.2 Define email message domain models and recipient types in `internal/domain/email.go`.
- [x] 2.3 Write unit tests for envelope serialization and payload validation in `internal/domain/event_test.go`.

## 3. Redis Idempotency Store

- [x] 3.1 Implement `IdempotencyStore` interface and Redis client wrapper in `internal/idempotency/redis.go` using atomic `SETNX` with 24-hour TTL.
- [x] 3.2 Write unit/integration tests for the idempotency guard in `internal/idempotency/redis_test.go`.

## 4. Email Templates & Template Engine

- [x] 4.1 Create base layout wrapper (`templates/layout/base.html`) with responsive HTML styling and plain-text counterpart.
- [x] 4.2 Create Auth OTP email templates (`templates/auth/registration_otp.html`, `templates/auth/password_reset_otp.html` and `.txt`).
- [x] 4.3 Create Order email templates (`templates/order/order_created.html`, `templates/order/order_paid.html`, `templates/order/order_cancelled.html`, `templates/order/order_fulfilled.html`).
- [x] 4.4 Create Store Admin alert template (`templates/admin/admin_order_alert.html`).
- [x] 4.5 Implement template rendering engine in `internal/template/renderer.go` with embedded FS (`//go:embed`) and test in `internal/template/renderer_test.go`.

## 5. SMTP Mailer Transport

- [x] 5.1 Define `Mailer` interface and implement SMTP mailer client in `internal/mailer/smtp.go` supporting TLS/STARTTLS, auth, and retry logic.
- [x] 5.2 Implement a mock mailer for testing in `internal/mailer/mock.go`.
- [x] 5.3 Write unit tests for email composition and header generation in `internal/mailer/smtp_test.go`.

## 6. Event Handlers

- [x] 6.1 Implement `AuthHandler` in `internal/handler/auth.go` to process registration and password reset OTP events.
- [x] 6.2 Implement `OrderHandler` in `internal/handler/order.go` to process order created, paid, cancelled, and fulfilled events (notifying user and store admins).
- [x] 6.3 Implement central event router `internal/handler/router.go` delegating events by `event_type` after idempotency checks.
- [x] 6.4 Write unit tests for event handlers in `internal/handler/handler_test.go`.

## 7. Kafka Consumer Group

- [x] 7.1 Implement Kafka consumer group reader in `internal/consumer/kafka.go` using `segmentio/kafka-go` consuming from `auth.events` and `order.events`.
- [x] 7.2 Implement graceful worker pool dispatching, context cancellation, and offset commit logic.
- [x] 7.3 Write consumer parsing tests and mock listener tests.

## 8. HTTP Server & Documentation

- [x] 8.1 Write OpenAPI 3.1 specification files (`docs/openapi.yaml`, `docs/openapi.json`) defining health endpoints and Kafka event schema definitions.
- [x] 8.2 Implement HTTP handlers for `/health` and `/ready` probes in `internal/http/health.go`.
- [x] 8.3 Implement documentation endpoints serving OpenAPI files and interactive Scalar / Swagger UI in `internal/http/docs.go`.
- [x] 8.4 Set up Chi router in `internal/http/server.go`.

## 9. Application Bootstrap & Verification

- [x] 9.1 Wire dependencies, signal handling, and graceful shutdown in `cmd/server/main.go`.
- [x] 9.2 Add `Dockerfile` and `Makefile` for building and running tests.
- [x] 9.3 Verify end-to-end event handling flow with automated tests.
