package registry

import (
	"gohole/config"
	"gohole/internal/controller/dns"
	"gohole/internal/controller/http"
	"gohole/internal/database"
	"gohole/internal/filter"
	"gohole/internal/query"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Registry struct {
	QueryRepository database.Repository
	QueryService    query.Service
	QueryRouter     *http.QueryRouter

	DNSHandler *dns.Handler
	DNSCache   *dns.Cache
}

func NewRegistry(
	blockedDomains []string,
	allowedDomains []string,
	filterStrategy filter.Strategy,
	conn driver.Conn,
	cfg *config.Config,
) *Registry {
	blockFilter := filter.NewFilter(filterStrategy, blockedDomains)
	allowFilter := filter.NewFilter(filterStrategy, allowedDomains)

	repo := database.NewRepository(conn)

	queryService := query.NewService(blockFilter, allowFilter, repo)

	dnsCache := dns.NewCache()

	return &Registry{
		QueryRepository: repo,
		QueryService:    queryService,
		QueryRouter:     http.NewQueryRouter(queryService),

		DNSCache:   dnsCache,
		DNSHandler: dns.NewHandler(queryService, dnsCache, &cfg.DNS),
	}
}
