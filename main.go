package main

import (
	"bufio"
	"context"
	"log"
	"os"

	"codeberg.org/miekg/dns"
)

const (
	upstream = "1.1.1.1:53"
	address  = ":53"
)

type dnsHandler struct {
	filter Filter
}

// handleRequest forwards DNS queries to the upstream server
func (d *dnsHandler) handleRequest(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) {
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

	resp, _, err := c.Exchange(ctx, r, "udp", upstream)
	if err != nil {
		log.Printf("ERROR failed to query upstream: %v", err)
		return
	}

	_, err = resp.WriteTo(w)
	if err != nil {
		log.Printf("ERROR failed to write response: %v", err)
	}
}

func mustParseBlockList(fileName string) []string {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return lines
}

func main() {
	domains := mustParseBlockList("block.txt")

	// Create DNS server
	d := &dnsHandler{
		filter: Trie(domains),
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
