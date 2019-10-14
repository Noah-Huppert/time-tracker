/*
Time tracker HTTP API.

Requests pass data using URL parameters for all HTTP verbs.
JSON encoded bodies may be used for the POST, PUT, PATCH,
DELETE HTTP verbs.

Responses will always be JSON encoded.

See data-model.png for models.
*/
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func main() {
	// {{{1 Initialization
	ctx, cancelCtx := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	go func() {
		<-sigs
		cancelCtx()
	}()

	// {{{1 Start HTTP server
	handler := mux.NewRouter()
	handler.Handle("/api/v0/health", WrapEndpoint(HealthEndpoint{}))

	server := http.Server{
		Addr: ":8000",
		Handler: PanicHandler{
			Handler: handler,
		},
	}
	wg.Add(1)

	go func() {
		log.Info().Msg("starting HTTP server")

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error().
				Str("error", err.Error()).
				Msg("failed to start HTTP server")
		}
	}()

	go func() {
		<-ctx.Done()

		log.Info().Msg("stopping HTTP server")

		err := server.Shutdown(context.Background())
		if err != nil {
			log.Error().
				Str("error", err.Error()).
				Msg("failed to shut down HTTP server")
		}

		wg.Done()
	}()

	wg.Wait()

	log.Info().Msg("exiting")
}
