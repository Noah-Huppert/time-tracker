package main

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog/log"
)

// PanicHandler recovers from a panic and logs the details
type PanicHandler struct {
	// Handler to wrap
	Handler http.Handler
}

// ServeHTTP recovers from any panics which occur in .Handler
func (h PanicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if recovery := recover(); recovery != nil {
			log.Fatal().
				Str("method", r.Method).
				Str("path", r.URL.String()).
				Str("stack", string(debug.Stack())).
				Str("recovery", fmt.Sprintf("%#v", recovery))
		}
	}()

	h.Handler.ServeHTTP(w, r)
}
