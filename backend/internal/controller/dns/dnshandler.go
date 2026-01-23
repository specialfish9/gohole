package dns

import (
	"context"
	"gohole/internal/database"
	"gohole/internal/query"
	"log/slog"
	"strings"
	"time"

	"codeberg.org/miekg/dns"
)

type handler struct {
	upstream     string
	queryService query.Service
}

// handleRequest forwards DNS queries to the upstream server
func (d *handler) handleRequest(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) {
	startTime := time.Now()
	c := new(dns.Client)

	// Filter queries
	question := r.Question[0]
	name := question.Header().Name
	host := strings.Split(w.RemoteAddr().String(), ":")[0]

	allow, err := d.queryService.ShouldAllow(name)
	if err != nil || !allow {
		if err != nil {
			slog.Error("Filtering", "name", name, "error", err.Error())
		}
		m := new(dns.Msg)
		m.Rcode = dns.RcodeRefused
		m.ID = r.ID

		_, err := m.WriteTo(w)
		if err != nil {
			slog.Error("Failed to write refusal response", "name", name, "error", err.Error())
		}

		millis := time.Since(startTime).Milliseconds()
		slog.Info("SMASH", "name", name, "host", host, "durationMS", millis)

		// Save query as blocked
		q := database.NewQuery(name, question.Header().Class, host, true, millis)
		if err := d.queryService.Save(ctx, q); err != nil {
			slog.Error("Error saving blocked query", "name", name, "error", err.Error())
		}

		return
	}

	resp, _, err := c.Exchange(ctx, r, "udp", d.upstream)
	if err != nil {
		slog.Error("Failed to query upstream", "name", name, "error", err)
		return
	}

	_, err = resp.WriteTo(w)
	if err != nil {
		slog.Error("Failed to write response", "name", name, "error", err)
	}

	millis := time.Since(startTime).Milliseconds()
	slog.Info("PASS", "name", name, "host", host, "durationMS", millis)

	// Save query
	q := database.NewQuery(name, question.Header().Class, host, false, millis)
	if err := d.queryService.Save(ctx, q); err != nil {
		slog.Error("Error when saving passed query", "name", name, "error", err)
	}
}
