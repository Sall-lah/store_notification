## Context

The store_notification microservice provides interactive documentation via Scalar and Swagger UI under `/docs/notifications`. Currently, the HTML templates embedded in `internal/http/docs.go` reference the OpenAPI definition using absolute domain root paths (`/docs/notifications/openapi.yaml`). In deployment environments where the service is placed behind an API Gateway, Ingress, or reverse proxy prefix (such as `/services/notifications/docs/...`), absolute URLs bypass the gateway prefix, leading to 404 Not Found errors when fetching the OpenAPI spec.

## Goals / Non-Goals

**Goals:**
- Enable interactive documentation (Scalar and Swagger UI) to fetch specifications using relative path `./openapi.yaml`.
- Ensure OpenAPI schema definitions in `docs/openapi.yaml` and `docs/openapi.json` use relative server base paths (`url: ./`) for proxy compatibility.
- Keep the design simple without introducing dynamic templating libraries.

**Non-Goals:**
- Modifying the underlying Chi HTTP route definitions (`/docs/notifications/...` remains unchanged).
- Altering Kafka event schemas or health probe endpoints.

## Decisions

### Decision 1: Relative spec URL in Scalar and Swagger UI templates
- **Choice**: Use `./openapi.yaml` in `data-url` (Scalar) and `url` (Swagger UI).
- **Rationale**: Browser URL resolution (RFC 3986) resolves `./openapi.yaml` against the parent resource directory (`.../docs/notifications/`), making it agnostic to whether the service is served at domain root or under a gateway subpath.
- **Alternatives Considered**:
  - Absolute root `/docs/notifications/openapi.yaml`: Breaks behind path-prefix gateways.
  - Server-side template rendering with dynamic Host/Header injection: Adds runtime parsing overhead and configuration complexity without added benefit.

### Decision 2: Relative server URL in OpenAPI specifications
- **Choice**: Update the OpenAPI `servers` section to specify `url: ./` in place of `url: /`.
- **Rationale**: Enables interactive API testing in Scalar/Swagger to execute requests relative to the gateway mount point rather than domain root.

## Risks / Trade-offs

- **[Risk] Trailing Slash URL Resolution**: If a client requests `/scalar/` instead of `/scalar`, relative path `./openapi.yaml` might resolve to `/scalar/openapi.yaml`.
  - **Mitigation**: Chi routes are configured without trailing slashes on sub-resources, and router mounts both `/scalar` and `/swagger` canonical paths. Tests will verify standard path access.
