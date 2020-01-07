// API HTTP server and management tool
package main

import (
	"context"
	"database/sql"
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
	"github.com/labstack/echo"
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
		e.GET("/health", func(c echo.Context) error {
			c.JSON(http.StatusOK, map[string]interface{}{
				"ok": true,
			})
			return nil
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
			if err := e.Shutdown(ctxPair.Harsh()); err != nil {
				log.Fatalf("failed to shutdown HTTP server: %s", err)
			}
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
