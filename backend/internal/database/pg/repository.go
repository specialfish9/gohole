package pg

import (
	"context"
	"fmt"
	"gohole/internal/database"
	"time"

	"github.com/jackc/pgx/v5"
)

type repositoryImpl struct {
	mngr *Manager
}

func NewRepository(manager *Manager) database.Repository {
	r := &repositoryImpl{
		mngr: manager,
	}
	return r
}

func (r *repositoryImpl) Close() error {
	return nil
}

func (r *repositoryImpl) SaveQuery(ctx context.Context, q database.Query) error {
	_, err := r.mngr.pool.Exec(ctx, `
		INSERT INTO query (name, type, blocked, host, timestamp, millis)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		q.Name,
		q.Type,
		q.Blocked,
		q.Host,
		time.Unix(q.Timestamp, 0),
		q.Millis,
	)

	if err != nil {
		return fmt.Errorf("repository: insert query failed: %w", err)
	}

	return nil
}

func (r *repositoryImpl) FindAll(ctx context.Context) ([]database.Query, error) {
	return r.FindAllLimit(ctx, -1, "")
}

func (r *repositoryImpl) FindAllLimit(ctx context.Context, limit int, name string) ([]database.Query, error) {
	base := `
		SELECT name, type, host, blocked, timestamp, millis
		FROM query
	`

	args := []any{}
	i := 1

	if name != "" {
		base += fmt.Sprintf(" WHERE name ILIKE $%d", i)
		args = append(args, "%"+name+"%")
		i++
	}

	base += " ORDER BY timestamp DESC"

	if limit > 0 {
		base += fmt.Sprintf(" LIMIT $%d", i)
		args = append(args, limit)
	}

	rows, err := r.mngr.pool.Query(ctx, base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[database.Query])
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *repositoryImpl) FindAllByInterval(ctx context.Context, since time.Time) ([]database.Query, error) {
	rows, err := r.mngr.pool.Query(ctx, `
		SELECT name, type, blocked, timestamp
		FROM query
		WHERE timestamp >= $1
		ORDER BY timestamp DESC
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[database.Query])
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *repositoryImpl) FindHostStats(ctx context.Context, since time.Time) ([]database.HostStat, error) {
	rows, err := r.mngr.pool.Query(ctx, `
		SELECT
			host,
			COUNT(*) AS query_count,
			SUM(CASE WHEN blocked THEN 1 ELSE 0 END) AS blocked_count,
			ROUND(100.0 * SUM(CASE WHEN blocked THEN 1 ELSE 0 END) / COUNT(*), 2) AS block_rate
		FROM query
		WHERE timestamp >= $1
		GROUP BY host
		ORDER BY query_count DESC
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats, err := pgx.CollectRows(rows, pgx.RowToStructByName[database.HostStat])
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *repositoryImpl) FindDomainStats(ctx context.Context, since time.Time) (database.DomainStats, error) {
	var stats database.DomainStats

	err := r.mngr.pool.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT CASE WHEN blocked THEN name END) AS blocked_count,
			COUNT(DISTINCT name) AS total
		FROM query
		WHERE timestamp >= $1
	`, since).Scan(&stats.BlockedCount, &stats.Total)

	return stats, err
}

func (r *repositoryImpl) FindTopDomains(ctx context.Context, blocked bool, since time.Time, limit int) ([]database.TopDomain, error) {
	rows, err := r.mngr.pool.Query(ctx, `
		SELECT name, COUNT(*) AS cnt
		FROM query
		WHERE blocked = $1 AND timestamp >= $2
		GROUP BY name
		ORDER BY cnt DESC
		LIMIT $3
	`, blocked, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[database.TopDomain])
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *repositoryImpl) FindDomainDetailsPoints(
	ctx context.Context,
	name string,
	since time.Time,
	granularity time.Duration,
) ([]database.Point, error) {

	var bucket string

	switch granularity {
	case time.Minute:
		bucket = "date_trunc('minute', timestamp)"
	case time.Hour:
		bucket = "date_trunc('hour', timestamp)"
	case 24 * time.Hour:
		bucket = "date_trunc('day', timestamp)"
	default:
		return nil, fmt.Errorf("unsupported granularity")
	}

	q := fmt.Sprintf(`
		SELECT %s AS time, COUNT(*)
		FROM query
		WHERE name = $1 AND timestamp >= $2
		GROUP BY time
		ORDER BY time
	`, bucket)

	rows, err := r.mngr.pool.Query(ctx, q, name, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[database.Point])
	if err != nil {
		return nil, err
	}

	return res, nil
}
