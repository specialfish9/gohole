package config

import (
	"fmt"
	"gohole/internal/filter"

	"github.com/go-playground/validator/v10"
	"github.com/specialfish9/confuso"
)

type Config struct {
	// Upstream is the address of the upstream DNS server to which queries will be forwarded.
	Upstream string `confuso:"upstream" validate:"required"`
	// DNSAddress is the address on which the DNS server will listen for incoming queries.
	DNSAddress string `confuso:"dns_address" validate:"required"`
	// FilterStrategy is the strategy used to filter domains (e.g., "basic", "trie2").
	FilterStrategy filter.Strategy `confuso:"filter_strategy" validate:"required,oneof=basic trie trie2"`
	// HTTPAddress is the address on which the HTTP server will listen for incoming requests.
	HTTPAddress string `confuso:"http_address" validate:"required"`
	// ServeFrontend indicates whether to serve the frontend or not.
	ServeFrontend bool `confuso:"serve_frontend"`
	// DBAddress is the address of the database.
	DBAddress string `confuso:"db_address" validate:"required"`
	// DBPort is the port of the database.
	DBUser string `confuso:"db_user" validate:"required"`
	// DBPassword is the password of the database.
	DBPassword string `confuso:"db_password" validate:"required"`
	// DBName is the name of the database.
	DBName string `confuso:"db_name" validate:"required"`
	// BlocklistFile is the path to the file containing the list of blocklists URLs.
	BlocklistFile string `confuso:"blocklist_file" validate:"required"`
	// LocalBlockList is the path to a local file containing a list of domains to block.
	LocalBlockList string `confuso:"local_blocklist"`
	// LocalAllowList is the path to a local file containing a list of domains to allow.
	LocalAllowList string `confuso:"local_allowlist"`
	// LogLevel is the level of logging (e.g., "debug", "info", "warn", "error").
	LogLevel string `confuso:"log_level" validate:"required"`
}

func New(fileName string) (*Config, error) {
	var config Config

	if err := confuso.LoadConf(fileName, &config); err != nil {
		return nil, fmt.Errorf("config: loading config file: %w", err)
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(&config); err != nil {
		return nil, fmt.Errorf("config: validating config: %w", err)
	}

	return &config, nil
}
