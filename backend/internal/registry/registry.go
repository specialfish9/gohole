package registry

import (
	"gohole/internal/database"
	"gohole/internal/filter"
	"gohole/internal/query"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Registry struct {
	QueryRepository database.Repository
	QueryService    query.Service
}

func NewRegistry(blockedDomains []string, allowedDomains []string, filterStrategy filter.Strategy, conn driver.Conn) *Registry {
	blockFilter := filter.NewFilter(filterStrategy, blockedDomains)
	allowFilter := filter.NewFilter(filterStrategy, allowedDomains)

	repo := database.NewRepository(conn)

	return &Registry{
		QueryRepository: repo,
		QueryService:    query.NewService(blockFilter, allowFilter, repo),
	}
}
