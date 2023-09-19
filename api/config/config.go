package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config is the API server configuration
type Config struct {
	// HTTPListen is the address on which to listen for HTTP traffic
	HTTPListen string `default:":4000"`
}

// NewConfig parses configuration from the environment
func NewConfig() (*Config, error) {
	var config Config
	if err := envconfig.Process("TIME_TRACKER_API", &config); err != nil {
		return nil, fmt.Errorf("failed to parse config from environment: %s", err)
	}

	return &config, nil
}
