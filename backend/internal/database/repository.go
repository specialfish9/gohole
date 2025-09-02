package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Repository interface {
	SaveQuery(ctx context.Context, q Query) error
	FindAll(ctx context.Context) ([]Query, error)
	FindAllByInterval(ctx context.Context, interval int64) ([]Query, error)
}

type repositoryImpl struct {
	conn driver.Conn
}

func NewRepository(conn driver.Conn) Repository {
	return &repositoryImpl{
		conn: conn,
	}
}

func (r *repositoryImpl) SaveQuery(ctx context.Context, q Query) error {
	err := r.conn.Exec(ctx, `
    INSERT INTO query (name, type, blocked, timestamp)
    VALUES (?, ?, ?, ?)
    `,
		q.Name,
		uint16(0), // TODO
		q.Blocked,
		time.Unix(q.Timestamp, 0), // Convert int64 to time.Time
		q.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("repository: cannot save query: %w", err)
	}

	return nil
}

func (r *repositoryImpl) FindAll(ctx context.Context) ([]Query, error) {
	rows, err := r.conn.Query(ctx, `
    SELECT name, type, blocked, timestamp
    FROM query
    ORDER BY timestamp DESC
	`)

	if err != nil {
		return nil, fmt.Errorf("repository: cannot fetch all queries: %w", err)
	}
	defer rows.Close()

	var queries []Query

	for rows.Next() {
		var q Query
		var blockedUInt8 uint8

		err := rows.Scan(&q.Name, &q.Type, &blockedUInt8, &q.Timestamp)
		if err != nil {
			slog.Error("scan failed", "error", err)
			continue
		}

		q.Blocked = blockedUInt8 != 0
		queries = append(queries, q)
	}

	return queries, nil
}

// FindAllByInterval fetches all queries made within the last 'interval' seconds.
func (r *repositoryImpl) FindAllByInterval(ctx context.Context, interval int64) ([]Query, error) {
	rows, err := r.conn.Query(ctx, `
    SELECT name, type, blocked, timestamp
    FROM query
		WHERE timestamp >= now() - INTERVAL ? SECOND
    ORDER BY timestamp DESC
	`, interval)

	if err != nil {
		return nil, fmt.Errorf("repository: cannot fetch all queries: %w", err)
	}

	defer rows.Close()

	var queries []Query

	for rows.Next() {
		var q Query
		var blockedUInt8 uint8
		if err := rows.Scan(&q.Name, &q.Type, &blockedUInt8, &q.Timestamp); err != nil {
			slog.Error("scan failed", "error", err)
			continue
		}
		q.Blocked = blockedUInt8 != 0
		queries = append(queries, q)
	}

	return queries, nil
}
