## Context

The `store_notification` microservice is an asynchronous, event-driven notification engine that consumes events from Apache Kafka (`auth.events` and `order.events`), validates message envelopes, enforces distributed idempotency via Redis, renders multi-part transactional emails (HTML and plaintext), and delivers them via SMTP.

To ensure consistency across the microservices ecosystem (as exemplified by `store_order`), `store_notification` requires a standardized `README.md` detailing architecture diagrams, consumer schemas, idempotency mechanisms, email scenarios, environment variables, local development steps, and simulation CLI tools.

## Goals / Non-Goals

**Goals:**
- Create a complete, production-grade `README.md` at the repository root matching the ecosystem standard format and visual styling.
- Provide accurate Mermaid architecture flowcharts representing Kafka consumer loops, Redis idempotency caching, domain handlers, template rendering, SMTP dispatch, and HTTP documentation/health probes.
- Detail all inbound Kafka event schemas (`auth.events` and `order.events`) with trigger conditions, payload structures, and recipient targets.
- Document configuration parameters (`.env.example`), local setup commands, CLI testing tools (`cmd/kafka-producer`, `cmd/test-send`), unit testing, and Docker deployment.

**Non-Goals:**
- Modifying Go source code, handler logic, or Kafka consumer implementations.
- Altering existing OpenAPI specifications or HTML email templates.

## Decisions

### 1. Structure and Badge Layout
- **Decision**: Align badges and sections with the `store_order` documentation structure:
  - Header with technology badges: Go 1.26+, Chi Router, Kafka Streaming, Redis Idempotency, SMTP / go-mail, OpenAPI 3.1 & Scalar.
  - Table of Contents with quick jump links.
  - Mermaid Architecture Overview and Event Processing Sequence diagrams.
  - Key Features, Tech Stack, Repository Directory Tree.
  - Prerequisites & Environment Configuration (`.env`).
  - Getting Started (Local Development) & Health/Docs Endpoints.
  - Kafka Inbound Topics & Event Catalog table.
  - Redis Idempotency & Deduplication Strategy.
  - Transactional Email Templates & Dispatch Scenarios.
  - Developer Simulation Tools (`cmd/test-send`, `cmd/kafka-producer`).
  - Testing (`go test -race -cover ./...`) and Docker Deployment.

### 2. Accurate Architectural Representation
- **Decision**: Clearly illustrate that `store_notification` acts primarily as a Kafka consumer worker service rather than a CRUD API service, with an embedded HTTP server dedicated to health probes (`/health`, `/ready`) and interactive OpenAPI documentation (`/docs/notifications/*`).
- **Rationale**: Accurately reflects `cmd/server/main.go` and `internal/http/server.go`.

### 3. Comprehensive Kafka Event Matrix
- **Decision**: Document all 7 supported event types across `auth.events` (`auth.registration_otp`, `auth.password_reset_otp`) and `order.events` (`order.created`, `order.paid`, `order.cancelled`, `order.expired`, `order.fulfilled`), including dual-dispatch behavior (customer email + admin order alert for `order.paid`).

## Risks / Trade-offs

- **[Risk] Configuration Divergence**: Documented environment variables might diverge from `internal/config/config.go`.
  - **Mitigation**: Verified every variable and default value directly against `internal/config/config.go` and `.env.example`.
- **[Risk] Path / Route Inaccuracies**: Documentation routes could mismatch internal routing.
  - **Mitigation**: Verified routes against `internal/http/server.go` (`/docs/notifications/scalar`, `/docs/notifications/swagger`, `/health`, `/ready`).
