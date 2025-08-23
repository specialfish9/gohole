package database

type Repository interface {
	SaveQuery(q Query) error
	FindAll() ([]Query, error)
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
