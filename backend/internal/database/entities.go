package database

type Query struct {
	Name    string `json:"name"`
	Type    uint16 `json:"type"`
	Blocked bool   `json:"blocked"`
}

func NewQuery(name string, qtype uint16, blocked bool) Query {
	return Query{
		Name:    name,
		Type:    qtype,
		Blocked: blocked,
	}
}
