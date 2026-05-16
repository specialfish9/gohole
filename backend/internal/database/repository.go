package database

//go:generate go tool go.uber.org/mock/mockgen -destination=../mock/database/repository.go -typed -package mockrepo gohole/internal/database Repository

import (
	"context"
	"time"
)

type Repository interface {
	SaveQuery(ctx context.Context, q Query) error
	FindAll(ctx context.Context) ([]Query, error)
	// FindAllLimit retrieves all queries from the database with an optional limit and name filter.
	// If `limit` is greater than 0, it limits the number of results returned. If `name` is not empty,
	// it filters the results to include only queries with the specified name.
	FindAllLimit(ctx context.Context, limit int, name string) ([]Query, error)
	// FindAllByInterval retrieves all queries from the database that were made
	// after the specified `since`.
	FindAllByInterval(ctx context.Context, since time.Time) ([]Query, error)
	FindHostStats(ctx context.Context, since time.Time) ([]HostStat, error)
	FindDomainStats(ctx context.Context, since time.Time) (DomainStats, error)
	FindTopDomains(ctx context.Context, blocked bool, since time.Time, limit int) ([]TopDomain, error)
	FindDomainDetailsPoints(ctx context.Context, name string, since time.Time, granularity time.Duration) ([]Point, error)
	// Close flushes any buffered writes and stops the background batch worker.
	// Call this on application shutdown.
	Close() error
}
