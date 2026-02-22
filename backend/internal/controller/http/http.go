package http

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	srv      http.Server
	l        *slog.Logger
	frontend bool
}

func NewServer(cfg *Config, qr *QueryRouter) *Server {
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.Logger)

	// Basic CORS
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Get("/api/queries", errorHandler(qr.getAll))
	r.Get("/api/queries/stats", errorHandler(qr.getStats))
	r.Get("/api/queries/stats/history", errorHandler(qr.getStatsHistory))
	r.Get("/api/hosts/stats", errorHandler(qr.getHostStats))
	r.Get("/api/domains/stats", errorHandler(qr.getDomainStats))

	r.Get("/api/blocklist/stats", errorHandler(qr.getBlockListStats))

	fe := cfg.ServeFrontend.Or(true)
	if fe {
		serveStatic(r)
	}

	return &Server{
		srv: http.Server{
			Addr:    cfg.Address,
			Handler: r,
		},
		l:        slog.With("component", "httpsrv"),
		frontend: fe,
	}
}

func (s *Server) ID() string {
	return "HTTP-server"
}

func (s *Server) Start() error {
	s.l.Info("Started HTTP server", "address", s.srv.Addr, "frontend", s.frontend)
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http: starting server: %w", err)
	}

	return nil
}

func (s *Server) Stop() error {
	s.l.Info("Stopping HTTP server", "address", s.srv.Addr)
	return s.srv.Close()
}
