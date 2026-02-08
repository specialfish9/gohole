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

func NewRegistry(blockedDomains []string, allowedDomains []string, conn driver.Conn) *Registry {
	blockFilter := filter.NewTrie(blockedDomains)
	allowFilter := filter.NewTrie(allowedDomains)

	repo := database.NewRepository(conn)

	return &Registry{
		QueryRepository: repo,
		QueryService:    query.NewService(blockFilter, allowFilter, repo),
	}
}
