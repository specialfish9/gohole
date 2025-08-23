package query

import "gohole/internal/database"

type Service interface {
	Save(q database.Query) error
	GetAll() ([]database.Query, error)
	ShouldAllow(name string) (bool, error)
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
