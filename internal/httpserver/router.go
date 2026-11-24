// Package httpserver assembles the chi router, middleware chain and route
// table for the service. It is deliberately transport-only: business logic
// lives in the handlers package.
package httpserver

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/aircloud/template-backend-go/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Options configures the router.
type Options struct {
	Logger      *slog.Logger
	Version     string
	ServiceName string
	OTelEnabled bool
	// Health is the probe handler; when nil a fresh one is created. Pass it in
	// to flip readiness (SetReady(false)) during graceful shutdown.
	Health *handlers.Health
}

// New builds the fully-wired HTTP handler.
func New(opts Options) http.Handler {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	r := chi.NewRouter()

	// Global middleware chain, ordered outermost-first.
	r.Use(RequestID)
	r.Use(Recover(logger))
	r.Use(SecurityHeaders)
	r.Use(Logging(logger))
	r.Use(middleware.Timeout(30 * time.Second))

	if opts.OTelEnabled {
		r.Use(func(next http.Handler) http.Handler {
			return otelhttp.NewHandler(next, "http.server",
				otelhttp.WithServerName(opts.ServiceName))
		})
	}

	health := opts.Health
	if health == nil {
		health = handlers.NewHealth(opts.Version)
	}
	r.Get("/healthz", health.Live)
	r.Get("/readyz", health.Ready)

	widgets := handlers.NewWidgets(logger)
	r.Route("/api/v1", func(api chi.Router) {
		api.Route("/widgets", func(wr chi.Router) {
			wr.Get("/", widgets.List)
			wr.Post("/", widgets.Create)
			wr.Get("/{id}", widgets.Get)
			wr.Put("/{id}", widgets.Update)
			wr.Delete("/{id}", widgets.Delete)
		})
	})

	r.NotFound(handlers.NotFound)
	r.MethodNotAllowed(handlers.MethodNotAllowed)

	return r
}
