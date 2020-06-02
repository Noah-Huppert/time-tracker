// API HTTP server and management tool
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/Noah-Huppert/goconf"
	"github.com/Noah-Huppert/gointerrupt"
	"github.com/Noah-Huppert/golog"
	"github.com/golang-migrate/migrate/v4"
	migratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Config for API
type Config struct {
	// HTTPAddr is the address for the HTTP server
	HTTPAddr string `default:":8000" validate:"required"`

	// DB is database configuration
	DB struct {
		// Address of server
		Address string `default:"localhost:5432" validate:"required"`

		// Database
		Database string `default:"dev-time-tracker" validate:"required"`

		// User
		User string `default:"dev-time-tracker" validate:"required"`

		// Password
		Password string `default:"dev-time-tracker" validate:"required"`

		// Options are all the connection options formated as key=value
		// delimited by ampersands.
		// Example: one=value&two=value
		Options string `default:"sslmode=disable"`
	}
}

// DBURI returns a postgres URI with all the .DB parameters
func (c Config) DBURI(redact bool) string {
	optsStr := ""
	if len(c.DB.Options) > 0 {
		optsStr = fmt.Sprintf("?%s", c.DB.Options)
	}

	return fmt.Sprintf("postgres://%s:%s@%s/%s%s", c.Redact(redact, c.DB.User),
		c.Redact(redact, c.DB.Password),
		c.DB.Address, c.DB.Database, optsStr)
}

// Redact erases a sensative value if the redact arg is true, it ensures
// redaction keeps intacted if the value was empty or not for debugging.
// If the redact arg is false raw is returned.
func (c Config) Redact(redact bool, raw string) string {
	if redact {
		if len(raw) == 0 {
			return "REDACTED_EMPTY"
		} else {
			return "REDACTED_NOT_EMPTY"
		}
	} else {
		return raw
	}
}

// RouteHandler responds to HTTP endpoint requests
type RouteHandler interface {
	// Handle
	Handle(c RouteContext) Responder
}

// Responder responds to requests
type Responder interface {
	// Respond to request
	Respond(c RouteContext) error

	// FinishResponse returns if this responder should complete the process
	// of handling a request.
	FinishResponse() bool
}

// RouteContext is a custom echo.Context which contains custom fields
// for API routes
type RouteContext interface {
	// Request to route
	Request() *http.Request

	// Writer for route
	Writer() http.ResponseWriter

	// Log information
	Log() golog.Logger
}

// DefaultRouteContext is a default implementation of RouteContext
type DefaultRouteContext struct {
	// request
	request *http.Request

	// writer
	writer http.ResponseWriter

	// log
	log golog.Logger
}

// Request
func (c DefaultRouteContext) Request() *http.Request {
	return c.request
}

// Writer
func (c DefaultRouterContext) Writer() http.ResponseWriter {
	return c.writer
}

// Log
func (c DefaultRouterContext) Log() golog.Logger {
	return c.log
}

// StatusResponder writes a response status.
type StatusResponder struct {
	// Status to code
	Status int
}

// NewStatusResponder creates a StatusResponder.
func NewStatusResponder(status int) StatusResponder {
	return StatusResponder{
		Status: status,
	}
}

// Respond
func (r StatusResponder) Respond(c RouteContext) error {
	c.Writer().WriteHeader(r.Status)
}

// FinishResponse indicates a response is not finished, so a body can
// be written.
func (r StatusResponder) FinishResponse() bool {
	return false
}

// JSONResponder is a responder which writes data as JSON.
type JSONResponder struct {
	// Status
	Status StatusResponder

	// Data to respond
	Data interface{}
}

// NewJSONResponder creates a JSONResponder.
func NewJSONResponder(data interface{}) JSONResponder {
	return JSONResponder{
		Data: data,
	}
}

// Respond
func (r JSONResponder) Respond(c RouteContext) error {
	encoder := json.NewEncoder(c.Writer())

	if err := encoder.Encode(r.Data); err != nil {
		return fmt.ErrorF("failed to JSON encode: %s", err)
	}

	return nil
}

// FinishResponse ends a response so all data is JSON encoded.
func (r JSONRespondeR) FinishResponse() bool {
	return true
}

// AuthValid ensures a valid authentication token is present in the
// Authorization header.
func AuthValid(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c RouteContext) error {
		auth := c.Request().Header.Get("Authorization")
		if len(auth) == 0 {
			return echo.NewHTTPError(http.StatusUnauthorized,
				"no credentials provided")
		}

		return next(c)
	}
}

// RoutesWrapper wraps Routes in a Handler
type RouteWrapper struct {
	// Routes to wrap
	Routes []Route

	// Log
	Log golog.Logger
}

// ServeHTTP calls all routes in sequential order, stops the first Route which
// writes to the response.
func (r WrouteWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	routeCtx := DefaultRouteContext{
		request: r,
		writer:  w,
		log:     r.Log.GetChild(fmt.Sprintf("%s %s", r.Method, r.URL)),
	}

	for route := range r.Routes {
		resp := route.Handle(routeCtx)
		err := resp.Respond(routeCtx)
		if err != nil {
			routeCtx.Log().Errorf("responder failed: %s", err)

			errResp := NewJSONResponder(map[string]interface{}{
				"error": "internal server error",
			})

			if err := errResp.Respond(routeCtx); err != nil {
				routeCtx.Log().Errorf("failed to send internal server "+
					"error response after responder failed: %s", err)
			}
		}
	}
}

func main() {
	ctxPair := gointerrupt.NewCtxPair(context.Background())
	log := golog.NewLogger("api")

	// Load config
	cfgLdr := goconf.NewLoader()
	cfgLdr.AddConfigPath("/etc/time-tracker/*")
	var cfg Config
	if err := cfgLdr.Load(&cfg); err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", cfg.DBURI(false))
	if err != nil {
		log.Fatalf("failed to open db connection to \"%s\": %s",
			cfg.DBURI(true), err)
	}

	if err := db.PingContext(ctxPair.Graceful()); err != nil {
		log.Fatalf("failed to ping db \"%s\": %s", cfg.DBURI(true), err)
	}

	dbx, err := sqlx.NewDb(db, "postgres")
	if err != nil {
		log.Fatalf("failed to initialize db library: %s", err)
	}

	// Determine what to do
	cmdArg := ""
	if len(os.Args) >= 2 {
		cmdArg = os.Args[1]
	}

	switch cmdArg {
	case "server":
		var wg sync.WaitGroup

		// API routes
		e := echo.New()
		e.GET("/api/v0/health", func(c echo.Context) error {
			c.JSON(http.StatusOK, map[string]interface{}{
				"ok": true,
			})
			return nil
		})

		resCfg := map[string]ResourceRouteCfg{}
		e.Any("/api/v0/:resource", func(c echo.Context) error {

		})

		// Start API server
		log.Infof("starting HTTP server on %s", cfg.HTTPAddr)

		wg.Add(1)
		go func() {
			if err := e.Start(cfg.HTTPAddr); err != nil &&
				err != http.ErrServerClosed {
				log.Fatalf("failed to run HTTP server: %s", err)
			}
			wg.Done()
		}()

		// Gracefully shutdown server
		go func() {
			<-ctxPair.Graceful().Done()

			log.Info("shutting down HTTP server gracefully")

			if err := e.Shutdown(ctxPair.Harsh()); err != nil {
				log.Fatalf("failed to shutdown HTTP server: %s", err)
			}

			log.Info("shut down HTTP server")
		}()

		wg.Wait()
		break
	case "migrate":
		// Setup migrator
		dbDriver, err := migratePostgres.WithInstance(db,
			&migratePostgres.Config{})
		if err != nil {
			log.Fatalf("failed to create migrate database driver: %s", err)
		}

		migrator, err := migrate.NewWithDatabaseInstance(
			"file://db-migrations", "postgres", dbDriver)
		if err != nil {
			log.Fatalf("failed to create migrator: %s", err)
		}

		// Get pre migration status
		version, dirty, err := migrator.Version()
		if err != nil {
			log.Fatalf("failed to get pre migration status: %s", err)
		}

		log.Debugf("pre migration status: version=%d, dirty=%t", version,
			dirty)

		// Run migrations
		log.Infof("running database migrations")

		if err := migrator.Up(); err != nil {
			log.Fatalf("failed to migrate database: %s", err)
		}

		// Get post migration status

		version, dirty, err = migrator.Version()
		if err != nil {
			log.Fatalf("failed to get post migration status: %s", err)
		}

		log.Debugf("post migration status: version=%d, dirty=%t", version,
			dirty)
		break
	default:
		log.Fatalf("command must be \"server\" or \"migrate\"")
		break
	}
}
