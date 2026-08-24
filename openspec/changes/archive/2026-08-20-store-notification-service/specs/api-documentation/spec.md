## ADDED Requirements

### Requirement: OpenAPI 3.1 Specification Serving
The system SHALL expose an OpenAPI 3.1 compliant specification in both YAML (`/docs/notifications/openapi.yaml`) and JSON (`/docs/notifications/openapi.json`) formats defining health endpoints and Kafka event schema definitions.

#### Scenario: Requesting OpenAPI specification in YAML
- **WHEN** a client or API Gateway requests `GET /docs/notifications/openapi.yaml`
- **THEN** the system returns HTTP 200 with `Content-Type: application/yaml` containing the raw OpenAPI 3.1 specification.

#### Scenario: Requesting OpenAPI specification in JSON
- **WHEN** a client or API Gateway requests `GET /docs/notifications/openapi.json`
- **THEN** the system returns HTTP 200 with `Content-Type: application/json` containing the OpenAPI 3.1 specification.

### Requirement: Interactive Scalar and Swagger UI Documentation
The system SHALL serve an interactive Scalar UI at `/docs/notifications/scalar` (or `/docs/notifications/`) and Swagger UI at `/docs/notifications/swagger` referencing the OpenAPI specification.

#### Scenario: Accessing interactive documentation via browser
- **WHEN** a user or gateway navigates to `GET /docs/notifications/scalar`
- **THEN** the system serves the standalone Scalar API documentation web page.

### Requirement: Health and Readiness Probes
The system SHALL provide `/health` and `/ready` HTTP endpoints reporting operational status, Kafka consumer status, Redis connectivity, and SMTP readiness.

#### Scenario: Health probe check
- **WHEN** a client or container orchestrator requests `GET /health`
- **THEN** the system returns HTTP 200 with JSON status `{"status":"UP","service":"store_notification"}`.
