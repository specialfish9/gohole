package dns

import (
	"fmt"
	"log/slog"
	"net"
	"net/netip"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"codeberg.org/miekg/dns/rdata"
	"github.com/google/uuid"
)

func addDefaultPort(addr string) (string, error) {
	_, _, err := net.SplitHostPort(addr)
	if err == nil {
		// port already present
		return addr, nil
	}

	// Check it's actually a missing-port error and not a malformed address
	ip := net.ParseIP(addr)
	if ip == nil {
		return "", fmt.Errorf("invalid address: %s", addr)
	}

	return net.JoinHostPort(addr, "53"), nil
}

func freshTraceID() string {
	trace, err := uuid.NewV7()
	if err != nil {
		// In case of error generating the trace id, we log it and continue
		slog.Error("Failed to generate trace ID", "error", err.Error())
		trace = uuid.New()
	}

	return trace.String()
}

func responseFromAnswer(a []dns.RR, req *dns.Msg) *dns.Msg {
	resp := new(dns.Msg)
	dnsutil.SetReply(resp, req)
	resp.Answer = append(resp.Answer, a...)
	return resp
}

// blockedResponse creates a DNS response for a blocked query based on the specified blocking strategy.
func blockedResponse(req *dns.Msg, strategy BlockingStrategy) *dns.Msg {
	if strategy == BlockingStrategyIP {
		resp := new(dns.Msg)
		dnsutil.SetReply(resp, req)

		if len(req.Question) == 0 {
			resp.Rcode = dns.RcodeFormatError
			return resp
		}

		question := req.Question[0]

		var addr netip.Addr
		switch dns.RRToType(question) {
		case dns.TypeA:
			addr = netip.MustParseAddr("0.0.0.0")
		case dns.TypeAAAA:
			addr = netip.MustParseAddr("::")
		default:
			// For unsupported query types, we default to returning NXDOMAIN
			slog.Warn(
				"Unsupported query type for blocked response using 'ip' blocking strategy. Defaulting to NXDOMAIN",
				"query_type",
				dns.RRToType(question),
			)
			return blockedResponse(req, BlockingStrategyNXDOMAIN)
		}

		answer, err := answerFromQuestion(question, addr)
		if err != nil {
			// This should not happen since we are only handling A and AAAA types above. In case it does,
			// we log the error and return an NXDOMAIN response.
			slog.Error(
				"Failed to create answer for blocked response using 'ip' blocking strategy. Using NXDOMAIN instead",
				"error",
				err.Error(),
			)
			return blockedResponse(req, BlockingStrategyNXDOMAIN)
		}
		resp.Answer = append(resp.Answer, answer)

		return resp
	} else if strategy == BlockingStrategyNXDOMAIN {
		resp := new(dns.Msg)
		dnsutil.SetReply(resp, req)
		resp.Rcode = dns.RcodeNameError
		return resp
	}

	// If the strategy is unknown, we log a warning and default to NXDOMAIN
	slog.Warn("Unknown blocking strategy. Defaulting to NXDOMAIN", "strategy", strategy)
	return blockedResponse(req, BlockingStrategyNXDOMAIN)
}

func answerFromQuestion(question dns.RR, addr netip.Addr) (dns.RR, error) {
	switch dns.RRToType(question) {
	case dns.TypeA:
		if !addr.Is4() {
			return nil, fmt.Errorf(
				"question is A, but IP address '%s' is not an IPv4 address",
				addr.String(),
			)
		}
		return &dns.A{
			Hdr: dns.Header{
				Name:  question.Header().Name,
				Class: dns.ClassINET,
				TTL:   60,
			},
			A: rdata.A{Addr: addr},
		}, nil

	case dns.TypeAAAA:
		if !addr.Is6() {
			return nil, fmt.Errorf(
				"question is AAAA, but IP address '%s' is not an IPv6 address",
				addr.String(),
			)
		}

		rr := &dns.AAAA{
			Hdr: dns.Header{
				Name:  question.Header().Name,
				Class: dns.ClassINET,
				TTL:   60,
			},
			AAAA: rdata.AAAA{Addr: addr},
		}
		return rr, nil

	default:
		return nil, fmt.Errorf(
			"unsupported query type %d for question name '%s'",
			dns.RRToType(question),
			question.Header().Name,
		)
	}
}
