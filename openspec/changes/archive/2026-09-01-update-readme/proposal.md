## Why

The `store_notification` microservice currently lacks a comprehensive, standardized `README.md` matching the documentation standard established across the store microservices ecosystem (e.g., `store_order`). A complete, production-grade README is essential for onboarding developers, detailing architecture flowcharts, explaining Kafka consumer event schemas, documenting environment configurations, and illustrating local testing tools.

## What Changes

- Add a standardized, comprehensive `README.md` in the repository root matching the ecosystem documentation design and badges.
- Include complete Mermaid architecture overview and event consumption pipeline diagrams.
- Document all Kafka inbound event topics (`auth.events`, `order.events`), payload structures, and dispatch targets.
- Detail the Redis sliding-window idempotency mechanism and fallback behaviors.
- Document email templates and notification scenarios (Customer OTP, Order lifecycle, Admin alerts).
- Document configuration parameters (`.env.example`), local development steps, CLI simulation tools (`cmd/kafka-producer`, `cmd/test-send`), unit testing, and Docker deployment.

## Capabilities

### New Capabilities
- `project-documentation`: Standardized repository documentation, architecture diagrams, Kafka event catalog, configuration matrix, and developer simulation workflows.

### Modified Capabilities

## Impact

- Repository documentation (`README.md`).
- No breaking changes to existing Go code, HTTP endpoints, or Kafka consumers.
