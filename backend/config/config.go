package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/specialfish9/confuso"
)

type Config struct {
	Upstream      string `confuso:"upstream" validate:"required"`
	DNSAddress    string `confuso:"dns_address" validate:"required"`
	HTTPAddress   string `confuso:"http_address" validate:"required"`
	DBAddress     string `confuso:"db_address" validate:"required"`
	DBUser        string `confuso:"db_user" validate:"required"`
	DBPassword    string `confuso:"db_password" validate:"required"`
	DBName        string `confuso:"db_name" validate:"required"`
	BlocklistFile string `confuso:"blocklist_file" validate:"required"`
	LogLevel      string `confuso:"log_level" validate:"required"`
	ServeFrontend bool   `confuso:"serve_frontend"`
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
