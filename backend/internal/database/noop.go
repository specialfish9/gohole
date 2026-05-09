package database

import (
	"context"
	"time"
)

type NoOpManager struct {
}

func NewNoOpManager() *NoOpManager {
	return &NoOpManager{}
}

func (m *NoOpManager) Repository() Repository {
	return NewNoOpRepostory()
}

func (m *NoOpManager) Connect(ctx context.Context) error {
	return nil
}

func (m *NoOpManager) Init(ctx context.Context) error {
	return nil
}

type NoOpRepository struct {
}

func NewNoOpRepostory() Repository {
	return &NoOpRepository{}
}

func (r *NoOpRepository) Close() error {
	return nil
}

func (r *NoOpRepository) SaveQuery(ctx context.Context, q Query) error {
	return nil
}

func (r *NoOpRepository) FindAll(ctx context.Context) ([]Query, error) {
	return []Query{}, nil
}

func (r *NoOpRepository) FindAllLimit(ctx context.Context, limit int, name string) ([]Query, error) {
	return []Query{}, nil
}

func (r *NoOpRepository) FindAllByInterval(ctx context.Context, since time.Time) ([]Query, error) {
	return []Query{}, nil
}

func (r *NoOpRepository) FindHostStats(ctx context.Context, since time.Time) ([]HostStat, error) {
	return []HostStat{}, nil
}

func (r *NoOpRepository) FindDomainStats(ctx context.Context, since time.Time) (DomainStats, error) {
	return DomainStats{}, nil
}

func (r *NoOpRepository) FindTopDomains(ctx context.Context, blocked bool, since time.Time, limit int) ([]TopDomain, error) {
	return []TopDomain{}, nil
}

func (r *NoOpRepository) FindDomainDetailsPoints(
	ctx context.Context,
	name string,
	since time.Time,
	granularity time.Duration,
) ([]Point, error) {
	return nil, nil
}
