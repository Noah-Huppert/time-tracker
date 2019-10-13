package main

import (
	"net/http"
)

// Endpoint handles HTTP requests
type Endpoint interface {
	// Handle HTTP request and return a http.Handler which will
	// write a response
	Handle(r *http.Request) http.Handler
}

// WrapEndpoint wraps a Endpoint to make it a http.Handler
func WrapEndpoint(endpoint Endpoint) http.Handler {
	return endpointHandler{
		endpoint: endpoint,
	}
}

// endpointHandler is used by WrapEndpoint to run a Endpoint as
// an http.Handler
type endpointHandler struct {
	endpoint Endpoint
}

// ServeHTTP runs a Endpoint
func (h endpointHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := h.endpoint.Handle(r)
	resp.ServeHTTP(w, r)
}
