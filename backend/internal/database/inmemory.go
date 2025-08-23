package database

import "context"

type inMemoryRepository struct {
	queries []Query
}

func NewInMemoryRepository() *inMemoryRepository {
	return &inMemoryRepository{
		queries: make([]Query, 0),
	}
}

func (r *inMemoryRepository) SaveQuery(_ context.Context, q Query) error {
	r.queries = append(r.queries, q)
	return nil
}

func (r *inMemoryRepository) FindAll(_ context.Context) ([]Query, error) {
	return r.queries, nil
}

func (r *inMemoryRepository) FindAllByInterval(_ context.Context, interval string) ([]Query, error) {
	// For simplicity, this implementation ignores the interval and returns all queries.
	// A real implementation would filter based on the interval.
	return r.queries, nil
}
