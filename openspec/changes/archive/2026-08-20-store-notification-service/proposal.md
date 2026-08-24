## Why

The e-commerce platform requires a dedicated, decoupled notification service to dispatch transactional emails to customers and store administrators. Handling email delivery and template rendering synchronously inside core domain services (`store_auth` and `store_order`) introduces unnecessary latency, tight coupling, and reliability risks. A standalone Go notification service consuming domain events via Apache Kafka ensures asynchronous, resilient, and non-blocking email delivery with deduplication and unified template management.

## What Changes

- Create a standalone Go notification microservice consuming domain events from Apache Kafka using `segmentio/kafka-go`.
- Implement Redis-backed idempotency guard to prevent duplicate email dispatch during Kafka rebalances and retries.
- Implement an email rendering engine supporting responsive HTML and plain-text templates using Go `html/template`.
- Implement SMTP email dispatcher with connection pooling and configurable retry mechanisms.
- Implement consumers and handlers for Auth events (`auth.registration_otp`, `auth.password_reset_otp`).
- Implement consumers and handlers for Order events (`order.created`, `order.paid`, `order.cancelled`, `order.fulfilled`), delivering customer receipts/updates and store administrator alerts (`STORE_ADMIN_EMAILS`).
- Provide an OpenAPI 3.1 specification, Scalar/Swagger UI documentation, and health check endpoints (`/health`, `/ready`) compatible with the central API Gateway reverse proxy pattern (`/docs/notifications/...`).

## Capabilities

### New Capabilities
- `notification-core`: Standard Kafka event envelope ingestion, type routing, and Redis-based idempotency filtering.
- `email-dispatch`: Template rendering engine and SMTP client abstraction for sending transactional HTML/text emails.
- `auth-notifications`: Email notification workflows for user registration verification OTP and password reset OTP.
- `order-notifications`: Email notification workflows for order creation, payment confirmation, cancellation, and fulfillment updates for customers and store admins.
- `api-documentation`: OpenAPI 3.1 specification, Scalar UI integration, and HTTP health probes for API Gateway proxying.

### Modified Capabilities
<!-- None: Clean slate repository -->

## Impact

- **External Dependencies**: Connects to existing Apache Kafka cluster (topics `auth.events`, `order.events`) and existing Redis instance for idempotency locks.
- **Gateway Integration**: Exposes HTTP documentation endpoints (`/docs/notifications/scalar`, `/docs/notifications/openapi.yaml`, `/docs/notifications/openapi.json`) and health probes (`/health`, `/ready`).
- **Configuration**: Requires environment variables for SMTP server credentials, Kafka broker addresses, Redis URL, and `STORE_ADMIN_EMAILS`.
