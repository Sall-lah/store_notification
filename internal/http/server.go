package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// ServerConfig wraps dependencies needed to configure the HTTP API router.
type ServerConfig struct {
	HealthHandler *HealthHandler
	DocsHandler   *DocsHandler
}

// NewRouter constructs a configured Chi HTTP handler with standard middleware.
func NewRouter(cfg *ServerConfig) http.Handler {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// CORS configuration for cross-origin documentation and gateway tools
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-User-Id", "X-User-Email", "X-User-Role"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health & readiness probes
	r.Get("/health", cfg.HealthHandler.LivenessProbe)
	r.Get("/ready", cfg.HealthHandler.ReadinessProbe)

	// Direct documentation routes
	r.Route("/docs/notifications", func(docRouter chi.Router) {
		docRouter.Get("/openapi.yaml", cfg.DocsHandler.ServeOpenAPIYAML)
		docRouter.Get("/openapi.json", cfg.DocsHandler.ServeOpenAPIJSON)
		docRouter.Get("/scalar", cfg.DocsHandler.ServeScalar)
		docRouter.Get("/scalar/*", cfg.DocsHandler.ServeScalar)
		docRouter.Get("/swagger", cfg.DocsHandler.ServeSwagger)
		docRouter.Get("/swagger/*", cfg.DocsHandler.ServeSwagger)
		docRouter.Get("/", cfg.DocsHandler.ServeScalar)
	})

	return r
}
