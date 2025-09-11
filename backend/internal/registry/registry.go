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

func NewRegistry(domains []string, conn driver.Conn) *Registry {
	filter := filter.Trie(domains)
	repo := database.NewRepository(conn)

	return &Registry{
		QueryRepository: repo,
		QueryService:    query.NewService(filter, repo),
	}
}
