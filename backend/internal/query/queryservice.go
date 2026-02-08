package query

import (
	"context"
	"fmt"
	"gohole/internal/database"
	"gohole/internal/filter"
	"math"
	"time"
)

type Service interface {
	Save(ctx context.Context, q database.Query) error
	GetAll(ctx context.Context, limit int) ([]database.Query, error)
	GetStats(ctx context.Context, interval Interval) (*Stats, error)
	GetHistory(ctx context.Context, interval Interval, granularity Granularity) ([]QueryHistoryPoint, error)
	GetBlockListStats() (*BlockListStats, error)
	GetHostStats(ctx context.Context, interival Interval) ([]database.HostStat, error)
	GetDomainStats(ctx context.Context, interval Interval) (DomainStats, error)
	ShouldAllow(name string) (bool, error)
}

type serviceImpl struct {
	repo        database.Repository
	blockFilter filter.Filter
	allowFilter filter.Filter
}

func NewService(blockFilter filter.Filter, allowFilter filter.Filter, repo database.Repository) Service {
	return &serviceImpl{
		blockFilter: blockFilter,
		allowFilter: allowFilter,
		repo:        repo,
	}
}

func (s *serviceImpl) Save(ctx context.Context, q database.Query) error {
	return s.repo.SaveQuery(ctx, q)
}

func (s *serviceImpl) GetAll(ctx context.Context, limit int) ([]database.Query, error) {
	return s.repo.FindAllLimit(ctx, limit)
}

// ShouldAllow checks if a query should be allowed or blocked based on the allow and block filters.
// It returns true if the query should be allowed, false if it should be blocked.
func (s *serviceImpl) ShouldAllow(name string) (bool, error) {
	if name[len(name)-1] == '.' {
		name = name[:len(name)-1]
	}

	isAllowed, err := s.allowFilter.Filter(name)
	if err != nil {
		return false, fmt.Errorf("query service: error checking allow filter: %w", err)
	}

	if isAllowed {
		return true, nil
	}

	isBlocked, err := s.blockFilter.Filter(name)
	if err != nil {
		return false, fmt.Errorf("query service: error checking block filter: %w", err)
	}

	return !isBlocked, nil
}

func (s *serviceImpl) GetStats(ctx context.Context, interval Interval) (*Stats, error) {
	var err error
	var queries []database.Query

	if interval == "" {
		queries, err = s.repo.FindAll(ctx)
	} else {
		queries, err = s.repo.FindAllByInterval(ctx, time.Now().UTC().Add(-time.Duration(interval.ToDuration())))
	}

	if err != nil {
		return nil, fmt.Errorf("query service: cannot fetch queries: %w", err)
	}

	var blocked, allowed int
	var blockRate float64
	total := len(queries)

	for _, q := range queries {
		if q.Blocked {
			blocked++
		} else {
			allowed++
		}
	}

	// x : 100 = blocked : total
	blockRate = math.Round(float64(100.0*blocked) / float64(total))

	return &Stats{
		TotalQueries:   total,
		BlockedQueries: blocked,
		AllowedQueries: allowed,
		BlockRate:      blockRate,
	}, nil
}

func (s *serviceImpl) GetHistory(ctx context.Context, interval Interval, granularity Granularity) ([]QueryHistoryPoint, error) {
	granularityStep := granularity.ToDuration().Seconds()
	stepsNo := interval.ToDuration().Seconds() / granularityStep

	history := make([]QueryHistoryPoint, int(math.Ceil(stepsNo)))

	// FIXME: this does not handle the timezone properly. Fix it later.
	startTs := time.Now().UTC().Add(-interval.ToDuration())

	// First, set all the timestamps
	for i := range history {
		ts := startTs.Add(granularity.ToDuration() * time.Duration(i))
		history[i].Time = ts.Format(time.RFC3339)
	}

	// Fetch all the queries
	queries, err := s.repo.FindAllByInterval(ctx, startTs)
	if err != nil {
		return nil, fmt.Errorf("query service: cannot fetch queries: %w", err)
	}

	if len(queries) == 0 {
		// early exit in case there's no data
		return history, nil
	}

	// Then, update all the history points
	for _, query := range queries {
		// Index represents which history point this query belongs to
		index := int((query.Timestamp - startTs.Unix()) / int64(granularity.ToDuration().Seconds()))

		if query.Blocked {
			history[index].Blocked++
		} else {
			history[index].Allowed++
		}
	}

	return history, nil
}

func (s *serviceImpl) GetBlockListStats() (*BlockListStats, error) {
	return &BlockListStats{
		TotalEntries: s.blockFilter.Size(),
	}, nil
}

func (s *serviceImpl) GetHostStats(ctx context.Context, interval Interval) ([]database.HostStat, error) {
	since := time.Now().UTC().Add(-interval.ToDuration())
	return s.repo.FindHostStats(ctx, since)
}

func (s *serviceImpl) GetDomainStats(ctx context.Context, interval Interval) (DomainStats, error) {
	since := time.Now().UTC().Add(-interval.ToDuration())
	ds, err := s.repo.FindDomainStats(ctx, since)
	if err != nil {
		return DomainStats{}, fmt.Errorf("query service: cannot fetch domain stats: %w", err)
	}

	// Fetch top blocked domains
	blocked, err := s.repo.FindTopDomains(ctx, true, since, 10)
	if err != nil {
		return DomainStats{}, fmt.Errorf("query service: cannot fetch top blocked domains: %w", err)
	}

	// Fetch top allowed domains
	allowed, err := s.repo.FindTopDomains(ctx, false, since, 10)
	if err != nil {
		return DomainStats{}, fmt.Errorf("query service: cannot fetch top allowed domains: %w", err)
	}

	var ret DomainStats
	ret.Total = ds.Total
	ret.Blocked = ds.BlockedCount
	ret.TopBlocked = blocked
	ret.TopAllowed = allowed

	return ret, nil
}
