package database

type Query struct {
	Name    string
	Type    uint16
	Blocked bool
}

func NewQuery(name string, qtype uint16, blocked bool) Query {
	return Query{
		Name:    name,
		Type:    qtype,
		Blocked: blocked,
	}
}
