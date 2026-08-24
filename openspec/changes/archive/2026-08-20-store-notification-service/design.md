## Context

The store microservices ecosystem (`store_gateway`, `store_auth`, `store_order`, `store_product`) operates in a decoupled, event-driven architecture with Apache Kafka and Redis. `store_notification` is a standalone Go microservice responsible for ingesting asynchronous domain events from Kafka topics and delivering transactional emails to customers and store administrators via SMTP.

The central API Gateway (`store_gateway`) proxies documentation routes (`/docs/notifications/...`) to the notification service's HTTP server and offloads identity headers where needed.

## Goals / Non-Goals

**Goals:**
- Implement a standalone Go service using `segmentio/kafka-go` to consume events from `auth.events` and `order.events`.
- Enforce strict message idempotency using Redis atomic `SETNX` operations with a 24-hour expiration.
- Provide a robust email engine with Go `html/template` rendering responsive, branded HTML emails and plaintext fallbacks.
- Dispatch transactional emails over SMTP supporting STARTTLS, authentication, and configurable sender details.
- Handle Auth events: registration verification OTP and password reset OTP.
- Handle Order events: order created, paid, cancelled, and fulfilled, notifying customers and store admins (`STORE_ADMIN_EMAILS`).
- Serve OpenAPI 3.1 specifications (`.yaml`, `.json`), Scalar UI, Swagger UI, and health probes (`/health`, `/ready`).
- Provide clean, human-readable, modular Go code with comprehensive TSDoc/JSDoc-style Go comments.

**Non-Goals:**
- SMS, Push, or WhatsApp notification channels (deferred to future changes).
- Product inventory/catalog events (explicitly excluded for current scope).
- Database persistence for notification logs (relies on structured logging and Kafka offset tracking).

## Decisions

### 1. Kafka Client: `segmentio/kafka-go`
- **Choice**: `segmentio/kafka-go` pure Go library.
- **Rationale**: Avoids CGO dependencies and cross-compilation friction associated with `confluent-kafka-go` (librdkafka). Provides idiomatic Go `Reader` and `Writer` abstractions with built-in consumer group support and clean graceful cancellation via `context.Context`.
- **Alternatives Considered**: `confluent-kafka-go` (CGO required), `twmb/franz-go` (lower-level API).

### 2. Idempotency Store: Redis `SETNX`
- **Choice**: Atomic key insertion `SET key value NX EX 86400` where `key = notif:evt:{event_id}`.
- **Rationale**: Kafka provides at-least-once delivery; rebalances or restarts can cause identical events to be re-processed. An atomic Redis key check ensures that only one worker processes and sends the email for a given `event_id`.
- **Alternatives Considered**: Relational database table with unique constraint (adds unnecessary RDBMS migration overhead for transient notification deduplication).

### 3. Template Engine: Embedded Go `html/template`
- **Choice**: Native Go `html/template` and `text/template` using `//go:embed` for template files.
- **Rationale**: Compiles directly into the Go binary for single-artifact deployment, zero external runtime dependencies, automatic contextual escaping (XSS prevention), and sub-millisecond render times.
- **Alternatives Considered**: External template service or MJML CLI runtime (adds heavy Node.js or remote API dependency).

### 4. Mailer Transport: SMTP Adapter Pattern
- **Choice**: Standard SMTP client using Go `net/smtp` / `go-mail` wrapped in a `Mailer` domain interface.
- **Rationale**: Enables seamless development with local SMTP catchers (Mailpit, Mailtrap, MailHog) and production deployment with cloud SMTP relays (AWS SES, SendGrid, Postmark) without code changes.

### 5. HTTP & Documentation Gateway Pattern
- **Choice**: Lightweight Chi router serving `/health`, `/ready`, `/docs/notifications/openapi.yaml`, `/docs/notifications/openapi.json`, and `/docs/notifications/scalar`.
- **Rationale**: Matches the existing ecosystem convention established by `store_auth` and `store_order`, allowing the central API Gateway to proxy documentation effortlessly.

## Risks / Trade-offs

- **[Risk] SMTP Server Downtime or Rate Limiting**
  - *Mitigation*: The dispatcher implements exponential backoff retries for transient network/server errors and logs structured errors for alerting.
- **[Risk] Kafka Consumer Group Rebalances Causing Message Re-delivery**
  - *Mitigation*: Redis-backed atomic `SETNX` idempotency guard prevents duplicate email dispatches.
- **[Risk] Missing or Misconfigured Environment Variables**
  - *Mitigation*: The service validates all mandatory configuration parameters (`KAFKA_BROKERS`, `REDIS_URL`, `SMTP_HOST`, etc.) at startup and fails fast with actionable error messages.
- **[Risk] Invalid or Malformed Message Payloads**
  - *Mitigation*: The consumer parses payloads against strict domain schemas, logging errors with raw payload context and committing offsets to prevent blocking partition progress.
