package main

import (
	"context"
	golangLog "log"
	"strings"
	"sync"

	"github.com/Noah-Huppert/gointerrupt"
	"github.com/Noah-Huppert/time-tracker/api/config"
	"github.com/Noah-Huppert/time-tracker/api/models"
	"github.com/Noah-Huppert/time-tracker/api/server"
	"go.uber.org/zap"
)

func main() {
	ctxPair := gointerrupt.NewCtxPair(context.Background())
	var wg sync.WaitGroup

	// Setup logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		golangLog.Fatalf("failed to initialize logger: %s\n", err)
	}

	go func() {
		wg.Add(1)
		<-ctxPair.Graceful().Done()
		if err := logger.Sync(); err != nil && !strings.Contains(err.Error(), "handle is invalid") {
			golangLog.Fatalf("failed flush logger: %s\n", err)
		}
	}()

	// Load configuration
	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal("failed to load configuration", zap.Error(err))
	}
	logger.Debug("loaded configuration")

	// Start server
	server := server.NewServer(server.NewServerOpts{
		Logger: logger.With(zap.String("component", "api")),
		TimeEntryRepo: models.NewCSVTimeEntryRepo(models.NewCSVTimeEntryRepoOpts{
			InDir:           "./data/times",
			Timezone:        "EST",
			ColumnStartTime: "time started",
			ColumnEndTime:   "time ended",
			ColumnComment:   "comment",
		}),
		InvoiceSettingsRepo: models.NewJSONInvoiceSettingsRepo(models.NewJSONInvoiceSettingsRepoOpts{
			FilePath: "./data/invoice-settings.json",
		}),
	})

	if err := server.Listen(ctxPair, cfg.HTTPListen); err != nil {
		logger.Fatal("failed to run HTTP server", zap.Error(err))
	}

	wg.Done()
}
