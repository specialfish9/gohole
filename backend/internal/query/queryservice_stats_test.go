package query_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"gohole/internal/database"
	"gohole/internal/query"
)

// ---- GetStats ----

func TestGetStats_AllInterval(t *testing.T) {
	svc, repo, _, _ := newService(t)
	queries := []database.Query{
		{Blocked: true},
		{Blocked: false},
		{Blocked: false},
	}
	repo.EXPECT().FindAll(gomock.Any()).Return(queries, nil)

	stats, err := svc.GetStats(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalQueries != 3 {
		t.Errorf("expected 3 total, got %d", stats.TotalQueries)
	}
	if stats.BlockedQueries != 1 {
		t.Errorf("expected 1 blocked, got %d", stats.BlockedQueries)
	}
	if stats.AllowedQueries != 2 {
		t.Errorf("expected 2 allowed, got %d", stats.AllowedQueries)
	}
}

func TestGetStats_WithInterval(t *testing.T) {
	svc, repo, _, _ := newService(t)
	queries := []database.Query{
		{Blocked: true},
		{Blocked: true},
	}
	repo.EXPECT().FindAllByInterval(gomock.Any(), gomock.Any()).Return(queries, nil)

	stats, err := svc.GetStats(context.Background(), query.Interval1H)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.BlockedQueries != 2 {
		t.Errorf("expected 2 blocked, got %d", stats.BlockedQueries)
	}
	if stats.AllowedQueries != 0 {
		t.Errorf("expected 0 allowed, got %d", stats.AllowedQueries)
	}
}

func TestGetStats_RepoError(t *testing.T) {
	svc, repo, _, _ := newService(t)
	repo.EXPECT().FindAll(gomock.Any()).Return(nil, errors.New("db error"))

	_, err := svc.GetStats(context.Background(), "")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// ---- GetHistory ----

func TestGetHistory_EmptyQueries(t *testing.T) {
	svc, repo, _, _ := newService(t)
	repo.EXPECT().FindAllByInterval(gomock.Any(), gomock.Any()).Return(nil, nil)

	points, err := svc.GetHistory(context.Background(), query.Interval1H, query.Granularity5M)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1h / 5m = 12 steps
	if len(points) != 12 {
		t.Errorf("expected 12 history points, got %d", len(points))
	}
	for _, p := range points {
		if p.Blocked != 0 || p.Allowed != 0 {
			t.Error("expected all points to have zero counts for empty query set")
		}
	}
}

func TestGetHistory_WithQueries(t *testing.T) {
	svc, repo, _, _ := newService(t)

	// Place two queries 30 seconds into the 1h window (index 0 of 5m buckets)
	startTs := time.Now().UTC().Add(-query.Interval1H.ToDuration())
	q1 := database.Query{Timestamp: startTs.Unix() + 30, Blocked: true}
	q2 := database.Query{Timestamp: startTs.Unix() + 30, Blocked: false}

	repo.EXPECT().
		FindAllByInterval(gomock.Any(), gomock.Any()).
		Return([]database.Query{q1, q2}, nil)

	points, err := svc.GetHistory(context.Background(), query.Interval1H, query.Granularity5M)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if points[0].Blocked != 1 {
		t.Errorf("expected 1 blocked in bucket 0, got %d", points[0].Blocked)
	}
	if points[0].Allowed != 1 {
		t.Errorf("expected 1 allowed in bucket 0, got %d", points[0].Allowed)
	}
}

func TestGetHistory_RepoError(t *testing.T) {
	svc, repo, _, _ := newService(t)
	repo.EXPECT().FindAllByInterval(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

	_, err := svc.GetHistory(context.Background(), query.Interval1H, query.Granularity5M)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// ---- GetHostStats ----

func TestGetHostStats_OK(t *testing.T) {
	svc, repo, _, _ := newService(t)
	expected := []database.HostStat{{Host: "host1", QueryCount: 5}}
	repo.EXPECT().FindHostStats(gomock.Any(), gomock.Any()).Return(expected, nil)

	got, err := svc.GetHostStats(context.Background(), query.Interval1H)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Host != "host1" {
		t.Errorf("unexpected result: %v", got)
	}
}

func TestGetHostStats_Error(t *testing.T) {
	svc, repo, _, _ := newService(t)
	repo.EXPECT().FindHostStats(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

	_, err := svc.GetHostStats(context.Background(), query.Interval1H)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
