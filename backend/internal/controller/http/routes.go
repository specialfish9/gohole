package http

import (
	"encoding/json"
	"gohole/internal/query"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Router interface {
	Register(r chi.Router)
}

type QueryRouter struct {
	queryService query.Service
}

func NewQueryRouter(queryService query.Service) *QueryRouter {
	return &QueryRouter{
		queryService: queryService,
	}
}

func (qr *QueryRouter) Register(r chi.Router) {
	r.Get("/api/queries", errorHandler(qr.getAll))
	r.Get("/api/queries/stats", errorHandler(qr.getStats))
	r.Get("/api/queries/stats/history", errorHandler(qr.getStatsHistory))
	r.Get("/api/hosts/stats", errorHandler(qr.getHostStats))
	r.Get("/api/domains/stats", errorHandler(qr.getDomainStats))
	r.Get("/api/domains/{name}", errorHandler(qr.getDomainDetails))
	r.Get("/api/blocklist/stats", errorHandler(qr.getBlockListStats))
}

func (qr *QueryRouter) getAll(w http.ResponseWriter, r *http.Request) error {
	name := r.URL.Query().Get("name")

	queries, err := qr.queryService.GetAll(r.Context(), 100, name)
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
	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}

func (qr *QueryRouter) getStats(w http.ResponseWriter, r *http.Request) error {
	interval := query.Interval(r.URL.Query().Get("interval"))
	if interval != "" && !interval.IsValid() {
		return newHTTPErr(http.StatusBadRequest, "invalid interval value '%s'", interval)
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
	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}

func (qr *QueryRouter) getStatsHistory(w http.ResponseWriter, r *http.Request) error {
	interval := query.Interval(r.URL.Query().Get("interval"))
	granularity := query.Granularity(r.URL.Query().Get("granularity"))

	if !interval.IsValid() {
		return newHTTPErr(http.StatusBadRequest, "invalid interval parameter value '%s'", interval)
	} else if !granularity.IsValid() {
		return newHTTPErr(http.StatusBadRequest, "invalid granularity parameter value: '%s'", granularity)
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
	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}

func (qr *QueryRouter) getBlockListStats(w http.ResponseWriter, _ *http.Request) error {
	stats, err := qr.queryService.GetBlockListStats()
	if err != nil {
		return err
	}

	b, err := json.Marshal(&stats)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}

func (qr *QueryRouter) getHostStats(w http.ResponseWriter, r *http.Request) error {
	interval := query.Interval(r.URL.Query().Get("interval"))
	if !interval.IsValid() {
		return newHTTPErr(http.StatusBadRequest, "invalid interval value \"%s\"", interval)
	}

	stats, err := qr.queryService.GetHostStats(r.Context(), interval)
	if err != nil {
		return err
	}

	b, err := json.Marshal(&stats)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}

func (qr *QueryRouter) getDomainStats(w http.ResponseWriter, r *http.Request) error {
	interval := query.Interval(r.URL.Query().Get("interval"))
	if !interval.IsValid() {
		return newHTTPErr(http.StatusBadRequest, "invalid interval value \"%s\"", interval)
	}

	stats, err := qr.queryService.GetDomainStats(r.Context(), interval)
	if err != nil {
		return err
	}

	b, err := json.Marshal(&stats)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}

func (qr *QueryRouter) getDomainDetails(w http.ResponseWriter, r *http.Request) error {
	interval := query.Interval(r.URL.Query().Get("interval"))
	granularity := query.Granularity(r.URL.Query().Get("granularity"))

	if !interval.IsValid() {
		return newHTTPErr(http.StatusBadRequest, "invalid interval parameter value '%s'", interval)
	} else if !granularity.IsValid() {
		return newHTTPErr(http.StatusBadRequest, "invalid granularity parameter value: '%s'", granularity)
	}

	if granularity != query.Granularity1M &&
		granularity != query.Granularity1D &&
		granularity != query.Granularity1H {
		return newHTTPErr(
			http.StatusBadRequest,
			"granularity parameter value must be one of 'minute', 'hour', or 'day'",
		)
	}

	name := chi.URLParam(r, "name")
	if name == "" {
		return newHTTPErr(http.StatusBadRequest, "missing 'name' parameter")
	}

	details, err := qr.queryService.GetDomainDetails(r.Context(), name, interval, granularity)
	if err != nil {
		return err
	}

	b, err := json.Marshal(&details)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}
