package dns

import (
	"context"
	"fmt"
	"log/slog"

	"codeberg.org/miekg/dns"
)

type Server struct {
	srv dns.Server
	l   *slog.Logger
	cfg *Config
}

func NewServer(cfg *Config, handler *Handler) *Server {
	mux := dns.NewServeMux()
	mux.HandleFunc(".", recoverMiddleware(handler.handleRequest)) // "." = catch-all

	return &Server{
		srv: dns.Server{
			Addr:    cfg.Address,
			Net:     "udp",
			Handler: mux,
		},
		l:   slog.With("component", "dnssrv"),
		cfg: cfg,
	}
}

func (s *Server) ID() string {
	return "DNS-server"
}

func (s *Server) Start() error {
	s.l.Info("Started DNS server", "address", s.cfg.Address, "upstream", s.cfg.Upstream, "cache", s.cfg.CacheEnabled.Or(false))
	if err := s.srv.ListenAndServe(); err != nil {
		return fmt.Errorf("dns: starting server: %w", err)
	}

	return nil
}

func (s *Server) Stop() error {
	s.l.Info("Stopping DNS server", "address", s.srv.Addr)
	s.srv.Shutdown(context.Background())
	return nil
}

func recoverMiddleware(next func(context.Context, dns.ResponseWriter, *dns.Msg)) func(context.Context, dns.ResponseWriter, *dns.Msg) {
	return func(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("PANIC!", "message", err)
			}
		}()

		next(ctx, w, r)
	}
}
