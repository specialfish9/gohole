package dns_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"gohole/internal/controller/dns"

	gdns "codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"codeberg.org/miekg/dns/rdata"
)

func TestDohClient_Exchange(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/dns-message" {
			t.Errorf(
				"expected Content-Type application/dns-message, got %s",
				r.Header.Get("Content-Type"),
			)
		}
		if r.Header.Get("Accept") != "application/dns-message" {
			t.Errorf("expected Accept application/dns-message, got %s", r.Header.Get("Accept"))
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		reqMsg := new(gdns.Msg)
		reqMsg.Data = body
		if err := reqMsg.Unpack(); err != nil {
			t.Fatalf("failed to unpack request: %v", err)
		}

		respMsg := new(gdns.Msg)
		dnsutil.SetReply(respMsg, reqMsg)

		aRecord := &gdns.A{
			Hdr: gdns.Header{
				Name:  "example.com.",
				Class: gdns.ClassINET,
				TTL:   300,
			},
			A: rdata.A{Addr: netip.MustParseAddr("93.184.216.34")},
		}
		respMsg.Answer = append(respMsg.Answer, aRecord)

		if err := respMsg.Pack(); err != nil {
			t.Fatalf("failed to pack response: %v", err)
		}

		w.Header().Set("Content-Type", "application/dns-message")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respMsg.Data)
	}))
	defer ts.Close()

	client := dns.NewDohClient()
	req := gdns.NewMsg("example.com", gdns.TypeA)

	resp, _, err := client.Exchange(context.Background(), req, "udp", ts.URL)
	if err != nil {
		t.Fatalf("Exchange failed: %v", err)
	}

	if len(resp.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(resp.Answer))
	}
	aRecord, ok := resp.Answer[0].(*gdns.A)
	if !ok {
		t.Fatalf("expected A record answer, got %T", resp.Answer[0])
	}
	if aRecord.A.String() != "93.184.216.34" {
		t.Errorf("expected IP 93.184.216.34, got %s", aRecord.A.String())
	}
}

func TestDohClient_Exchange_BadStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := dns.NewDohClient()
	req := gdns.NewMsg("example.com", gdns.TypeA)

	_, _, err := client.Exchange(context.Background(), req, "udp", ts.URL)
	if err == nil {
		t.Fatal("expected error on internal server error, got nil")
	}
}

func TestDohClient_Exchange_IdMismatch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		reqMsg := new(gdns.Msg)
		reqMsg.Data = body
		if err := reqMsg.Unpack(); err != nil {
			t.Fatalf("failed to unpack request: %v", err)
		}

		respMsg := new(gdns.Msg)
		dnsutil.SetReply(respMsg, reqMsg)
		respMsg.ID = reqMsg.ID + 1 // mismatch ID

		if err := respMsg.Pack(); err != nil {
			t.Fatalf("failed to pack response: %v", err)
		}

		w.Header().Set("Content-Type", "application/dns-message")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respMsg.Data)
	}))
	defer ts.Close()

	client := dns.NewDohClient()
	req := gdns.NewMsg("example.com", gdns.TypeA)

	_, _, err := client.Exchange(context.Background(), req, "udp", ts.URL)
	if err == nil {
		t.Fatal("expected error on message ID mismatch, got nil")
	}
}
