package database

import "time"

type Query struct {
	Name      string `json:"name"`
	Type      uint16 `json:"type"`
	Blocked   bool   `json:"blocked"`
	Timestamp int64  `json:"timestamp"`
}

func NewQuery(name string, qtype uint16, blocked bool) Query {
	return Query{
		Name:      name,
		Type:      qtype,
		Blocked:   blocked,
		Timestamp: time.Now().Unix(),
	}
}
