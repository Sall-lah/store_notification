# project-documentation Specification

## Purpose
Comprehensive repository documentation, architecture diagrams, Kafka event catalog, configuration matrix, and developer simulation workflows.
## Requirements
### Requirement: Standardized Repository Readme
The repository SHALL maintain a root `README.md` file adhering to the store microservices ecosystem documentation template, containing structured metadata badges, architectural Mermaid diagrams, technology stack, directory hierarchy, configuration matrix, and developer guides.

#### Scenario: Developer views repository documentation
- **WHEN** a developer inspects the root `README.md`
- **THEN** the document displays technology badges (Go 1.26+, Chi v5, Kafka, Redis, SMTP, OpenAPI), system architecture flowcharts, inbound event schemas, environment variable tables, local development instructions, and simulation CLI guides.

### Requirement: Architecture and Event Pipeline Visualizations
The `README.md` SHALL provide Mermaid diagrams modeling the service topology and the idempotency-guaranteed event processing lifecycle.

#### Scenario: Inspecting architectural flow
- **WHEN** reading the Architecture Overview section
- **THEN** the reader can view a Mermaid diagram illustrating Kafka inbound streaming (`auth.events`, `order.events`), Redis idempotency check, template rendering, and SMTP mail dispatch to customer and admin inboxes.

### Requirement: Kafka Event Schema and Template Catalog
The `README.md` SHALL document all 7 supported Kafka event types, their payload properties, and the corresponding transactional email templates dispatched.

#### Scenario: Verifying event contract
- **WHEN** integrating an upstream service (`store_auth` or `store_order`) with `store_notification`
- **THEN** the developer can reference the event catalog table detailing event types (`auth.registration_otp`, `auth.password_reset_otp`, `order.created`, `order.paid`, `order.cancelled`, `order.expired`, `order.fulfilled`), trigger conditions, payloads, and recipient types.

### Requirement: CLI Simulation and Local Testing Documentation
The `README.md` SHALL document the developer CLI tools for Kafka event dispatch simulation (`cmd/kafka-producer`) and direct SMTP template verification (`cmd/test-send`).

#### Scenario: Running local notification simulations
- **WHEN** a developer executes test simulation commands described in `README.md`
- **THEN** they can trigger realistic email dispatches using `go run cmd/test-send/main.go` and `go run cmd/kafka-producer/main.go`.

