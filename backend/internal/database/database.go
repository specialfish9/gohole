package database

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

func Connect(address string, db string, username string, password string, debug bool) (driver.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{address},
		Auth: clickhouse.Auth{
			Database: db,
			Username: username,
			Password: password,
		},
		DialTimeout: 5 * time.Second,
		Debug:       debug, // prints queries for debugging
	})
	if err != nil {
		return nil, fmt.Errorf("db: cannot connect to db: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("db: cannot ping db: %w", err)
	}

	return conn, nil
}

func Init(ctx context.Context, conn driver.Conn) error {
	ctx = clickhouse.Context(ctx, clickhouse.WithSettings(clickhouse.Settings{
		"max_execution_time": 60,
	}))

	// Create the queries table if it doesn't exist
	queries := []string{`
		CREATE TABLE IF NOT EXISTS query (
			name String,
			type UInt16,
			blocked UInt8,
			timestamp DateTime64(3, 'UTC')  -- stored with ms precision
		) ENGINE = MergeTree() 
			ORDER BY (timestamp, type);
		`,
		`
		CREATE TABLE IF NOT EXISTS blocklist (
			url String
		) ENGINE = MergeTree()
		ORDER BY url;
			`,
	}

	for i, query := range queries {
		if err := conn.Exec(ctx, query); err != nil {
			return fmt.Errorf("db: cannot create initial table (%d): %w", i, err)
		}
	}

	return nil
}
