package dns

import (
	"gohole/internal/registry"
	"log/slog"
	"os"
	"sync"

	"codeberg.org/miekg/dns"
)

func Start(wg *sync.WaitGroup, reg *registry.Registry, address string, upstream string) {
	defer wg.Done()

	// Create DNS server
	d := &handler{
		upstream:     upstream,
		queryService: reg.QueryService,
	}

	dns.HandleFunc(".", d.handleRequest) // "." = catch-all

	server := &dns.Server{
		Addr: address,
		Net:  "udp",
	}

	slog.Info("Started DNS server", "address", address, "upstream", upstream)
	err := server.ListenAndServe()
	if err != nil {
		slog.Error("Failed to start server", "error", err.Error())
		os.Exit(1)
	}
}
