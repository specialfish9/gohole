package http

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

func serveStatic(r *chi.Mux) {
	workDir, _ := os.Getwd()
	feDir := filepath.Join(workDir, "frontend")

	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(feDir, r.URL.Path)

		// Check if the file exists
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			// Fall back to index.html for client-side routing
			http.ServeFile(w, r, filepath.Join(feDir, "index.html"))
			return
		}

		http.FileServer(http.Dir(feDir)).ServeHTTP(w, r)
	}))
}
