package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"ttanalytic/internal/api/handlers"
	"ttanalytic/internal/config"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Router handles HTTP routing
type Router struct {
	server   *http.Server
	handlers *handlers.Handler
}

func NewRouter(cfg *config.Config, handler *handlers.Handler) *Router {
	r := chi.NewRouter()

	// health
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/videos/track", handler.TrackVideo)
		//to do
	})

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	server := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.ServerOpts.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.ServerOpts.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.ServerOpts.IdleTimeout) * time.Second,
	}
	return &Router{
		server:   server,
		handlers: handler,
	}
}

// Start starts the HTTP server
func (r *Router) Start() error {
	return r.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (r *Router) Shutdown(ctx context.Context) error {
	if err := r.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	return nil
}
