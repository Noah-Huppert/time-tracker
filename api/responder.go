package main

import (
	"net/http"
)

// Responder responds to an HTTP request
type Responder interface {
	// Respond to request
	Respond(ectx EndpointContext, w http.ResponseWriter, r *http.Request) error
}
