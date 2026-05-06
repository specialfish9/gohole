package dns

import (
	"context"
	"fmt"
	"gohole/internal/database"
	"gohole/internal/query"
	"log/slog"
	"net"
	"net/netip"
	"strings"
	"time"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/google/uuid"
)

type Handler struct {
	upstream      string
	cacheEnabled  bool
	queryService  query.Service
	protocol      Protocol
	customDomains map[string]netip.Addr
	cache         *Cache
}

func NewHandler(queryService query.Service, protocol Protocol, cache *Cache, cfg *Config) (*Handler, error) {
	upstream, err := addDefaultPort(cfg.Upstream)
	if err != nil {
		panic(fmt.Sprintf("invalid upstream address: %v", err))
	}

	customDomains := make(map[string]netip.Addr)
	if cfg.CustomDomains.Ok {
		slog.Debug("Parsing custom domains", "count", len(cfg.CustomDomains.Value))
		var err error
		customDomains, err = parseCustomDomains(cfg.CustomDomains.Value)
		if err != nil {
			return nil, fmt.Errorf("parsing custom domains: %w", err)
		}
	}

	return &Handler{
		upstream:      upstream,
		cacheEnabled:  cfg.CacheEnabled.Or(false),
		queryService:  queryService,
		cache:         cache,
		protocol:      protocol,
		customDomains: customDomains,
	}, nil
}

// handleRequest forwards DNS queries to the upstream server
func (h *Handler) handleRequest(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) {
	// Generate a trace id for this request
	trace := freshTraceID()

	l := slog.With("protcol", h.protocol, "ID", r.ID, "trace", trace)
	startTime := time.Now()

	// TODO handle multiple questions!
	question := r.Question[0]
	// Extract the requested name
	name := question.Header().Name
	// And the client host address
	host := strings.Split(w.RemoteAddr().String(), ":")[0]

	l.Debug("Received request", "name", name, "from", host, "questions", len(r.Question))

	var response *dns.Msg
	var allow bool
	var err error
	var cached bool

	// Check if the requested name is in the custom domains list
	ip, local := h.customDomains[name]
	if local {
		l.Debug("Name is in custom domains list, generating response", "name", name, "ip", ip)

		response, err = h.customDomainResponse(w, r, ip)
		if err != nil {
			l.Error("Error generating custom domain response", "name", name, "error", err.Error())

			// Nothing to do here, we will just return an empty response
			response = new(dns.Msg)
			dnsutil.SetReply(response, r)
		}
		allow = true
	} else {
		response, allow, cached, err = h.handleRemote(ctx, name, r)
		if err != nil {
			l.Error("Error handling remote request", "name", name, "error", err.Error())

			// In case of errors when handling the remote request, we still want to answer a client
			response = new(dns.Msg)
			dnsutil.SetReply(response, r)
			allow = true
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
		l.Info("PASS", "name", name, "host", host, "durationMS", millis, "cached", cached, "local", local)
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

func (h *Handler) handleRemote(ctx context.Context, name string, r *dns.Msg) (*dns.Msg, bool, bool, error) {
	l := slog.With("ID", r.ID)

	response := new(dns.Msg)
	dnsutil.SetReply(response, r)

	var allow bool
	var err error

	// First, check the cache
	var cached bool // cached false by default
	cacheKey := NewCacheKey(r)

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
		allow, err = h.queryService.ShouldAllow(name)
		if err != nil {
			l.Error("Filtering", "name", name, "error", err.Error())

			// In case of errors when filtering, we still want to answer a client
			allow = true
		}

		l.Debug("Filtering result", "allow", allow)

		// Build the response
		if allow {
			l.Debug("Forwarding response", "upstream", h.upstream)
			response, err = h.forwardResp(ctx, r, l)
			if err != nil {
				// Nothing to do here
				return nil, allow, cached, fmt.Errorf("forwarding response: %w", err)
			}
			l.Debug("Received response from upstream", "upstream", h.upstream, "answers", len(response.Answer), "truncated", response.Truncated)

			// Update the cache (only if there is something to cache)
			if len(response.Answer) > 0 {
				ttl := uint32(response.Answer[0].Header().TTL)
				l.Debug("Updating cache", "TTL", ttl)
				h.cache.Set(cacheKey, response, ttl)
			} else {
				l.Debug("No answer to cache")
			}

		} else {
			l.Debug("Setting response Rcode to refused")
			response.Rcode = dns.RcodeRefused
			l.Debug("Updating cache")
			h.cache.SetBlocked(cacheKey, response)
		}
	}

	return response, allow, cached, nil
}

// forwardResp forwards the query to the `upstream` server and returns the response.
func (h *Handler) forwardResp(ctx context.Context, r *dns.Msg, l *slog.Logger) (*dns.Msg, error) {
	c := new(dns.Client)

	resp, _, err := c.Exchange(ctx, r.Copy(), h.protocol, h.upstream)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange with upstream over %s: %w", h.protocol, err)
	}

	resp.ID = r.ID

	return resp, nil
}

func addDefaultPort(addr string) (string, error) {
	_, _, err := net.SplitHostPort(addr)
	if err == nil {
		// port already present
		return addr, nil
	}

	// Check it's actually a missing-port error and not a malformed address
	ip := net.ParseIP(addr)
	if ip == nil {
		return "", fmt.Errorf("invalid address: %s", addr)
	}

	return net.JoinHostPort(addr, "53"), nil
}

func freshTraceID() string {
	trace, err := uuid.NewV7()
	if err != nil {
		// In case of error generating the trace id, we log it and continue
		slog.Error("Failed to generate trace ID", "error", err.Error())
		trace = uuid.New()
	}

	return trace.String()
}
