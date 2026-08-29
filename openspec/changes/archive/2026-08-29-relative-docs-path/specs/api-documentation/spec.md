## MODIFIED Requirements

### Requirement: Interactive Scalar and Swagger UI Documentation
The system SHALL serve an interactive Scalar UI at `/docs/notifications/scalar` (or `/docs/notifications/`) and Swagger UI at `/docs/notifications/swagger` referencing the OpenAPI specification using relative paths (`./openapi.yaml`) so that documentation functions seamlessly behind reverse proxies and API gateway prefixes.

#### Scenario: Accessing interactive documentation via browser
- **WHEN** a user or gateway navigates to `GET /docs/notifications/scalar` or `GET /docs/notifications/swagger`
- **THEN** the system serves the interactive documentation page configured with relative spec URL `./openapi.yaml`.
