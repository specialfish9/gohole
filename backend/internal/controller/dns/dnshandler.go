package dns

import (
	"fmt"
	"gohole/internal/database"
	"gohole/internal/query"
	"log/slog"
	"net/netip"

	"codeberg.org/miekg/dns"
)

type Handler struct {
	upstream      string
	cacheEnabled  bool
	queryService  query.Service
	protocol      Protocol
	customDomains map[string]netip.Addr
	cache         Cache
	client        Client
}

func NewHandler(
	queryService query.Service,
	protocol Protocol,
	cache Cache,
	cfg *Config,
	client Client,
) (*Handler, error) {
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
		client:        client,
	}, nil
}

// HandleRequest forwards DNS queries to the upstream server
func (h *Handler) HandleRequest(rc *ReqCtx, w dns.ResponseWriter, r *dns.Msg) {
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
	rc.Allowed = allow

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

// forwardRequest forwards the request to the upstream and updates the handler cache.
func (h *Handler) forwardRequest(rc *ReqCtx, r *dns.Msg) (*dns.Msg, error) {
	rc.Logger.Debug("Forwarding request to upstream", "name", rc.Name, "upstream", h.upstream)

	response, _, err := h.client.Exchange(rc.Context, r.Copy(), h.protocol, h.upstream)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange with upstream over %s: %w", h.protocol, err)
	}

	response.ID = r.ID

	// Update the cache (only if there is something to cache)
	if len(response.Answer) > 0 {
		answer := response.Answer[0]
		if h.cacheEnabled {
			ttl := answer.Header().TTL
			cacheKey := NewCacheKey(answer)
			rc.Logger.Debug("Updating cache", "key", cacheKey, "TTL", ttl)
			h.cache.Set(cacheKey, answer, ttl)
		}
	} else {
		rc.Logger.Debug("No answer to cache", "name", rc.Name)
	}

	return responseFromAnswer(response.Answer[0], r), nil
}

// persistenceMiddleware stores the query in the database after the request has been handled.
func (h *Handler) persistenceMiddleware(next handlerFunc) handlerFunc {
	return func(rc *ReqCtx, w dns.ResponseWriter, r *dns.Msg) {
		next(rc, w, r)
		q := database.NewQuery(
			rc.Name,
			rc.Host,
			!rc.Allowed,
			rc.End.Sub(rc.Start).Milliseconds(),
		)
		err := h.queryService.Save(rc.Context, q)
		if err != nil {
			rc.Logger.Error("Failed to save query to database", "error", err.Error())
		}
		rc.Logger.Debug("Persisting query", "name", q.Name)
	}
}
