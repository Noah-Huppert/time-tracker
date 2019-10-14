package main

import (
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Endpoint handles HTTP requests
type Endpoint interface {
	// Handle HTTP request and return a http.Handler which will
	// write a response. If an error is returned an
	// http.StatusInternalServerError response will be sent and the error will
	// be logged.
	Handle(ectx EndpointContext, r *http.Request) (Responder, error)
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
	// Build EndpointContext
	ectx := EndpointContext{
		Log: log.With().
			Str("method", r.Method).
			Str("path", r.URL.String()).
			Logger(),
	}

	// Handle request
	ectx.Log.Info().Send()
	resp, err := h.endpoint.Handle(ectx, r)

	// If error handling request
	if err != nil {
		ectx.Log.Error().
			Str("error", err.Error()).
			Msg("error handling request")

		resp = JSONResponder{
			Status: http.StatusInternalServerError,
			Data: map[string]interface{}{
				"error": "internal server error",
			},
		}
	}

	// Respond to request
	err = resp.Respond(ectx, w, r)

	// If error responding to request
	if err != nil {
		ectx.Log.Error().
			Str("error", err.Error()).
			Msg("error responding to request")

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")

		_, err = io.WriteString(w, "{\"error\": \"internal server error\"}")
		if err != nil {
			ectx.Log.Fatal().
				Str("error", err.Error()).
				Msg("returned Responder encountered an error " +
					", then an attempt to return a hardcoded error " +
					"response failed")
		}
	}
}
