package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"ttanalytic/internal/api/handlers"
	"ttanalytic/internal/config"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Router handles HTTP routing
type Router struct {
	server   *http.Server
	handlers *handlers.Handler
}

func NewRouter(cfg *config.Config, handler *handlers.Handler) *Router {
	r := chi.NewRouter()

	// middleware
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// health
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	//chi router
	r.Route("/api/v1", func(r chi.Router) {
		//create or get video
		r.Post("/videos", handler.TrackVideo)
		r.Get("/videos/{tiktok_id}", handler.GetVideo)
	})

	//server
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
