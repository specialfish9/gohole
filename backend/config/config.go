package config

import (
	"fmt"
	"gohole/internal/controller/dns"
	"gohole/internal/controller/http"
	"gohole/internal/database"
	"gohole/internal/filter"

	"github.com/go-playground/validator/v10"
	"github.com/specialfish9/confuso/v2"
)

type Config struct {
	App struct {
		// LogLevel is the level of logging (e.g., "debug", "info", "warn", "error").
		LogLevel LogLevel `confuso:"log_level" validate:"required"`
	} `confuso:"app"`

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

	HTTP http.Config `confuso:"http"`

	DNS dns.Config `confuso:"dns"`

	DB database.Config `confuso:"db"`
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
