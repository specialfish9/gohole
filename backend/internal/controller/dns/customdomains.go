package dns

import (
	"fmt"
	"net/netip"
	"strings"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"codeberg.org/miekg/dns/rdata"
)

func parseCustomDomains(customDomains []string) (map[string]netip.Addr, error) {
	res := make(map[string]netip.Addr)
	for _, entry := range customDomains {
		parts := strings.Split(entry, " ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("expected '<name> <ip addr>', got '%s'", entry)
		}

		ip, err := netip.ParseAddr(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid IP address '%s' in entry '%s': %s", parts[1], entry, err)
		}

		name := parts[0]
		res[name] = ip
	}

	return res, nil
}

func (h *Handler) customDomainResponse(w dns.ResponseWriter, r *dns.Msg, ip netip.Addr) (*dns.Msg, error) {
	response := new(dns.Msg)
	dnsutil.SetReply(response, r)
	response.Authoritative = true

	// TODO handle multiple questions!
	question := r.Question[0]

	switch dns.RRToType(question) {
	case dns.TypeA:
		if !ip.Is4() {
			return nil, fmt.Errorf("question is A, but IP address '%s' is not an IPv4 address", ip)
		}
		rr := &dns.A{
			Hdr: dns.Header{
				Name:  question.Header().Name,
				Class: dns.ClassINET,
				TTL:   60,
			},
			A: rdata.A{Addr: ip},
		}
		response.Answer = append(response.Answer, rr)

	case dns.TypeAAAA:
		if !ip.Is6() {
			return nil, fmt.Errorf("question is AAAA, but IP address '%s' is not an IPv6 address", ip)
		}

		rr := &dns.AAAA{
			Hdr: dns.Header{
				Name:  question.Header().Name,
				Class: dns.ClassINET,
				TTL:   60,
			},
			AAAA: rdata.AAAA{Addr: ip},
		}
		response.Answer = append(response.Answer, rr)

	default:
		return nil, fmt.Errorf("unsupported query type %d for custom domain '%s'", dns.RRToType(question), question.Header().Name)
	}

	return response, nil
}
