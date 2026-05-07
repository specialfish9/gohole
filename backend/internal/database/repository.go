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
	// FindAllLimit retrieves all queries from the database with an optional limit and name filter.
	// If `limit` is greater than 0, it limits the number of results returned. If `name` is not empty,
	// it filters the results to include only queries with the specified name.
	FindAllLimit(ctx context.Context, limit int, name string) ([]Query, error)
	// FindAllByInterval retrieves all queries from the database that were made
	// after the specified `since`.
	FindAllByInterval(ctx context.Context, since time.Time) ([]Query, error)
	FindHostStats(ctx context.Context, since time.Time) ([]HostStat, error)
	FindDomainStats(ctx context.Context, since time.Time) (DomainStats, error)
	FindTopDomains(ctx context.Context, blocked bool, since time.Time, limit int) ([]TopDomain, error)
	FindDomainDetailsPoints(ctx context.Context, name string, since time.Time, granularity time.Duration) ([]Point, error)
	// Close flushes any buffered writes and stops the background batch worker.
	// Call this on application shutdown.
	Close() error
}

const (
	batchSize       = 1000
	batchInterval   = 5 * time.Second
	flushRetryDelay = 1 * time.Second
	flushMaxRetries = 3
)

type repositoryImpl struct {
	conn    driver.Conn
	queue   chan Query
	done    chan struct{}
	stopped chan struct{}
}

// NewRepository creates a Repository backed by a ClickHouse connection.
// It starts a background goroutine that flushes buffered inserts every
// batchSize items or batchInterval, whichever comes first.
// Call Close() on shutdown to flush remaining items and stop the worker.
func NewRepository(conn driver.Conn) Repository {
	r := &repositoryImpl{
		conn:    conn,
		queue:   make(chan Query, batchSize*2), // buffer avoids blocking callers
		done:    make(chan struct{}),
		stopped: make(chan struct{}),
	}
	go r.batchWorker()
	return r
}

// SaveQuery enqueues a query for batched insertion.
// It returns immediately; the actual insert happens in the background worker.
func (r *repositoryImpl) SaveQuery(_ context.Context, q Query) error {
	select {
	case r.queue <- q:
		return nil
	case <-r.stopped:
		return fmt.Errorf("repository: already closed")
	}
}

// Close signals the batch worker to stop, waits for it to flush remaining
// items, then returns.
func (r *repositoryImpl) Close() error {
	close(r.done)
	<-r.stopped // wait until the worker has finished flushing
	return nil
}

// batchWorker runs in a goroutine and flushes the queue whenever 1000 items
// accumulate or 5 seconds elapse.
func (r *repositoryImpl) batchWorker() {
	defer close(r.stopped)

	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()

	batch := make([]Query, 0, batchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		for i := range flushMaxRetries {
			if err := r.flushBatch(batch); err != nil {
				slog.Error("batch flush failed", "error", err, "attempt", fmt.Sprintf("%d/%d", i, flushMaxRetries), "count", len(batch))
				if i == flushMaxRetries-1 {
					slog.Error("max flush retries reached, dropping batch", "count", len(batch))
					break
				}
				time.Sleep(flushRetryDelay)
			} else {
				break
			}
		}
		batch = batch[:0]
	}

	for {
		select {
		case q := <-r.queue:
			batch = append(batch, q)
			if len(batch) >= batchSize {
				flush()
			}

		case <-ticker.C:
			flush()

		case <-r.done:
			// Drain any remaining items in the channel before exiting.
			for {
				select {
				case q, ok := <-r.queue:
					if !ok {
						flush()
						return
					}
					batch = append(batch, q)
				default:
					flush()
					return
				}
			}
		}
	}
}

// flushBatch inserts a slice of queries in a single ClickHouse batch.
func (r *repositoryImpl) flushBatch(queries []Query) error {
	ctx := context.Background()

	b, err := r.conn.PrepareBatch(ctx, `
		INSERT INTO query (name, type, blocked, host, timestamp, millis)
	`)
	if err != nil {
		return fmt.Errorf("repository: prepare batch: %w", err)
	}

	for _, q := range queries {
		if err := b.Append(
			q.Name,
			uint16(0), // TODO
			q.Blocked,
			q.Host,
			time.Unix(q.Timestamp, 0),
			q.Millis,
		); err != nil {
			return fmt.Errorf("repository: append to batch: %w", err)
		}
	}

	if err := b.Send(); err != nil {
		return fmt.Errorf("repository: send batch: %w", err)
	}

	slog.Debug("batch flushed", "count", len(queries))
	return nil
}

func (r *repositoryImpl) FindAll(ctx context.Context) ([]Query, error) {
	return r.FindAllLimit(ctx, -1, "")
}

func (r *repositoryImpl) FindAllLimit(ctx context.Context, limit int, name string) ([]Query, error) {
	baseQuery := `
		SELECT name, type, host, blocked, timestamp, millis
		FROM query
  `

	args := []any{}

	if name != "" {
		baseQuery += " WHERE name ilike ?"
		args = append(args, "%"+name+"%")
	}

	baseQuery += "ORDER BY timestamp DESC"

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
	q := `
    SELECT name, type, blocked, timestamp
    FROM query
		WHERE timestamp >= ?
	  ORDER BY timestamp DESC`

	rows, err := r.conn.Query(ctx, q, since)

	if err != nil {
		return nil, fmt.Errorf("repository: cannot fetch queries: %w", err)
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

func (r *repositoryImpl) FindDomainDetailsPoints(ctx context.Context, name string, since time.Time, granularity time.Duration) ([]Point, error) {
	var aggr string

	switch granularity {
	case time.Minute:
		aggr = "toStartOfMinute(timestamp)"
	case time.Hour:
		aggr = "toStartOfHour(timestamp)"
	case 24 * time.Hour:
		aggr = "toStartOfDay(timestamp)"
	default:
		return nil, fmt.Errorf("repository: unsupported granularity: %v", granularity)
	}

	q := fmt.Sprintf(`
		SELECT %s AS time, count() AS count
		FROM query
		where name = ? AND timestamp >= ?
		GROUP BY time
		ORDER BY time;
	`, aggr)

	rows, err := r.conn.Query(ctx, q, name, since)
	if err != nil {
		return nil, fmt.Errorf("repository: cannot fetch domain details points: %w", err)
	}
	defer rows.Close()

	var points []Point

	for rows.Next() {
		var p Point
		if err := rows.Scan(&p.Time, &p.Count); err != nil {
			slog.Error("scan failed", "error", err)
			continue
		}
		points = append(points, p)
	}

	return points, nil
}
