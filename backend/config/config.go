package config

import (
	"fmt"
	"gohole/internal/filter"

	"github.com/go-playground/validator/v10"
	"github.com/specialfish9/confuso/v2"
)

type Config struct {
	App struct {
		// LogLevel is the level of logging (e.g., "debug", "info", "warn", "error").
		LogLevel string `confuso:"log_level" validate:"required"`
	} `confuso:"app"`

	HTTP struct {
		// Address is the address on which the HTTP server will listen for incoming requests.
		Address string `confuso:"address" validate:"required"`
		// ServeFrontend indicates whether to serve the frontend or not.
		ServeFrontend confuso.Optional[bool] `confuso:"serve_frontend"`
	} `confuso:"http"`

	DNS struct {
		// Upstream is the address of the upstream DNS server to which queries will be forwarded.
		Upstream string `confuso:"upstream" validate:"required"`
		// Address is the address on which the DNS server will listen for incoming queries.
		Address string `confuso:"address" validate:"required"`
	} `confuso:"dns"`

	DB struct {
		// Address is the address of the database.
		Address string `confuso:"address" validate:"required"`
		// User is the username for the database.
		User string `confuso:"user" validate:"required"`
		// Password is the password for the database.
		Password string `confuso:"password" validate:"required"`
		// Name is the name of the database.
		Name string `confuso:"name" validate:"required"`
	} `confuso:"db"`

	Blocking struct {
		// FilterStrategy is the strategy used to filter domains (e.g., "basic", "trie2").
		FilterStrategy filter.Strategy `confuso:"filter_strategy" validate:"required,oneof=basic trie trie2"`
		// BlocklistFile is the path to the file containing the list of blocklists URLs.
		BlocklistFile string `confuso:"blocklist_file" validate:"required"`
		// LocalBlockList is the path to a local file containing a list of domains to block.
		LocalBlockList confuso.Optional[string] `confuso:"local_blocklist"`
		// LocalAllowList is the path to a local file containing a list of domains to allow.
		LocalAllowList confuso.Optional[string] `confuso:"local_allowlist"`
	} `confuso:"blocking"`
}

func New(fileName string) (*Config, error) {
	var config Config

	if err := confuso.Do(fileName, &config); err != nil {
		return nil, fmt.Errorf("config: loading config file: %w", err)
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(&config); err != nil {
		return nil, fmt.Errorf("config: validating config: %w", err)
	}

	return &config, nil
}
