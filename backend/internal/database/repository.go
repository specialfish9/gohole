package database

type Repository interface {
	SaveQuery(q Query) error
	FindAll() ([]Query, error)
	FindAllByInterval(interval string) ([]Query, error)
}

type inMemoryRepository struct {
	queries []Query
}

func NewInMemoryRepository() *inMemoryRepository {
	return &inMemoryRepository{
		queries: make([]Query, 0),
	}
}

func (r *inMemoryRepository) SaveQuery(q Query) error {
	r.queries = append(r.queries, q)
	return nil
}

func (r *inMemoryRepository) FindAll() ([]Query, error) {
	return r.queries, nil
}

func (r *inMemoryRepository) FindAllByInterval(interval string) ([]Query, error) {
	// For simplicity, this implementation ignores the interval and returns all queries.
	// A real implementation would filter based on the interval.
	return r.queries, nil
}
