package main

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

type Daemon interface {
	ID() string
	Start() error
	Stop() error
}

type DaemonRegistry struct {
	daemons []Daemon
	// This is needed because, when using Clickhouse, we need to
	// close the repository before shutdown
	repo database.Repository
}

func NewDaemonRegistry(
	blockedDomains []string,
	allowedDomains []string,
	filterStrategy filter.Strategy,
	db database.Manager,
	cfg *config.Config,
) (*DaemonRegistry, error) {
	blockFilter := filter.NewFilter(filterStrategy, blockedDomains)
	allowFilter := filter.NewFilter(filterStrategy, allowedDomains)

	repo := db.Repository()

	queryService := query.NewService(blockFilter, allowFilter, repo)

	dnsCache := dns.NewCache()

	var dnsClient dns.Client
	if dns.IsURL(cfg.DNS.Upstream) {
		dnsClient = dns.NewDohClient()
	} else {
		dnsClient = dns2.NewClient()
	}

	tcpHandler, err := dns.NewHandler(queryService, dns.TCP, dnsCache, &cfg.DNS, dnsClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create TCP DNS handler: %w", err)
	}

	udpHandler, err := dns.NewHandler(queryService, dns.UDP, dnsCache, &cfg.DNS, dnsClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP DNS handler: %w", err)
	}

	queryRouter := http.NewQueryRouter(queryService)
	var httpRouters = []http.Router{queryRouter}

	if cfg.DNS.DoHEnabled.Or(false) {
		httpRouters = append(httpRouters, http.NewDoHRouter(tcpHandler))
	}

	daemons := []Daemon{
		http.NewServer(&cfg.HTTP, httpRouters...),
		dns.NewServer(&cfg.DNS, tcpHandler),
		dns.NewServer(&cfg.DNS, udpHandler),
	}

	return &DaemonRegistry{
		daemons: daemons,
		repo:    repo,
	}, nil
}

func (r *DaemonRegistry) Start() {
	for _, d := range r.daemons {
		go func(d Daemon) {
			if err := d.Start(); err != nil {
				logPanic(fmt.Sprintf("Starting daemon %s: %v", d.ID(), err))
			}
		}(d)
	}
}

func (r *DaemonRegistry) Stop() {
	for _, d := range r.daemons {
		if err := d.Stop(); err != nil {
			logPanic(fmt.Sprintf("Stopping daemon %s: %v", d.ID(), err))
		}
	}
}
