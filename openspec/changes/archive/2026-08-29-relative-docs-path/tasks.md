## 1. Documentation HTML Handlers

- [x] 1.1 Update Scalar UI template in `internal/http/docs.go` to use relative spec path `./openapi.yaml`
- [x] 1.2 Update Swagger UI template in `internal/http/docs.go` to use relative spec path `./openapi.yaml`

## 2. OpenAPI Specification Configuration

- [x] 2.1 Update server URLs in `docs/openapi.yaml` to include relative path `./`
- [x] 2.2 Update server URLs in `docs/openapi.json` to include relative path `./`

## 3. Testing and Verification

- [x] 3.1 Update HTTP docs endpoint tests in `internal/http/http_test.go` to verify relative spec path
- [x] 3.2 Run test suite `go test ./internal/http` to confirm all tests pass
