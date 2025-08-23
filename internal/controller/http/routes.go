package http

import (
	"encoding/json"
	"gohole/internal/query"
	"net/http"
)

type queryRouter struct {
	queryService query.Service
}

func (qr *queryRouter) getAll(w http.ResponseWriter, r *http.Request) error {
	query, err := qr.queryService.GetAll()
	if err != nil {
		return err
	}

	b, err := json.Marshal(&query)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)

	return nil
}
