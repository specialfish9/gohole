package pg

import (
	"context"
	"fmt"
	"gohole/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Manager struct {
	cfg  *database.Config
	pool *pgxpool.Pool
}

func NewManager(cfg *database.Config) *Manager {
	return &Manager{
		cfg: cfg,
	}
}

func (m *Manager) Repository() database.Repository {
	return NewRepository(m)
}

func (m *Manager) Connect(ctx context.Context) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		m.cfg.User,
		m.cfg.Password,
		m.cfg.Address,
		m.cfg.Name,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("postgres: unable to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("postgres: cannot ping db: %w", err)
	}

	m.pool = pool

	return nil
}

func (m *Manager) Init(ctx context.Context) error {
	if m.pool == nil {
		return fmt.Errorf("postgres: connection is not initialized")
	}

	queries := []string{`
		CREATE TABLE IF NOT EXISTS "query" (
			name TEXT,
			type SMALLINT,
			blocked BOOLEAN,
			host TEXT,
			timestamp TIMESTAMP,
			millis BIGINT
		);`,

		`CREATE INDEX IF NOT EXISTS query_timestamp_type_idx
		ON "query" (timestamp, type);`,
	}

	for i, query := range queries {
		if _, err := m.pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("postgres: cannot create initial table (%d): %w", i, err)
		}
	}

	return nil
}
