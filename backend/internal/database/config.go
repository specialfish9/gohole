package database

import "github.com/specialfish9/confuso/v2"

type Type = string

const (
	TypeClickHouse Type = "clickhouse"
	TypePostgres   Type = "postgres"
	TypeNone       Type = "none"
)

type Config struct {
	// Type is the type of the database (e.g., "clickhouse", "postgres"). Use "none" to disable database storage.
	Type Type `confuso:"type" validate:"required,oneof=clickhouse postgres none"`
	// Address is the address of the database.
	Address string `confuso:"address" validate:"required"`
	// User is the username for the database.
	User string `confuso:"user" validate:"required"`
	// Password is the password for the database.
	Password string `confuso:"password" validate:"required"`
	// Name is the name of the database.
	Name string `confuso:"name" validate:"required"`
	// Debug indicates whether to enable debug mode for the database (e.g., logging queries).
	// Default is false.
	Debug confuso.Optional[bool] `confuso:"debug"`
}
