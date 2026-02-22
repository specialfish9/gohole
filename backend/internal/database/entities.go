package database

import "time"

type Query struct {
	Name string `json:"name"`
	// Type is yet to be used.
	Type      uint16 `json:"type"`
	Blocked   bool   `json:"blocked"`
	Host      string `json:"host"`
	Timestamp int64  `json:"timestamp"`
	Millis    int64  `json:"millis"`
}

func NewQuery(name string, host string, blocked bool, millis int64) Query {
	return Query{
		Name:      name,
		Blocked:   blocked,
		Host:      host,
		Timestamp: time.Now().UTC().Unix(),
		Millis:    millis,
	}
}

type HostStat struct {
	Host         string  `json:"host"`
	QueryCount   uint64  `json:"queryCount"`
	BlockedCount uint64  `json:"blockedCount"`
	BlockRate    float64 `json:"blockRate"`
}

type DomainStats struct {
	Total        uint64 `json:"total"`
	BlockedCount uint64 `json:"blocked"`
}

type TopDomain struct {
	Domain string `json:"domain"`
	Count  uint64 `json:"count"`
}
