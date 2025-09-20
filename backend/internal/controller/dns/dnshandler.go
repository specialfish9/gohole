package dns

import (
	"context"
	"gohole/internal/database"
	"gohole/internal/query"
	"log/slog"

	"codeberg.org/miekg/dns"
)

type handler struct {
	upstream     string
	queryService query.Service
}

// handleRequest forwards DNS queries to the upstream server
func (d *handler) handleRequest(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) {
	c := new(dns.Client)

	// Filter queries
	question := r.Question[0]
	name := question.Header().Name
	host := w.RemoteAddr().String()

	allow, err := d.queryService.ShouldAllow(name)
	if err != nil || !allow {
		if err != nil {
			slog.Error("Filtering", "error", err.Error())
		}
		slog.Info("SMASH "+name, "host", host)
		m := new(dns.Msg)
		m.Rcode = dns.RcodeRefused
		m.ID = r.ID

		_, err := m.WriteTo(w)
		if err != nil {
			slog.Error("Failed to write refusal response", "error", err.Error())
		}

		// Save query as blocked
		q := database.NewQuery(name, question.Header().Class, host, true)
		if err := d.queryService.Save(ctx, q); err != nil {
			slog.Error("saving blocked query: " + err.Error())
		}

		return
	}

	slog.Info("PASS "+name, "host", host)

	resp, _, err := c.Exchange(ctx, r, "udp", d.upstream)
	if err != nil {
		slog.Error("failed to query upstream", "error", err)
		return
	}

	_, err = resp.WriteTo(w)
	if err != nil {
		slog.Error("failed to write response", "error", err)
	}

	// Save query
	q := database.NewQuery(name, question.Header().Class, host, false)
	if err := d.queryService.Save(ctx, q); err != nil {
		slog.Error("saving passed query", "error", err)
	}
}
