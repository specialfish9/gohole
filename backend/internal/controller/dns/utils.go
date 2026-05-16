package dns

import (
	"fmt"
	"log/slog"
	"net"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
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

func blockedResponse(req *dns.Msg) *dns.Msg {
	resp := new(dns.Msg)
	dnsutil.SetReply(resp, req)
	resp.Rcode = dns.RcodeNameError
	return resp
}
