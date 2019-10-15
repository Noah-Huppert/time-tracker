package main

import (
	"net/http"
)

// Endpoint handles HTTP requests
type Endpoint interface {
	// Handle HTTP request and return a http.Handler which will
	// write a response. If an error is returned an
	// http.StatusInternalServerError response will be sent and the error will
	// be logged.
	Handle(ectx EndpointContext, r *http.Request) (Responder, error)
}
