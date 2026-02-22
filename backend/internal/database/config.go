package database

import "github.com/specialfish9/confuso/v2"

type Config struct {
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
