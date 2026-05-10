package main

import (
	"fmt"
	"gohole/config"
	"gohole/internal/controller/dns"
	"gohole/internal/controller/http"
	"gohole/internal/database"
	"gohole/internal/filter"
	"gohole/internal/query"
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

	tcpHandler, err := dns.NewHandler(queryService, dns.TCP, dnsCache, &cfg.DNS)
	if err != nil {
		return nil, fmt.Errorf("failed to create TCP DNS handler: %w", err)
	}

	udpHandler, err := dns.NewHandler(queryService, dns.UDP, dnsCache, &cfg.DNS)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP DNS handler: %w", err)
	}

	queryRouter := http.NewQueryRouter(queryService)

	daemons := []Daemon{
		http.NewServer(&cfg.HTTP, queryRouter),
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
