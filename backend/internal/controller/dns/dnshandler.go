package dns

import (
	"context"
	"fmt"
	"gohole/internal/database"
	"gohole/internal/query"
	"log/slog"
	"strings"
	"time"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
)

type Handler struct {
	upstream     string
	cacheEnabled bool
	queryService query.Service
	cache        *Cache
}

func NewHandler(queryService query.Service, cache *Cache, cfg *Config) *Handler {
	return &Handler{
		upstream:     cfg.Upstream,
		cacheEnabled: cfg.CacheEnabled.Or(false),
		queryService: queryService,
		cache:        cache,
	}
}

// handleRequest forwards DNS queries to the upstream server
func (h *Handler) handleRequest(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) {
	l := slog.With("ID", r.ID)
	startTime := time.Now()

	question := r.Question[0]
	// Extract the requested name
	name := question.Header().Name
	// And the client host address
	host := strings.Split(w.RemoteAddr().String(), ":")[0]

	l.Debug("Recieved request", "name", name, "from", host, "start", startTime)

	response := new(dns.Msg)
	dnsutil.SetReply(response, r)

	var cacheKey CacheKey
	var allow bool
	// cached false by default
	var cached bool

	// First check the cache
	if h.cacheEnabled {
		var answer []dns.RR
		allow, answer, cached = h.cache.Get(cacheKey)
		l.Debug("Performed cache lookup", "hit", cached, "allow", allow)

		if cached {
			// Copy the ID of the request
			if allow {
				response.Answer = answer
			} else {
				response.Rcode = dns.RcodeRefused
			}
		}
	}

	if !cached {
		allow, err := h.queryService.ShouldAllow(name)
		if err != nil {
			l.Error("Filtering", "name", name, "error", err.Error())

			// In case of errors when filtering, we still want to answer a client
			allow = true
		}

		l.Debug("Filtering result", "allow", allow)

		// Build the response
		if allow {
			l.Debug("Forwarding response")
			response, err = h.forwardResp(ctx, r)
			if err != nil {
				l.Error("Error forwarding response", "name", name, "error", err.Error())
				// Nothing to do here
				return
			}

			// Update the cache (only if there is something to cache)
			if len(response.Answer) > 0 {
				ttl := uint32(response.Answer[0].Header().TTL)
				l.Debug("Updating cache", "TTL", ttl)
				h.cache.Set(cacheKey, response, ttl)
			}

		} else {
			l.Debug("Setting response Rcode to refused")
			response.Rcode = dns.RcodeRefused
			l.Debug("Updating cache")
			h.cache.SetBlocked(cacheKey, response)
		}
	}

	// Write the response to the client
	if _, err := response.WriteTo(w); err != nil {
		// In case of error sending the response, we log it and continue
		l.Error("Failed to write response to the client", "name", name, "host", host, "blocked", !allow)
	}

	l.Debug("Sent response to the client", "resID", response.ID)

	millis := time.Since(startTime).Milliseconds()

	if allow {
		l.Info("PASS", "name", name, "host", host, "durationMS", millis, "cached", cached)
	} else {
		l.Info("SMASH", "name", name, "host", host, "durationMS", millis, "cached", cached)
	}

	// save the query
	q := database.NewQuery(name, host, !allow, millis)
	if err := h.queryService.Save(ctx, q); err != nil {
		l.Error("Error saving blocked query", "name", name, "error", err.Error())
	}

	l.Debug("Saved record in the DB")
}

// forwardResp forwards the query to the `upstream` server and returns the response.
func (h *Handler) forwardResp(ctx context.Context, r *dns.Msg) (*dns.Msg, error) {
	c := new(dns.Client)

	resp, _, err := c.Exchange(ctx, r.Copy(), "udp", h.upstream)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange with upstream: %w", err)
	}

	resp.ID = r.ID

	return resp, nil
}
