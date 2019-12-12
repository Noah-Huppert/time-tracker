/*
Time tracker HTTP API.

If command line arguments are passed acts as a management interface. Use to
create users. Otherwise starts an HTTP server.

Requests pass data using URL parameters for all HTTP verbs.
JSON encoded bodies may be used for the POST, PUT, PATCH,
DELETE HTTP verbs.

Responses will always be JSON encoded.

See data-model.png for models.
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/theckman/go-securerandom"
)

func main() {
	// Initialization
	ctx, cancelCtx := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	go func() {
		<-sigs
		cancelCtx()
	}()

	// Load config
	cfg, err := NewConfig()
	if err != nil {
		log.Panic().
			Str("error", err.Error()).
			Msg("failed to load config")
	}

	// Connect to DB
	db, err := sqlx.Connect("postgres", fmt.Sprintf("dbname=%s user=%s "+
		"password=%s sslmode=%s", cfg.Db.Name, cfg.Db.User, cfg.Db.Password,
		cfg.Db.SSLMode))
	if err != nil {
		log.Panic().
			Str("error", err.Error()).
			Msg("failed to connect to database")
	}

	// Act as managment interface if command line args provided

	flag.Parse()
	mgmtCmd := flag.Arg(0)

	if len(mgmtCmd) > 0 {
		switch mgmtCmd {
		case "create-user":
			log.Debug().Msg("create-user")

			// TODO: Get username and fullname flag values
			// TODO: Generate secure random with: github.com/theckman/go-securerandom
			// TODO: Save in DB
		default:
			log.Panic().
				Str("management-command", mgmtCmd).
				Msg("Unknown management command")
		}

		log.Info().
			Str("management-command", mgmtCmd).
			Msg("Management command ran, exitting")
		return
	}

	// Start HTTP server
	handler := mux.NewRouter()
	wrapper := EndpointWrapper{
		Cfg: cfg,
		Db:  db,
	}
	handler.Handle("/api/v0/health", wrapper.Wrap(HealthEndpoint{}))

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
