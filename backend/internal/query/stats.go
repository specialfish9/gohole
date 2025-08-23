package query

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

func (i Interval) ToDuration() int64 {
	switch i {
	case Interval1H:
		return 3600
	case Interval6H:
		return 21600
	case Interval1D:
		return 86400
	case Interval7D:
		return 604800
	case Interval30D:
		return 2592000
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

// ToDuration converts the granularity to its equivalent duration in seconds.
func (g Granularity) ToDuration() int64 {
	switch g {
	case Granularity1M:
		return 60
	case Granularity5M:
		return 300
	case Granularity15M:
		return 900
	case Granularity1H:
		return 3600
	case Granularity6H:
		return 21600
	case Granularity1D:
		return 86400
	default:
		return 0
	}
}
