package http

import (
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Start(wg *sync.WaitGroup, address string) {
	defer wg.Done()

	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.Logger)

	qr := queryRouter{}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Get("/api/queries", errorHandler(qr.getAll))

	log.Printf("INFO HTTP server listening at %s\n", address)
	http.ListenAndServe(address, r)
}
