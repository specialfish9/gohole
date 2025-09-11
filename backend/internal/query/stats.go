package query

import "time"

type Stats struct {
	TotalQueries   int     `json:"totalQueries"`
	BlockedQueries int     `json:"blockedQueries"`
	AllowedQueries int     `json:"allowedQueries"`
	BlockRate      float64 `json:"blockRate"`
}

type QueryHistoryPoint struct {
	Time    string `json:"time"`
	Blocked int    `json:"blocked"`
	Allowed int    `json:"allowed"`
}

type Interval string

const (
	Interval1H  Interval = "1h"
	Interval6H  Interval = "6h"
	Interval1D  Interval = "24h"
	Interval7D  Interval = "7d"
	Interval30D Interval = "30d"
)

func (i Interval) IsValid() bool {
	switch i {
	case Interval1H, Interval6H, Interval1D, Interval7D, Interval30D:
		return true
	default:
		return false
	}
}

// ToDuration converts the interval to its equivalent duration in seconds.
func (i Interval) ToDuration() time.Duration {
	switch i {
	case Interval1H:
		return 3600 * time.Second
	case Interval6H:
		return 21600 * time.Second
	case Interval1D:
		return 86400 * time.Second
	case Interval7D:
		return 604800 * time.Second
	case Interval30D:
		return 2592000 * time.Second
	default:
		return 0
	}
}

type Granularity string

const (
	Granularity1M  Granularity = "1m"
	Granularity5M  Granularity = "5m"
	Granularity15M Granularity = "15m"
	Granularity1H  Granularity = "1h"
	Granularity6H  Granularity = "6h"
	Granularity1D  Granularity = "1d"
)

func (g Granularity) IsValid() bool {
	switch g {
	case Granularity1M, Granularity5M, Granularity15M, Granularity1H, Granularity6H, Granularity1D:
		return true
	default:
		return false
	}
}

// ToDuration converts the granularity to its equivalent duration.
func (g Granularity) ToDuration() time.Duration {
	switch g {
	case Granularity1M:
		return 60 * time.Second
	case Granularity5M:
		return 300 * time.Second
	case Granularity15M:
		return 900 * time.Second
	case Granularity1H:
		return 3600 * time.Second
	case Granularity6H:
		return 21600 * time.Second
	case Granularity1D:
		return 86400 * time.Second
	default:
		return 0
	}
}

type BlockListStats struct {
	TotalEntries int `json:"totalEntries"`
}
