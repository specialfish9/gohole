package registry

import "gohole/internal/query"

type Registry struct {
	QueryService query.Service
}

func NewRegistry(domains []string) *Registry {
	filter := query.Trie(domains)
	return &Registry{
		QueryService: query.NewService(filter),
	}
}
