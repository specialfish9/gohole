package dns_test

import (
	"context"
	"log/slog"
	"net/netip"
	"testing"
	"time"

	gdns "codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/rdata"
	"github.com/specialfish9/confuso/v2"
	"go.uber.org/mock/gomock"

	"gohole/internal/controller/dns"
	mockdns "gohole/internal/mock/dns"
	mockquery "gohole/internal/mock/query"
)

type tctx struct {
	ctrl         *gomock.Controller
	h            *dns.Handler
	cache        *mockdns.MockCache
	client       *mockdns.MockClient
	queryService *mockquery.MockService
}

func newCtx(t *testing.T, cfg *dns.Config) *tctx {
	ctrl := gomock.NewController(t)

	cache := mockdns.NewMockCache(ctrl)
	queryService := mockquery.NewMockService(ctrl)
	client := mockdns.NewMockClient(ctrl)

	h, err := dns.NewHandler(queryService, dns.UDP, cache, cfg, client)
	if err != nil {
		t.Fatal(err)
	}
	return &tctx{
		ctrl:         ctrl,
		cache:        cache,
		queryService: queryService,
		client:       client,
		h:            h,
	}

}

func newReqCtx() *dns.ReqCtx {
	return &dns.ReqCtx{
		Context: context.Background(),
		Logger:  slog.Default(),
	}
}

func TestHandleRequest(t *testing.T) {
	const domain = "example.com."

	t.Run("blocked", func(t *testing.T) {
		// Arrange
		var testCfg = &dns.Config{Upstream: "8.8.8.8"}
		tc := newCtx(t, testCfg)

		tc.queryService.EXPECT().ShouldAllow(domain).Return(false, nil)
		tc.cache.EXPECT().SetBlocked(gomock.Any())

		rc := newReqCtx()
		w := &fakeWriter{}
		r := gdns.NewMsg("example.com", gdns.TypeA)

		// Act
		tc.h.HandleRequest(rc, w, r)

		// Assert
		got, err := w.ParseMsg()
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if got.Rcode != gdns.RcodeNameError {
			t.Errorf("expected NXDOMAIN (Rcode %d), got Rcode %d", gdns.RcodeNameError, got.Rcode)
		}
		if len(got.Answer) != 0 {
			t.Errorf("expected no answers for blocked domain, got %d", len(got.Answer))
		}
	})

	t.Run("allow - forward, no cache", func(t *testing.T) {
		// Arrange
		var testCfg = &dns.Config{
			CacheEnabled: confuso.Optional[bool]{Value: false, Ok: true},
			Upstream:     "8.8.8.8:53",
		}
		tc := newCtx(t, testCfg)

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

		tc.queryService.EXPECT().ShouldAllow(domain).Return(true, nil)
		tc.client.EXPECT().
			Exchange(gomock.Any(), gomock.Any(), dns.UDP, testCfg.Upstream).
			Return(upstreamResp, time.Duration(0), nil)

		rc := newReqCtx()
		w := &fakeWriter{}
		r := gdns.NewMsg("example.com", gdns.TypeA)

		// Act
		tc.h.HandleRequest(rc, w, r)

		// Assert
		got, err := w.ParseMsg()
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if got.Rcode != gdns.RcodeSuccess {
			t.Errorf("expected NOERROR (Rcode 0), got Rcode %d", got.Rcode)
		}
		if len(got.Answer) == 0 {
			t.Fatal("expected at least one answer for allowed domain")
		}
	})

	t.Run("allow - cache miss", func(t *testing.T) {
		// Arrange
		var testCfg = &dns.Config{
			CacheEnabled: confuso.Optional[bool]{Value: true, Ok: true},
			Upstream:     "8.8.8.8:53",
		}
		tc := newCtx(t, testCfg)

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

		tc.queryService.EXPECT().ShouldAllow(domain).Return(true, nil)
		tc.client.EXPECT().
			Exchange(gomock.Any(), gomock.Any(), dns.UDP, testCfg.Upstream).
			Return(upstreamResp, time.Duration(0), nil)
		// Simulate cache miss
		tc.cache.EXPECT().Get(gomock.Any()).Return(false, nil, false)
		tc.cache.EXPECT().Set(gomock.Any(), []gdns.RR{aRecord}, uint32(300))

		rc := newReqCtx()
		w := &fakeWriter{}
		r := gdns.NewMsg("example.com", gdns.TypeA)

		// Act
		tc.h.HandleRequest(rc, w, r)

		// Assert
		got, err := w.ParseMsg()
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if got.Rcode != gdns.RcodeSuccess {
			t.Errorf("expected NOERROR (Rcode 0), got Rcode %d", got.Rcode)
		}
		if len(got.Answer) == 0 {
			t.Fatal("expected at least one answer for allowed domain")
		}
	})

	t.Run("custom domain - A record", func(t *testing.T) {
		// Arrange: no mock expectations — custom domains bypass filter, cache, and upstream.
		var testCfg = &dns.Config{
			Upstream: "8.8.8.8",
			CustomDomains: confuso.Optional[map[string]any]{
				Ok:    true,
				Value: map[string]any{"custom.local": "1.2.3.4"},
			},
		}
		tc := newCtx(t, testCfg)

		rc := newReqCtx()
		w := &fakeWriter{}
		r := gdns.NewMsg("custom.local", gdns.TypeA)

		// Act
		tc.h.HandleRequest(rc, w, r)

		// Assert
		got, err := w.ParseMsg()
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if got.Rcode != gdns.RcodeSuccess {
			t.Errorf("expected NOERROR (Rcode 0), got Rcode %d", got.Rcode)
		}
		if len(got.Answer) == 0 {
			t.Fatal("expected at least one answer for custom domain")
		}
		aRR, ok := got.Answer[0].(*gdns.A)
		if !ok {
			t.Fatalf("expected *dns.A record, got %T", got.Answer[0])
		}
		if aRR.A.Addr.String() != "1.2.3.4" {
			t.Errorf("expected IP 1.2.3.4, got %s", aRR.A.Addr)
		}
	})

	t.Run("custom domain - AAAA record", func(t *testing.T) {
		// Arrange: no mock expectations — custom domains bypass filter, cache, and upstream.
		var testCfg = &dns.Config{
			Upstream: "8.8.8.8",
			CustomDomains: confuso.Optional[map[string]any]{
				Ok:    true,
				Value: map[string]any{"custom.local": "2001:db8::1"},
			},
		}
		tc := newCtx(t, testCfg)

		rc := newReqCtx()
		w := &fakeWriter{}
		r := gdns.NewMsg("custom.local", gdns.TypeAAAA)

		// Act
		tc.h.HandleRequest(rc, w, r)

		// Assert
		got, err := w.ParseMsg()
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if got.Rcode != gdns.RcodeSuccess {
			t.Errorf("expected NOERROR (Rcode 0), got Rcode %d", got.Rcode)
		}
		if len(got.Answer) == 0 {
			t.Fatal("expected at least one answer for custom domain")
		}
		aaaaRR, ok := got.Answer[0].(*gdns.AAAA)
		if !ok {
			t.Fatalf("expected *dns.AAAA record, got %T", got.Answer[0])
		}
		if aaaaRR.AAAA.Addr.String() != "2001:db8::1" {
			t.Errorf("expected IP 2001:db8::1, got %s", aaaaRR.AAAA.Addr)
		}
	})

	t.Run("allow - cache hit", func(t *testing.T) {
		// Arrange
		var testCfg = &dns.Config{
			CacheEnabled: confuso.Optional[bool]{Value: true, Ok: true},
			Upstream:     "8.8.8.8:53",
		}
		tc := newCtx(t, testCfg)

		aRecord := &gdns.A{
			Hdr: gdns.Header{
				Name:  domain,
				Class: gdns.ClassINET,
				TTL:   300,
			},
			A: rdata.A{Addr: netip.MustParseAddr("93.184.216.34")},
		}

		// Simulate cache hit
		tc.cache.EXPECT().Get(gomock.Any()).Return(true, []gdns.RR{aRecord}, true)

		rc := newReqCtx()
		w := &fakeWriter{}
		r := gdns.NewMsg("example.com", gdns.TypeA)

		// Act
		tc.h.HandleRequest(rc, w, r)

		// Assert
		got, err := w.ParseMsg()
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if got.Rcode != gdns.RcodeSuccess {
			t.Errorf("expected NOERROR (Rcode 0), got Rcode %d", got.Rcode)
		}
		if len(got.Answer) == 0 {
			t.Fatal("expected at least one answer for allowed domain")
		}
	})
}
