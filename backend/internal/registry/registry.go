package registry

import (
	"fmt"
	"gohole/config"
	"gohole/internal/controller/dns"
	"gohole/internal/controller/http"
	"gohole/internal/database"
	"gohole/internal/filter"
	"gohole/internal/query"

	dns2 "codeberg.org/miekg/dns"
)

type Registry struct {
	QueryRepository database.Repository
	QueryService    query.Service
	QueryRouter     *http.QueryRouter

	UDPDNSHandler *dns.Handler
	TCPDNSHandler *dns.Handler
	DNSCache      dns.Cache
}

func NewRegistry(
	blockedDomains []string,
	allowedDomains []string,
	filterStrategy filter.Strategy,
	db database.Manager,
	cfg *config.Config,
) (*Registry, error) {
	blockFilter := filter.NewFilter(filterStrategy, blockedDomains)
	allowFilter := filter.NewFilter(filterStrategy, allowedDomains)

	repo := db.Repository()

	queryService := query.NewService(blockFilter, allowFilter, repo)

	dnsCache := dns.NewCache()
	dnsClient := &dns2.Client{}

	tcpHandler, err := dns.NewHandler(queryService, dns.TCP, dnsCache, &cfg.DNS, dnsClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create TCP DNS handler: %w", err)
	}

	udpHandler, err := dns.NewHandler(queryService, dns.UDP, dnsCache, &cfg.DNS, dnsClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP DNS handler: %w", err)
	}

	return &Registry{
		QueryRepository: repo,
		QueryService:    queryService,
		QueryRouter:     http.NewQueryRouter(queryService),

		DNSCache:      dnsCache,
		TCPDNSHandler: tcpHandler,
		UDPDNSHandler: udpHandler,
	}, nil
}
