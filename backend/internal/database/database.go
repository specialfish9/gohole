package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

func Connect(cfg *Config, attempts int) (driver.Conn, error) {
	if cfg == nil {
		panic("Can't inizialize DB with nil config!")
	}

	var conn driver.Conn
	var err error

	success := false

	for i := range attempts {
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{cfg.Address},
			Auth: clickhouse.Auth{
				Database: cfg.Name,
				Username: cfg.User,
				Password: cfg.Password,
			},
			DialTimeout: 5 * time.Second,
			Debug:       cfg.Debug.Or(false),
		})

		if err != nil {
			slog.Error(fmt.Sprintf("DB connection attempt %d failed: %v", i+1, err))
			time.Sleep(2 * time.Second)
		} else {
			success = true
			break
		}
	}

	if !success {
		return nil, fmt.Errorf("db: unable to connect to database")
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
			host String,
			timestamp DateTime,
			millis Int64,
		) ENGINE = MergeTree() 
			ORDER BY (timestamp, type);
		`,
		`
		CREATE TABLE IF NOT EXISTS blocklist (
			url String
		) ENGINE = MergeTree()
		ORDER BY url;
			`,
		`DELETE FROM blocklist WHERE true;`,
	}

	for i, query := range queries {
		if err := conn.Exec(ctx, query); err != nil {
			return fmt.Errorf("db: cannot create initial table (%d): %w", i, err)
		}
	}

	return nil
}
