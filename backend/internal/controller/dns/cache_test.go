package dns_test

import (
	"net/netip"
	"testing"
	"time"

	gdns "codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/rdata"

	"gohole/internal/controller/dns"
)

func newARecord(name string, addr string) gdns.RR {
	return &gdns.A{
		Hdr: gdns.Header{
			Name:  name,
			Class: gdns.ClassINET,
		},
		A: rdata.A{Addr: netip.MustParseAddr(addr)},
	}
}

func TestNewCacheKey(t *testing.T) {
	rr := newARecord("example.com.", "1.2.3.4")
	key := dns.NewCacheKey(rr)

	if key.Name != "example.com." {
		t.Errorf("expected Name %q, got %q", "example.com.", key.Name)
	}
	if key.Type != gdns.TypeA {
		t.Errorf("expected Type %d, got %d", gdns.TypeA, key.Type)
	}
	if key.Class != gdns.ClassINET {
		t.Errorf("expected Class %d, got %d", gdns.ClassINET, key.Class)
	}
}

func TestCache_GetMiss(t *testing.T) {
	c := dns.NewCache()
	key := dns.CacheKey{Name: "example.com.", Type: gdns.TypeA, Class: gdns.ClassINET}

	allowed, rr, found := c.Get(key)
	if found {
		t.Error("expected cache miss, got hit")
	}
	if allowed {
		t.Error("expected allowed=false on miss")
	}
	if rr != nil {
		t.Error("expected nil RR on miss")
	}
}

func TestCache_SetAndGet(t *testing.T) {
	c := dns.NewCache()
	rr := newARecord("example.com.", "1.2.3.4")
	key := dns.CacheKey{Name: "example.com.", Type: gdns.TypeA, Class: gdns.ClassINET}

	c.Set(key, rr, 60)

	allowed, got, found := c.Get(key)
	if !found {
		t.Fatal("expected cache hit, got miss")
	}
	if !allowed {
		t.Error("expected allowed=true for Set entry")
	}
	if got != rr {
		t.Error("expected the same RR back from cache")
	}
}

func TestCache_SetBlocked(t *testing.T) {
	c := dns.NewCache()
	key := dns.CacheKey{Name: "blocked.com.", Type: gdns.TypeA, Class: gdns.ClassINET}

	c.SetBlocked(key)

	allowed, rr, found := c.Get(key)
	if !found {
		t.Fatal("expected cache hit for blocked entry")
	}
	if allowed {
		t.Error("expected allowed=false for blocked entry")
	}
	if rr != nil {
		t.Error("expected nil RR for blocked entry")
	}
}

func TestCache_Expiration(t *testing.T) {
	c := dns.NewCache()
	rr := newARecord("ttl.com.", "5.6.7.8")
	key := dns.CacheKey{Name: "ttl.com.", Type: gdns.TypeA, Class: gdns.ClassINET}

	// Set with TTL of 0 seconds — expires immediately
	c.Set(key, rr, 0)

	// Wait briefly to ensure expiration
	time.Sleep(10 * time.Millisecond)

	_, _, found := c.Get(key)
	if found {
		t.Error("expected expired entry to be a cache miss")
	}
}

func TestCache_BlockedEntryDoesNotExpire(t *testing.T) {
	c := dns.NewCache()
	key := dns.CacheKey{Name: "neverexpire.com.", Type: gdns.TypeA, Class: gdns.ClassINET}

	c.SetBlocked(key)

	// Even after some time, blocked entries should remain
	time.Sleep(10 * time.Millisecond)

	allowed, _, found := c.Get(key)
	if !found {
		t.Error("expected blocked entry to persist (not expire)")
	}
	if allowed {
		t.Error("expected allowed=false for blocked entry")
	}
}

func TestCache_OverwriteEntry(t *testing.T) {
	c := dns.NewCache()
	rr1 := newARecord("overwrite.com.", "1.1.1.1")
	rr2 := newARecord("overwrite.com.", "2.2.2.2")
	key := dns.CacheKey{Name: "overwrite.com.", Type: gdns.TypeA, Class: gdns.ClassINET}

	c.Set(key, rr1, 60)
	c.Set(key, rr2, 60)

	_, got, found := c.Get(key)
	if !found {
		t.Fatal("expected cache hit after overwrite")
	}
	if got != rr2 {
		t.Error("expected second RR after overwrite")
	}
}

func TestCache_SetBlockedOverwritesAllowed(t *testing.T) {
	c := dns.NewCache()
	rr := newARecord("overwrite.com.", "1.1.1.1")
	key := dns.CacheKey{Name: "overwrite.com.", Type: gdns.TypeA, Class: gdns.ClassINET}

	c.Set(key, rr, 60)
	c.SetBlocked(key)

	allowed, _, found := c.Get(key)
	if !found {
		t.Fatal("expected cache hit after SetBlocked overwrote Set")
	}
	if allowed {
		t.Error("expected allowed=false after SetBlocked")
	}
}
