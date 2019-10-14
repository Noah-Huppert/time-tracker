package main

import (
	"net/http"
)

// HealthEndpoint returns an HTTP okay response indicating the
// server is running correctly
type HealthEndpoint struct{}

// Handle
func (e HealthEndpoint) Handle(ectx EndpointContext, r *http.Request) (Responder, error) {
	return JSONResponder{
		Data: map[string]interface{}{
			"ok": true,
		},
	}, nil
}
