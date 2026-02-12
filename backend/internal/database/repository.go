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
	FindAllLimit(ctx context.Context, limit int) ([]Query, error)
	// FindAllByInterval retrieves all queries from the database that were made
	// after the specified `since`.
	FindAllByInterval(ctx context.Context, since time.Time) ([]Query, error)
	FindHostStats(ctx context.Context, since time.Time) ([]HostStat, error)
	FindDomainStats(ctx context.Context, since time.Time) (DomainStats, error)
	FindTopDomains(ctx context.Context, blocked bool, since time.Time, limit int) ([]TopDomain, error)
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
    INSERT INTO query (name, type, blocked, host, timestamp, millis)
    VALUES (?, ?, ?, ?, ?, ?)
    `,
		q.Name,
		uint16(0), // TODO
		q.Blocked,
		q.Host,
		time.Unix(q.Timestamp, 0), // Convert int64 to time.Time
		q.Millis,
	)

	if err != nil {
		return fmt.Errorf("repository: cannot save query: %w", err)
	}

	return nil
}

func (r *repositoryImpl) FindAll(ctx context.Context) ([]Query, error) {
	return r.FindAllLimit(ctx, -1)
}

func (r *repositoryImpl) FindAllLimit(ctx context.Context, limit int) ([]Query, error) {
	baseQuery := `
		SELECT name, type, host, blocked, timestamp, millis
		FROM query
		ORDER BY timestamp DESC
  `
	args := []any{}

	if limit > 0 {
		baseQuery += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := r.conn.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("repository: cannot fetch all queries: %w", err)
	}
	defer rows.Close()

	var queries []Query

	for rows.Next() {
		var q Query
		var blockedUInt8 uint8

		err := rows.Scan(&q.Name, &q.Type, &q.Host, &blockedUInt8, &q.Timestamp, &q.Millis)
		if err != nil {
			slog.Error("scan failed", "error", err)
			continue
		}

		q.Blocked = blockedUInt8 != 0
		queries = append(queries, q)
	}

	return queries, nil
}

func (r *repositoryImpl) FindAllByInterval(ctx context.Context, since time.Time) ([]Query, error) {
	rows, err := r.conn.Query(ctx, `
    SELECT name, type, blocked, timestamp
    FROM query
		WHERE timestamp >= ?
    ORDER BY timestamp DESC
	`, since)

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

func (r *repositoryImpl) FindHostStats(ctx context.Context, since time.Time) ([]HostStat, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT
			host,
			COUNT(*) AS queryCount,
			SUM(blocked) AS blockedCount,
			ROUND(100.0 * SUM(blocked) / COUNT(*), 2) AS blockRate
		FROM query
		WHERE timestamp >= ?
		GROUP BY host
		ORDER BY queryCount DESC
	`, since)
	if err != nil {
		return nil, fmt.Errorf("repository: cannot fetch host stats: %w", err)
	}
	defer rows.Close()

	var stats []HostStat

	for rows.Next() {
		var hs HostStat
		if err := rows.Scan(&hs.Host, &hs.QueryCount, &hs.BlockedCount, &hs.BlockRate); err != nil {
			slog.Error("scan failed", "error", err)
			continue
		}
		stats = append(stats, hs)
	}

	return stats, nil
}

func (r *repositoryImpl) FindDomainStats(ctx context.Context, since time.Time) (DomainStats, error) {
	var stats DomainStats

	rows, err := r.conn.Query(ctx, `
		SELECT
    countDistinctIf(name, blocked = true)  AS blocked_count,
    countDistinct(name) AS total
		FROM query
		WHERE timestamp >= ?
	`, since)
	if err != nil {
		return stats, fmt.Errorf("repository: cannot fetch domain stats: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&stats.BlockedCount, &stats.Total); err != nil {
			return stats, fmt.Errorf("repository: cannot scan domain stats: %w", err)
		}
	} else {
		return stats, fmt.Errorf("repository: no domain stats found")
	}

	return stats, nil
}

func (r *repositoryImpl) FindTopDomains(ctx context.Context, blocked bool, since time.Time, limit int) ([]TopDomain, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT
			name AS domain,
			COUNT(*) AS blockedCount
		FROM query
		WHERE blocked = ? AND timestamp >= ?
		GROUP BY name
		ORDER BY blockedCount DESC
		LIMIT ?
	`, blocked, since, limit)
	if err != nil {
		return nil, fmt.Errorf("repository: cannot fetch top blocked domains: %w", err)
	}
	defer rows.Close()

	var domains []TopDomain

	for rows.Next() {
		var td TopDomain
		if err := rows.Scan(&td.Domain, &td.Count); err != nil {
			slog.Error("scan failed", "error", err)
			continue
		}
		domains = append(domains, td)
	}

	return domains, nil
}
