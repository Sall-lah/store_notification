package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sall-lah/store_notification/internal/idempotency"
	"github.com/sall-lah/store_notification/internal/mailer"
)

func TestHealthAndDocsEndpoints(t *testing.T) {
	store := idempotency.NewMemoryStore()
	mockMailer := mailer.NewMockMailer()
	healthHandler := NewHealthHandler(store, mockMailer, nil)
	docsHandler := NewDocsHandler()

	router := NewRouter(&ServerConfig{
		HealthHandler: healthHandler,
		DocsHandler:   docsHandler,
	})

	// Test /health
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for /health, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "UP") {
		t.Errorf("expected body to contain UP, got %s", rec.Body.String())
	}

	// Test /ready
	req = httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for /ready, got %d", rec.Code)
	}

	// Test /docs/notifications/openapi.yaml
	req = httptest.NewRequest(http.MethodGet, "/docs/notifications/openapi.yaml", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for openapi.yaml, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "openapi: 3.1.0") {
		t.Errorf("expected openapi.yaml content, got: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "url: ./") {
		t.Errorf("expected openapi.yaml to contain relative server url: ./, got: %s", rec.Body.String())
	}

	// Test /docs/notifications/scalar
	req = httptest.NewRequest(http.MethodGet, "/docs/notifications/scalar", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for scalar docs, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "@scalar/api-reference") {
		t.Errorf("expected scalar HTML content, got: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `data-url="./openapi.yaml"`) {
		t.Errorf("expected scalar HTML to contain relative data-url='./openapi.yaml', got: %s", rec.Body.String())
	}

	// Test /docs/notifications/swagger
	req = httptest.NewRequest(http.MethodGet, "/docs/notifications/swagger", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for swagger docs, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `url: "./openapi.yaml"`) {
		t.Errorf("expected swagger HTML to contain relative url: './openapi.yaml', got: %s", rec.Body.String())
	}
}
