package dns

import (
	"codeberg.org/miekg/dns"
	"context"
	"log"
)

type handler struct {
	upstream string
	filter   Filter
}

// handleRequest forwards DNS queries to the upstream server
func (d *handler) handleRequest(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) {
	c := new(dns.Client)

	// Filter queries
	question := r.Question[0]
	name := question.Header().Name

	filter, err := d.filter.Filter(name)
	if err != nil || !filter {
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

		return
	}

	log.Printf("IFNO PASS %s", name)

	resp, _, err := c.Exchange(ctx, r, "udp", d.upstream)
	if err != nil {
		log.Printf("ERROR failed to query upstream: %v", err)
		return
	}

	_, err = resp.WriteTo(w)
	if err != nil {
		log.Printf("ERROR failed to write response: %v", err)
	}
}

func Start(filter Filter, address string, upstream string) {
	// Create DNS server
	d := &handler{
		filter:   filter,
		upstream: upstream,
	}

	dns.HandleFunc(".", d.handleRequest) // "." = catch-all

	server := &dns.Server{
		Addr: address,
		Net:  "udp",
	}

	log.Printf("Starting DNS proxy on %s, forwarding to %s\n", address, upstream)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n", err.Error())
	}
}
