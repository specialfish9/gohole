package dns

import (
	"context"
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

	dns.HandleFunc(".", recoverMiddleware(d.handleRequest)) // "." = catch-all

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

func recoverMiddleware(next func(context.Context, dns.ResponseWriter, *dns.Msg)) func(context.Context, dns.ResponseWriter, *dns.Msg) {
	return func(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("PANIC!", "message", r)
			}
		}()

		next(ctx, w, r)
	}
}
