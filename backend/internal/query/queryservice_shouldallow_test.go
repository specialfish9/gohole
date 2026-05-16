package query_test

import (
	"errors"
	"testing"

	"gohole/internal/query"
)

func TestShouldAllow_AllowFilterMatches(t *testing.T) {
	svc, _, _, allowFilter := newService(t)
	// domain is on the allow-list → should be allowed regardless of block filter
	allowFilter.EXPECT().Filter("example.com").Return(true, nil)

	ok, err := svc.ShouldAllow("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected allowed=true")
	}
}

func TestShouldAllow_BlockFilterMatches(t *testing.T) {
	svc, _, blockFilter, allowFilter := newService(t)
	allowFilter.EXPECT().Filter("bad.com").Return(false, nil)
	blockFilter.EXPECT().Filter("bad.com").Return(true, nil)

	ok, err := svc.ShouldAllow("bad.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected allowed=false for blocked domain")
	}
}

func TestShouldAllow_NeitherFilter(t *testing.T) {
	svc, _, blockFilter, allowFilter := newService(t)
	allowFilter.EXPECT().Filter("neutral.com").Return(false, nil)
	blockFilter.EXPECT().Filter("neutral.com").Return(false, nil)

	ok, err := svc.ShouldAllow("neutral.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected allowed=true for domain in neither filter")
	}
}

func TestShouldAllow_StripTrailingDot(t *testing.T) {
	svc, _, blockFilter, allowFilter := newService(t)
	// The service must strip the trailing dot before querying filters
	allowFilter.EXPECT().Filter("example.com").Return(false, nil)
	blockFilter.EXPECT().Filter("example.com").Return(false, nil)

	ok, err := svc.ShouldAllow("example.com.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected allowed=true")
	}
}

func TestShouldAllow_AllowFilterError(t *testing.T) {
	svc, _, _, allowFilter := newService(t)
	allowFilter.EXPECT().Filter("example.com").Return(false, errors.New("allow filter error"))

	_, err := svc.ShouldAllow("example.com")
	if err == nil {
		t.Error("expected error from allow filter")
	}
}

func TestShouldAllow_BlockFilterError(t *testing.T) {
	svc, _, blockFilter, allowFilter := newService(t)
	allowFilter.EXPECT().Filter("example.com").Return(false, nil)
	blockFilter.EXPECT().Filter("example.com").Return(false, errors.New("block filter error"))

	_, err := svc.ShouldAllow("example.com")
	if err == nil {
		t.Error("expected error from block filter")
	}
}

// ---- GetBlockListStats ----

func TestGetBlockListStats(t *testing.T) {
	svc, _, blockFilter, _ := newService(t)
	blockFilter.EXPECT().Size().Return(42)

	stats, err := svc.GetBlockListStats()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalEntries != 42 {
		t.Errorf("expected 42 entries, got %d", stats.TotalEntries)
	}
}

// ---- Interval/Granularity helpers (no mocks needed) ----

func TestInterval_ToDuration(t *testing.T) {
	cases := []struct {
		interval query.Interval
		wantSec  float64
	}{
		{query.Interval1H, 3600},
		{query.Interval6H, 21600},
		{query.Interval1D, 86400},
		{query.Interval7D, 604800},
		{query.Interval30D, 2592000},
	}
	for _, tc := range cases {
		got := tc.interval.ToDuration().Seconds()
		if got != tc.wantSec {
			t.Errorf("interval %q: expected %v, got %v", tc.interval, tc.wantSec, got)
		}
	}
}

func TestInterval_IsValid(t *testing.T) {
	valid := []query.Interval{query.Interval1H, query.Interval6H, query.Interval1D, query.Interval7D, query.Interval30D}
	for _, i := range valid {
		if !i.IsValid() {
			t.Errorf("expected %q to be valid", i)
		}
	}
	if query.Interval("bad").IsValid() {
		t.Error("expected \"bad\" interval to be invalid")
	}
}

func TestGranularity_IsValid(t *testing.T) {
	valid := []query.Granularity{
		query.Granularity1M, query.Granularity5M, query.Granularity15M,
		query.Granularity1H, query.Granularity6H, query.Granularity1D,
	}
	for _, g := range valid {
		if !g.IsValid() {
			t.Errorf("expected %q to be valid", g)
		}
	}
	if query.Granularity("bad").IsValid() {
		t.Error("expected \"bad\" granularity to be invalid")
	}
}
