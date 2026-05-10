package dns

import (
	"context"
	"fmt"
	"gohole/internal/query"
	"log/slog"
	"net"
	"net/netip"

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
		return nil, fmt.Errorf("dns handler: invalid upstream address: %v", err)
	}

	customDomains := make(map[string]netip.Addr)
	if cfg.CustomDomains.Ok {
		slog.Debug("Parsing custom domains", "count", len(cfg.CustomDomains.Value))
		var err error
		customDomains, err = parseCustomDomains(cfg.CustomDomains.Value)
		if err != nil {
			return nil, fmt.Errorf("dns handler: parsing custom domains: %w", err)
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
func (h *Handler) handleRequest(rc *ReqCtx, w dns.ResponseWriter, r *dns.Msg) {
	rc.Logger.Debug("Handling DNS request", "from", rc.Host)

	// We only answer the first question
	if len(r.Question) > 1 {
		rc.Logger.Warn("reqeust has more than one question", "questions", len(r.Question))
	}

	question := r.Question[0]
	var response *dns.Msg

	// Extract the requested name
	rc.Name = normalizeName(question.Header().Name)

	allow, answer, err := h.tryAnswerQuestion(rc, question)
	if err != nil {
		// In case of error, return an error response
		rc.Error = fmt.Errorf("dns handler: error trying answer question: %w", err)
		response = blockedResponse(r)
	} else if answer != nil {
		response = responseFromAnswer(answer, r)
	} else if !allow {
		// Else, if the domain is blocked, then return a refused response
		response = blockedResponse(r)
	} else {
		// Else, if the domain is allowed, forward the request to the upstream
		response, err = h.forwardRequest(rc, r)
		if err != nil {
			// In case of error, return an error response
			rc.Error = fmt.Errorf("dns handler: error forwarding request to upstream: %w", err)
			response = blockedResponse(r)
		}
	}

	if _, err := response.WriteTo(w); err != nil {
		rc.Error = fmt.Errorf("dns handler: error writing response to client: %w", err)
	} else {
		rc.Logger.Debug("Sent response to client", "from", rc.Host)
	}
}

func (h *Handler) tryAnswerQuestion(rc *ReqCtx, q dns.RR) (bool, dns.RR, error) {
	rc.Logger.Debug("Answering question", "name", rc.Name)

	// First, check custom domains
	resp, err := h.checkCustomDomains(rc, q)
	if err != nil {
		return false, nil, err
	}
	if resp != nil {
		// Custom domains are always allowed by definition
		return true, resp, nil
	}

	// Second, check cache
	if h.cacheEnabled {
		allowed, resp, err := h.checkCache(rc, q)
		if err != nil {
			return false, nil, err
		}
		if resp != nil {
			return allowed, resp, nil
		}
	}

	// Third, check filter
	allowed, err := h.checkFilter(rc, q)
	if err != nil {
		return false, nil, err
	}

	return allowed, nil, nil
}

func (h *Handler) checkCache(rc *ReqCtx, q dns.RR) (bool, dns.RR, error) {
	key := NewCacheKey(q)
	rc.Logger.Debug("Performing cache lookup", "key", key)
	allow, answer, cached := h.cache.Get(key)
	if !cached {
		rc.Logger.Debug("Cache miss", "key", key)
		return false, nil, nil
	}

	rc.Logger.Debug("Cache hit", "key", key)

	rc.Cached = true

	return allow, answer, nil
}

func (h *Handler) checkFilter(rc *ReqCtx, q dns.RR) (bool, error) {
	rc.Logger.Debug("Checking filter", "name", rc.Name)
	allow, err := h.queryService.ShouldAllow(rc.Name)
	if err != nil {
		return false, fmt.Errorf("filtering query: %w", err)
	}

	rc.Logger.Debug("Filter result", "name", rc.Name, "allow", allow)

	if !allow {
		// Update the cache
		rc.Logger.Debug("Updating cache with new blocked entry", "name", rc.Name)
		cacheKey := NewCacheKey(q)
		h.cache.SetBlocked(cacheKey)
	}

	rc.Allowed = allow

	return allow, nil
}

// forward forwards the query to the `upstream` server and returns the response.
func (h *Handler) forward(ctx context.Context, r *dns.Msg) (*dns.Msg, error) {
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

// forwardRequest forwards the request to the upstream and updates the handler cache.
func (h *Handler) forwardRequest(rc *ReqCtx, r *dns.Msg) (*dns.Msg, error) {
	rc.Logger.Debug("Forwarding request to upstream", "name", rc.Name, "upstream", h.upstream)
	response, err := h.forward(rc.Context, r)
	if err != nil {
		return nil, err
	}

	// Update the cache (only if there is something to cache)
	if len(response.Answer) > 0 {
		answer := response.Answer[0]
		ttl := answer.Header().TTL
		cacheKey := NewCacheKey(answer)
		rc.Logger.Debug("Updating cache", "key", cacheKey, "TTL", ttl)
		h.cache.Set(cacheKey, answer, ttl)
	} else {
		rc.Logger.Debug("No answer to cache", "name", rc.Name)
	}

	return responseFromAnswer(response.Answer[0], r), nil
}

func responseFromAnswer(a dns.RR, req *dns.Msg) *dns.Msg {
	resp := new(dns.Msg)
	dnsutil.SetReply(resp, req)
	resp.Answer = []dns.RR{a}
	return resp
}

func blockedResponse(req *dns.Msg) *dns.Msg {
	resp := new(dns.Msg)
	dnsutil.SetReply(resp, req)
	resp.Rcode = dns.RcodeNameError
	return resp
}
