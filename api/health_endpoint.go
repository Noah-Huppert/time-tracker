package main

import (
	"net/http"
)

// HealthEndpoint returns an HTTP okay response indicating the
// server is running correctly
type HealthEndpoint struct{}

// Handle
func (e HealthEndpoint) Handle(r *http.Request) http.Handler {
	return JSONResponder{
		Data: map[string]interface{}{
			"ok": true,
		},
	}
}
