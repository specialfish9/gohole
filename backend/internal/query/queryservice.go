package query

import (
	"context"
	"fmt"
	"gohole/internal/database"
	"log/slog"
	"math"
	"time"
)

type Service interface {
	Save(ctx context.Context, q database.Query) error
	GetAll(ctx context.Context) ([]database.Query, error)
	GetStats(ctx context.Context, interval Interval) (*Stats, error)
	GetHistory(ctx context.Context, interval Interval, granularity Granularity) ([]QueryHistoryPoint, error)

	ShouldAllow(name string) (bool, error)
}

type serviceImpl struct {
	repo   database.Repository
	filter Filter
}

func NewService(filter Filter, repo database.Repository) Service {
	return &serviceImpl{
		filter: filter,
		repo:   repo,
	}
}

func (s *serviceImpl) Save(ctx context.Context, q database.Query) error {
	return s.repo.SaveQuery(ctx, q)
}

func (s *serviceImpl) GetAll(ctx context.Context) ([]database.Query, error) {
	return s.repo.FindAll(ctx)
}

func (s *serviceImpl) ShouldAllow(name string) (bool, error) {
	return s.filter.Filter(name)
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
	for i, _ := range history {
		ts := startTs.Add(granularity.ToDuration() * time.Duration(i))
		slog.Debug("history point", "i", i, "ts", ts)
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
		index := int((query.Timestamp - startTs.UnixMilli()) / int64(granularity.ToDuration().Milliseconds()))

		if query.Blocked {
			history[index].Blocked++
		} else {
			history[index].Allowed++
		}
	}

	return history, nil
}
