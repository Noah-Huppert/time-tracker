package main

import (
	"github.com/Noah-Huppert/goconf"
)

// Config is API configuration
type Config struct {
	// Db is database configuration
	Db struct {
		// Host
		Host string `validate:"required"`

		// SSLMode indicates if SSL should be used when connecting
		// See table 31-1: https://www.postgresql.org/docs/9.1/libpq-ssl.html
		SSLMode string `validate:"required" default:"verify-full"`

		// Name
		Name string `validate:"required"`

		// User name
		User string `validate:"required"`

		// Password
		Password string `validate:"required"`
	}
}

// NewConfig loads configuration from TOML files in $PWD and /etc/time-tracker
func NewConfig() (Config, error) {
	loader := goconf.NewDefaultLoader()

	loader.AddConfigPath("./*.toml")
	loader.AddConfigPath("/etc/time-tracker/*.toml")

	cfg := Config{}
	err := loader.Load(&cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
