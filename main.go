package main

import (
	"encoding/csv"
	"flag"
	golangLog "log"
	"os"
	"strings"

	"go.uber.org/zap"
)

func main() {
	// Parse options
	var inDir string
	flag.StringVar(&inDir, "in-dir", "times", "Directory containing time tracking CSV files")

	var outDir string
	flag.StringVar(&outDir, "out-dir", "reports", "Directory where reports for each billing period will be written")

	var billingPeriod string
	flag.StringVar(&billingPeriod, "billing-period", "bi-weekly", "How often bills will be issued, valid values: bi-weekly")

	var columnStartTime string
	flag.StringVar(&columnStartTime, "column-start-time", "time started", "Name of column which contains start time in format (24 hour time): YYYY-MM-DD HH:MM:SS")

	var columnEndTime string
	flag.StringVar(&columnEndTime, "column-end-time", "time ended", "Name of column which contains end time in format (24 hour time): YYYY-MM-DD HH:MM:SS")

	flag.Parse()

	// Setup logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		golangLog.Fatalf("failed to initialize logger: %s\n", err)
	}

	defer func() {
		if err := logger.Sync(); err != nil && !strings.Contains(err.Error(), "handle is invalid") {
			golangLog.Fatalf("failed flush logger: %s\n", err)
		}
	}()

	// Read in data
	inFileNames, err := os.ReadDir(inDir)
	if err != nil {
		logger.Fatal("failed to read input directory", zap.Error(err))
	}
	for _, fileEntry := range inFileNames {
		f, err := os.Open(fileEntry.Name())
		if err != nil {
			logger.Fatal("failed to open input file", zap.Error(err))
		}

		reader := csv.NewReader(f)
		headers, err := reader.Read()
		if err != nil {
			logger.Fatal("failed to read CSV headers", zap.Error(err))
		}
	}
}
