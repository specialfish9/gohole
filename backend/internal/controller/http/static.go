package http

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

func serveStatic(r *chi.Mux) {
	workDir, _ := os.Getwd()
	feDir := http.Dir(filepath.Join(workDir, "frontend"))

	fs := http.StripPrefix("/", http.FileServer(feDir))
	r.Handle("/*", fs)
}
