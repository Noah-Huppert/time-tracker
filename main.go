// API HTTP server and management tool
package main

import (
	"database/sql"
	"os"

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

	// DBURI is the database URI
	DBURI string `default:"postgres://localhost:5432/dev-time-tracker" validate:"required"`
}

func main() {
	ctx, _ := gointerrupt.NewCtx()
	log := golog.NewLogger("api")

	// Load config
	cfgLdr := goconf.NewLoader()
	cfgLdr.AddConfigPath("/etc/time-tracker/*")
	var cfg Config
	if err := cfgLdr.Load(&cfg); err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", cfg.DBURI)
	if err != nil {
		log.Fatalf("failed to open db connection: %s", err)
	}

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping db: %s", err)
	}

	// Determine what to do
	switch os.Args[0] {
	case "server":
		// API server
		e := echo.New()
		log.Infof("starting HTTP server on %s", cfg.HTTPAddr)

		if err := e.Start(cfg.HTTPAddr); err != nil {
			log.Fatalf("failed to run HTTP server")
		}
		break
	case "migrate":
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

		if err := migrator.Up(); err != nil {
			log.Fatalf("failed to migrate database: %s", err)
		}
		break
	default:
		log.Fatalf("command must be \"server\" or \"migrate\"")
		break
	}
}
