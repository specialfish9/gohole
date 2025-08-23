package query

import (
	"fmt"
	"gohole/internal/database"
	"log"
	"math"
	"time"
)

type Service interface {
	Save(q database.Query) error
	GetAll() ([]database.Query, error)
	ShouldAllow(name string) (bool, error)
	GetStats(interval string) (*Stats, error)
	GetHistory(interval Interval, granularity Granularity) ([]QueryHistoryPoint, error)
}

type serviceImpl struct {
	repo   database.Repository
	filter Filter
}

func NewService(filter Filter) Service {
	return &serviceImpl{
		repo:   database.NewInMemoryRepository(), // TODO
		filter: filter,
	}
}

func (s *serviceImpl) Save(q database.Query) error {
	return s.repo.SaveQuery(q)
}

func (s *serviceImpl) GetAll() ([]database.Query, error) {
	return s.repo.FindAll()
}

func (s *serviceImpl) ShouldAllow(name string) (bool, error) {
	return s.filter.Filter(name)
}

func (s *serviceImpl) GetStats(interval string) (*Stats, error) {
	var err error
	var queries []database.Query

	if interval == "" {
		queries, err = s.repo.FindAll()
	} else {
		queries, err = s.repo.FindAllByInterval(interval)
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

func (s *serviceImpl) GetHistory(interval Interval, granularity Granularity) ([]QueryHistoryPoint, error) {
	timeStep := granularity.ToDuration()
	stepsNo := interval.ToDuration() / timeStep

	history := make([]QueryHistoryPoint, stepsNo)
	log.Printf("Debug: interval=%s, granularity=%s, stepsNo=%d, timeStep=%d", interval, granularity, stepsNo, timeStep)

	startTs := time.Now().Unix() - interval.ToDuration()

	// First, set all the timestamps
	for i, _ := range history {
		ts := startTs + (int64(i) * timeStep)
		history[i].Time = time.Unix(ts, 0).Format(time.RFC3339)
	}

	// Fetch all the queries
	queries, err := s.repo.FindAllByInterval("TODO") // TODO: use interval
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
		index := int((query.Timestamp - startTs) / timeStep)

		if query.Blocked {
			history[index].Blocked++
		} else {
			history[index].Allowed++
		}
	}

	return history, nil
}
