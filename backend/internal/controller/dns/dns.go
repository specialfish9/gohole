package dns

import (
	"context"
	"fmt"
	"log/slog"

	"codeberg.org/miekg/dns"
)

type Protocol = string

const UDP Protocol = "udp"
const TCP Protocol = "tcp"

type middleware func(next handlerFunc) handlerFunc
type handlerFunc func(rc *ReqCtx, w dns.ResponseWriter, r *dns.Msg)

type Server struct {
	protocol Protocol
	srv      dns.Server
	l        *slog.Logger
	cfg      *Config
}

func NewServer(cfg *Config, handler *Handler) *Server {
	mux := dns.NewServeMux()
	mux.HandleFunc(
		".", // "." = catch-all
		applyMiddlewares(
			handler.handleRequest,
			recoverMiddleware,
			logMiddleware("proto", handler.protocol),
			timeMiddleware,
		))

	return &Server{
		srv: dns.Server{
			Addr:    cfg.Address,
			Net:     handler.protocol,
			Handler: mux,
		},
		protocol: handler.protocol,
		l:        slog.With("component", fmt.Sprintf("%ssrv", handler.protocol)),
		cfg:      cfg,
	}
}

func (s *Server) ID() string {
	return fmt.Sprintf("DNS-server (%s)", s.protocol)
}

func (s *Server) Start() error {
	s.l.Info(
		"Started DNS server",
		"address", s.cfg.Address,
		"protocol", s.protocol,
		"upstream", s.cfg.Upstream,
		"cache", s.cfg.CacheEnabled.Or(false),
	)
	if err := s.srv.ListenAndServe(); err != nil {
		return fmt.Errorf("dns: starting %s server: %w", s.protocol, err)
	}

	return nil
}

func (s *Server) Stop() error {
	s.l.Info("Stopping DNS server", "protocol", s.protocol, "address", s.srv.Addr)
	s.srv.Shutdown(context.Background())
	return nil
}
