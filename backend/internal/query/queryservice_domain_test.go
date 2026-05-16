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

// ---- GetDomainStats ----

func TestGetDomainStats_OK(t *testing.T) {
	svc, repo, _, _ := newService(t)

	repo.EXPECT().FindDomainStats(gomock.Any(), gomock.Any()).Return(database.DomainStats{Total: 100, BlockedCount: 40}, nil)
	repo.EXPECT().FindTopDomains(gomock.Any(), true, gomock.Any(), 10).Return([]database.TopDomain{{Domain: "bad.com", Count: 20}}, nil)
	repo.EXPECT().FindTopDomains(gomock.Any(), false, gomock.Any(), 10).Return([]database.TopDomain{{Domain: "ok.com", Count: 80}}, nil)

	stats, err := svc.GetDomainStats(context.Background(), query.Interval1H)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Total != 100 {
		t.Errorf("expected Total=100, got %d", stats.Total)
	}
	if stats.Blocked != 40 {
		t.Errorf("expected Blocked=40, got %d", stats.Blocked)
	}
	if len(stats.TopBlocked) != 1 || stats.TopBlocked[0].Domain != "bad.com" {
		t.Errorf("unexpected TopBlocked: %v", stats.TopBlocked)
	}
	if len(stats.TopAllowed) != 1 || stats.TopAllowed[0].Domain != "ok.com" {
		t.Errorf("unexpected TopAllowed: %v", stats.TopAllowed)
	}
}

func TestGetDomainStats_FindDomainStatsError(t *testing.T) {
	svc, repo, _, _ := newService(t)
	repo.EXPECT().FindDomainStats(gomock.Any(), gomock.Any()).Return(database.DomainStats{}, errors.New("db error"))

	_, err := svc.GetDomainStats(context.Background(), query.Interval1H)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetDomainStats_FindTopBlockedError(t *testing.T) {
	svc, repo, _, _ := newService(t)
	repo.EXPECT().FindDomainStats(gomock.Any(), gomock.Any()).Return(database.DomainStats{}, nil)
	repo.EXPECT().FindTopDomains(gomock.Any(), true, gomock.Any(), 10).Return(nil, errors.New("db error"))

	_, err := svc.GetDomainStats(context.Background(), query.Interval1H)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetDomainStats_FindTopAllowedError(t *testing.T) {
	svc, repo, _, _ := newService(t)
	repo.EXPECT().FindDomainStats(gomock.Any(), gomock.Any()).Return(database.DomainStats{}, nil)
	repo.EXPECT().FindTopDomains(gomock.Any(), true, gomock.Any(), 10).Return(nil, nil)
	repo.EXPECT().FindTopDomains(gomock.Any(), false, gomock.Any(), 10).Return(nil, errors.New("db error"))

	_, err := svc.GetDomainStats(context.Background(), query.Interval1H)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// ---- GetDomainDetails ----

func TestGetDomainDetails_OK(t *testing.T) {
	svc, repo, blockFilter, _ := newService(t)

	now := time.Now().UTC()
	step := query.Granularity1H.ToDuration()
	point := database.Point{Time: now.Add(-step).Truncate(step), Count: 5}

	repo.EXPECT().FindDomainDetailsPoints(gomock.Any(), "example.com", gomock.Any(), step).Return([]database.Point{point}, nil)
	blockFilter.EXPECT().Filter("example.com").Return(false, nil)

	detail, err := svc.GetDomainDetails(context.Background(), "example.com", query.Interval1H, query.Granularity1H)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.Blocked {
		t.Error("expected Blocked=false")
	}
	if detail.Count != 1 {
		t.Errorf("expected Count=1, got %d", detail.Count)
	}
	if len(detail.Points) == 0 {
		t.Error("expected at least one time point")
	}
}

func TestGetDomainDetails_Blocked(t *testing.T) {
	svc, repo, blockFilter, _ := newService(t)

	step := query.Granularity1H.ToDuration()
	repo.EXPECT().FindDomainDetailsPoints(gomock.Any(), "bad.com", gomock.Any(), step).Return(nil, nil)
	blockFilter.EXPECT().Filter("bad.com").Return(true, nil)

	detail, err := svc.GetDomainDetails(context.Background(), "bad.com", query.Interval1H, query.Granularity1H)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !detail.Blocked {
		t.Error("expected Blocked=true")
	}
}

func TestGetDomainDetails_RepoError(t *testing.T) {
	svc, repo, _, _ := newService(t)

	step := query.Granularity1H.ToDuration()
	repo.EXPECT().FindDomainDetailsPoints(gomock.Any(), "example.com", gomock.Any(), step).Return(nil, errors.New("db error"))

	_, err := svc.GetDomainDetails(context.Background(), "example.com", query.Interval1H, query.Granularity1H)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetDomainDetails_BlockFilterError(t *testing.T) {
	svc, repo, blockFilter, _ := newService(t)

	step := query.Granularity1H.ToDuration()
	repo.EXPECT().FindDomainDetailsPoints(gomock.Any(), "example.com", gomock.Any(), step).Return(nil, nil)
	blockFilter.EXPECT().Filter("example.com").Return(false, errors.New("filter error"))

	_, err := svc.GetDomainDetails(context.Background(), "example.com", query.Interval1H, query.Granularity1H)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
