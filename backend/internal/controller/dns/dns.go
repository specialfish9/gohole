package dns

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"codeberg.org/miekg/dns"
)

// Start creates the DNS server instance and runs it
func Start(wg *sync.WaitGroup, cfg *Config, handler *Handler) {
	defer wg.Done()

	dns.HandleFunc(".", recoverMiddleware(handler.handleRequest)) // "." = catch-all

	server := &dns.Server{
		Addr: cfg.Address,
		Net:  "udp",
	}

	slog.Info("Started DNS server", "address", cfg.Address, "upstream", cfg.Upstream, "cache", cfg.CacheEnabled.Or(false))
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
