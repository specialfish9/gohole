package dns

import (
	"gohole/internal/query"
	"log"
	"sync"

	"codeberg.org/miekg/dns"
)

func Start(wg *sync.WaitGroup, filter query.Filter, address string, upstream string) {
	defer wg.Done()

	// Create DNS server
	d := &handler{
		upstream:     upstream,
		queryService: query.NewService(filter),
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
