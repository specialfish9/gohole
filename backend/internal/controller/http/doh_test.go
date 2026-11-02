package http

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"gohole/internal/controller/dns"
	mockdns "gohole/internal/mock/dns"
	mockquery "gohole/internal/mock/query"

	gdns "codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/rdata"
	"github.com/specialfish9/confuso/v2"
	"go.uber.org/mock/gomock"
)

func TestDohServer_Post(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	queryService := mockquery.NewMockService(ctrl)
	dnsCache := mockdns.NewMockCache(ctrl)
	dnsClient := mockdns.NewMockClient(ctrl)

	testCfg := &dns.Config{
		Upstream:     "8.8.8.8",
		CacheEnabled: confuso.Optional[bool]{Value: false, Ok: true},
	}

	dnsHandler, err := dns.NewHandler(queryService, dns.UDP, dnsCache, testCfg, dnsClient)
	if err != nil {
		t.Fatalf("failed to create DNS handler: %v", err)
	}

	domain := "example.com."
	queryService.EXPECT().ShouldAllow(domain).Return(true, nil)

	aRecord := &gdns.A{
		Hdr: gdns.Header{
			Name:  domain,
			Class: gdns.ClassINET,
			TTL:   300,
		},
		A: rdata.A{Addr: netip.MustParseAddr("93.184.216.34")},
	}
	upstreamResp := new(gdns.Msg)
	upstreamResp.Answer = []gdns.RR{aRecord}

	dnsClient.EXPECT().
		Exchange(gomock.Any(), gomock.Any(), dns.UDP, gomock.Any()).
		Return(upstreamResp, time.Duration(0), nil)

	queryService.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

	qr := NewDoHRouter(dnsHandler)

	reqMsg := gdns.NewMsg("example.com", gdns.TypeA)
	if err := reqMsg.Pack(); err != nil {
		t.Fatalf("failed to pack request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/dns-query", bytes.NewReader(reqMsg.Data))
	req.Header.Set("Content-Type", "application/dns-message")
	req.RemoteAddr = "127.0.0.1:40212"

	w := httptest.NewRecorder()
	qr.handleDoH(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/dns-message" {
		t.Errorf(
			"expected Content-Type application/dns-message, got %s",
			resp.Header.Get("Content-Type"),
		)
	}

	respMsg := new(gdns.Msg)
	respMsg.Data = w.Body.Bytes()
	if err := respMsg.Unpack(); err != nil {
		t.Fatalf("failed to unpack response: %v", err)
	}

	if len(respMsg.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(respMsg.Answer))
	}
	aRec, ok := respMsg.Answer[0].(*gdns.A)
	if !ok {
		t.Fatalf("expected A record answer, got %T", respMsg.Answer[0])
	}
	if aRec.A.String() != "93.184.216.34" {
		t.Errorf("expected IP 93.184.216.34, got %s", aRec.A.String())
	}
}

func TestDohServer_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	queryService := mockquery.NewMockService(ctrl)
	dnsCache := mockdns.NewMockCache(ctrl)
	dnsClient := mockdns.NewMockClient(ctrl)

	testCfg := &dns.Config{
		Upstream:     "8.8.8.8",
		CacheEnabled: confuso.Optional[bool]{Value: false, Ok: true},
	}

	dnsHandler, err := dns.NewHandler(queryService, dns.UDP, dnsCache, testCfg, dnsClient)
	if err != nil {
		t.Fatalf("failed to create DNS handler: %v", err)
	}

	domain := "example.com."
	queryService.EXPECT().ShouldAllow(domain).Return(true, nil)

	aRecord := &gdns.A{
		Hdr: gdns.Header{
			Name:  domain,
			Class: gdns.ClassINET,
			TTL:   300,
		},
		A: rdata.A{Addr: netip.MustParseAddr("93.184.216.34")},
	}
	upstreamResp := new(gdns.Msg)
	upstreamResp.Answer = []gdns.RR{aRecord}

	dnsClient.EXPECT().
		Exchange(gomock.Any(), gomock.Any(), dns.UDP, gomock.Any()).
		Return(upstreamResp, time.Duration(0), nil)

	queryService.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

	dohRouter := NewDoHRouter(dnsHandler)

	reqMsg := gdns.NewMsg("example.com", gdns.TypeA)
	if err := reqMsg.Pack(); err != nil {
		t.Fatalf("failed to pack request: %v", err)
	}

	dnsParam := base64.RawURLEncoding.EncodeToString(reqMsg.Data)
	req := httptest.NewRequest(http.MethodGet, "/dns-query?dns="+dnsParam, nil)
	req.RemoteAddr = "127.0.0.1:40212"

	w := httptest.NewRecorder()
	dohRouter.handleDoH(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	respMsg := new(gdns.Msg)
	respMsg.Data = w.Body.Bytes()
	if err := respMsg.Unpack(); err != nil {
		t.Fatalf("failed to unpack response: %v", err)
	}

	if len(respMsg.Answer) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(respMsg.Answer))
	}
}
