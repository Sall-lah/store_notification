## 1. Documentation Setup & Badges

- [x] 1.1 Draft header, technology badges (Go 1.26+, Chi v5, Kafka, Redis, SMTP, OpenAPI), and Table of Contents in `README.md`
- [x] 1.2 Document service summary, purpose, and key features

## 2. Architecture & Pipeline Diagrams

- [x] 2.1 Construct Mermaid architecture flowchart for Kafka event ingestion, Redis idempotency store, and SMTP dispatch
- [x] 2.2 Construct Mermaid event processing lifecycle & idempotency check state diagram

## 3. Technology Stack, Directory Structure & Configuration

- [x] 3.1 Document technology stack and complete repository directory tree
- [x] 3.2 Document prerequisites and comprehensive environment configuration (`.env.example` matrix)

## 4. Event Schemas, Templates & API Endpoints

- [x] 4.1 Document HTTP endpoints for health/readiness probes and Scalar/Swagger OpenAPI documentation
- [x] 4.2 Document inbound Kafka topics (`auth.events`, `order.events`) with event schemas and payload examples
- [x] 4.3 Document Redis idempotency deduplication strategy and fallback behavior
- [x] 4.4 Document HTML and plain-text email templates for Auth and Order scenarios

## 5. Developer Guide, Simulation Tools & Deployment

- [x] 5.1 Document local development startup steps
- [x] 5.2 Document developer CLI simulation tools (`cmd/kafka-producer`, `cmd/test-send`)
- [x] 5.3 Document unit testing and Docker container deployment
