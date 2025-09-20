package http

import (
	"encoding/json"
	"fmt"
	"gohole/internal/query"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type queryRouter struct {
	queryService query.Service
}

func (qr *queryRouter) getAll(w http.ResponseWriter, r *http.Request) error {
	queries, err := qr.queryService.GetAll(r.Context(), 100)
	if err != nil {
		return err
	}

	jsonQueries := make([]query.Query, len(queries))
	for i, q := range queries {
		jsonQueries[i] = query.QueryFromDB(q)
	}

	b, err := json.Marshal(&jsonQueries)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)

	return nil
}

func (qr *queryRouter) getStats(w http.ResponseWriter, r *http.Request) error {
	interval := query.Interval(chi.URLParam(r, "interval"))
	if !interval.IsValid() {
		return fmt.Errorf("invalid interval value %s", interval)
	}

	stats, err := qr.queryService.GetStats(r.Context(), interval)
	if err != nil {
		return err
	}

	b, err := json.Marshal(&stats)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)

	return nil
}

func (qr *queryRouter) getStatsHistory(w http.ResponseWriter, r *http.Request) error {
	interval := query.Interval(r.URL.Query().Get("interval"))
	granularity := query.Granularity(r.URL.Query().Get("granularity"))

	if !interval.IsValid() {
		return fmt.Errorf("invalid interval parameter value: '%s'", interval)
	} else if !granularity.IsValid() {
		return fmt.Errorf("invalid granularity parameter value: '%s'", granularity)
	}

	history, err := qr.queryService.GetHistory(r.Context(), interval, granularity)
	if err != nil {
		return err
	}

	b, err := json.Marshal(&history)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)

	return nil
}

func (qr *queryRouter) getBlockListStats(w http.ResponseWriter, r *http.Request) error {
	stats, err := qr.queryService.GetBlockListStats()
	if err != nil {
		return err
	}

	b, err := json.Marshal(&stats)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)

	return nil
}

func (qr *queryRouter) getHostStats(w http.ResponseWriter, r *http.Request) error {
	stats, err := qr.queryService.GetHostStats(r.Context())
	if err != nil {
		return err
	}

	b, err := json.Marshal(&stats)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)

	return nil
}
