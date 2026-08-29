## Why

Currently, interactive documentation interfaces (Scalar and Swagger UI) reference the OpenAPI specification using absolute paths (`/docs/notifications/openapi.yaml`). When the service is hosted behind an API Gateway, Ingress, or reverse proxy mounted under a subpath or path prefix (such as `/services/notifications/`), the browser attempts to fetch the specification from the domain root, resulting in 404 Not Found errors. Serving documentation assets and specification links using relative paths (`./`) ensures documentation works transparently in standalone development, behind reverse proxies, and across arbitrary gateway subpaths.

## What Changes

- Update the Scalar API reference template in `internal/http/docs.go` to reference `./openapi.yaml` (or relative path) instead of `/docs/notifications/openapi.yaml`.
- Update the Swagger UI template in `internal/http/docs.go` to reference `./openapi.yaml` instead of `/docs/notifications/openapi.yaml`.
- Update the OpenAPI 3.1 definitions in `docs/openapi.yaml` and `docs/openapi.json` to configure relative server paths (`url: ./`) for API Gateway proxy deployments.
- Update documentation route tests in `internal/http/http_test.go` to assert relative specification referencing.

## Capabilities

### New Capabilities
<!-- Capabilities being introduced. Replace <name> with kebab-case identifier (e.g., user-auth, data-export, api-rate-limiting). Each creates specs/<name>/spec.md -->

### Modified Capabilities
- `api-documentation`: Update interactive documentation requirement so that Scalar and Swagger UI reference the OpenAPI specification via relative path (`./openapi.yaml`) instead of fixed absolute root paths.

## Impact

- `internal/http/docs.go`: Updated HTML templates for Scalar and Swagger UI.
- `docs/openapi.yaml` & `docs/openapi.json`: Updated `servers` configuration.
- `internal/http/http_test.go`: Updated test assertions.
- `openspec/specs/api-documentation/spec.md`: Delta spec to capture relative path specification requirement.
