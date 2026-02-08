package http

import (
	"errors"
	"fmt"
	"net/http"
)

type HTTPError struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

var _ error = (*HTTPError)(nil)

func (e *HTTPError) Error() string {
	return e.Message
}

func newHTTPErr(status int, format string, args ...any) *HTTPError {
	return &HTTPError{
		Message: fmt.Sprintf(format, args...),
		Status:  status,
	}
}

func isHTTPError(err error) (bool, *HTTPError) {
	var httpErr HTTPError
	if ok := errors.As(err, &httpErr); ok {
		return true, &httpErr
	}
	return false, nil
}

func errorHandler(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)

		if isHTTP, httpErr := isHTTPError(err); isHTTP {
			http.Error(w, httpErr.Message, httpErr.Status)
			return
		}

		// For any other error, return a generic 500 Internal Server Error
		if err != nil {
			http.Error(w, err.Error(), 500)
		}

	}
}
