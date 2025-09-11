package http

import (
	"gohole/internal/registry"
	"log"
	"log/slog"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func Start(wg *sync.WaitGroup, reg *registry.Registry, address string, serverFrontend bool) {
	defer wg.Done()

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

	qr := &queryRouter{
		reg.QueryService,
	}

	r.Get("/api/queries", errorHandler(qr.getAll))
	r.Get("/api/queries/stats", errorHandler(qr.getStats))
	r.Get("/api/queries/stats/history", errorHandler(qr.getStatsHistory))

	r.Get("/api/blocklist/stats", errorHandler(qr.getBlockListStats))

	if serverFrontend {
		serveStatic(r)
	}

	slog.Info("Started HTTP server", "address", address)
	if err := http.ListenAndServe(address, r); err != nil {
		log.Fatal(err)
	}
}
