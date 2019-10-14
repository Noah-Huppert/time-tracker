package main

import (
	"github.com/rs/zerolog"
)

// EndpointContext holds objects which most endpoints and responders need
type EndpointContext struct {
	// Log outputs information to a log file
	Log zerolog.Logger

	// Cfg is application config
	Cfg Config
}
