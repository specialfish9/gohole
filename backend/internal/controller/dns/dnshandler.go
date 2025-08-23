package dns

import (
	"context"
	"gohole/internal/database"
	"gohole/internal/query"
	"log"

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

	allow, err := d.queryService.ShouldAllow(name)
	if err != nil || !allow {
		if err != nil {
			log.Printf("ERROR Filtering: %v", err)
		}
		log.Printf("INFO SMASH %s", name)
		m := new(dns.Msg)
		m.Rcode = dns.RcodeRefused
		m.ID = r.ID

		_, err := m.WriteTo(w)
		if err != nil {
			log.Printf("ERROR Failed to write refusal response: %v", err)
		}

		// Save query as blocked
		q := database.NewQuery(name, question.Header().Class, true)
		if err := d.queryService.Save(ctx, q); err != nil {
			log.Printf("ERROR saving blocked query: %v\n", err)
		}

		return
	}

	log.Printf("INFO PASS %s", name)

	resp, _, err := c.Exchange(ctx, r, "udp", d.upstream)
	if err != nil {
		log.Printf("ERROR failed to query upstream: %v", err)
		return
	}

	_, err = resp.WriteTo(w)
	if err != nil {
		log.Printf("ERROR failed to write response: %v", err)
	}

	// Save query
	q := database.NewQuery(name, question.Header().Class, false)
	if err := d.queryService.Save(ctx, q); err != nil {
		log.Printf("ERROR saving passed query: %v\n", err)
	}
}
