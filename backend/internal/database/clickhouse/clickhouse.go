package clickhouse

import (
	"context"
	"fmt"
	"gohole/internal/database"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Manager struct {
	conn driver.Conn
	cfg  *database.Config
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
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{m.cfg.Address},
		Auth: clickhouse.Auth{
			Database: m.cfg.Name,
			Username: m.cfg.User,
			Password: m.cfg.Password,
		},
		DialTimeout: 5 * time.Second,
		Debug:       m.cfg.Debug.Or(false),
	})

	if err != nil {
		return fmt.Errorf("clickhouse: unable to connect to database: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("clickhouse: cannot ping db: %w", err)
	}

	m.conn = conn

	return nil
}

func (m *Manager) Init(ctx context.Context) error {
	if m.conn == nil {
		return fmt.Errorf("clickhouse: connection is not initialized")
	}

	ctx = clickhouse.Context(ctx, clickhouse.WithSettings(clickhouse.Settings{
		"max_execution_time": 60,
	}))

	// Create the queries table if it doesn't exist
	queries := []string{`
		CREATE TABLE IF NOT EXISTS query (
			name String,
			type UInt16,
			blocked UInt8,
			host String,
			timestamp DateTime,
			millis Int64,
		) ENGINE = MergeTree() 
			ORDER BY (timestamp, type);
		`,
	}

	for i, query := range queries {
		if err := m.conn.Exec(ctx, query); err != nil {
			return fmt.Errorf("clickhouse: cannot create initial table (%d): %w", i, err)
		}
	}

	return nil
}
