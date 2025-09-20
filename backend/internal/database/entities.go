package database

import "time"

type Query struct {
	Name      string `json:"name"`
	Type      uint16 `json:"type"`
	Blocked   bool   `json:"blocked"`
	Host      string `json:"host"`
	Timestamp int64  `json:"timestamp"`
}

func NewQuery(name string, qtype uint16, host string, blocked bool) Query {
	return Query{
		Name:      name,
		Type:      qtype,
		Blocked:   blocked,
		Host:      host,
		Timestamp: time.Now().UTC().Unix(),
	}
}

type HostStat struct {
	Host         string  `json:"host"`
	QueryCount   uint64  `json:"queryCount"`
	BlockedCount uint64  `json:"blockedCount"`
	BlockRate    float64 `json:"blockRate"`
}
