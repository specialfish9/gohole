package http

import (
	"net/http"
)

func errorHandler(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err == nil {
			return
		}

		// TODO handle errors properly
		http.Error(w, err.Error(), 500)
	}
}
