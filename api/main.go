/*
Time tracker HTTP API.

Requests pass data using URL parameters for all HTTP verbs.
JSON encoded bodies may be used for the POST, PUT, PATCH,
DELETE HTTP verbs.

Responses will always be JSON encoded.

RESTful API. With resources: users, auth tokens, clients, projects, time entry.
*/
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/gorilla/mux"
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
		Addr:    ":8000",
		Handler: handler,
	}
	wg.Add(1)

	go func() {
		log.Printf("starting HTTP server")

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start HTTP server: %s", err.Error())
		}
	}()

	go func() {
		<-ctx.Done()

		log.Printf("stopping HTTP server")

		err := server.Shutdown(context.Background())
		if err != nil {
			log.Fatalf("failed to shut down HTTP server: %s", err.Error())
		}

		wg.Done()
	}()

	wg.Wait()

	log.Printf("exiting")
}
