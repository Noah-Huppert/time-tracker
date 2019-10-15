package main

import (
	"io"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

// EndpointWrapper turns Endpoints into http.Handlers
type EndpointWrapper struct {
	// Cfg is application config
	Cfg Config

	// Db is a database connection instance
	Db *sqlx.DB
}

// Wrap an Endpoint to make it a http.Handler
func (wrapper EndpointWrapper) Wrap(endpoint Endpoint) http.Handler {
	return endpointHandler{
		wrapper:  wrapper,
		endpoint: endpoint,
	}
}

// endpointHandler is used by WrapEndpoint to run a Endpoint as
// an http.Handler
type endpointHandler struct {
	// wrapper fields are used to construct an EndpointContext for use by
	// Endpoints and Responders.
	wrapper EndpointWrapper

	// endpoint run to handle requests in ServeHTTP
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
		Cfg: h.wrapper.Cfg,
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
