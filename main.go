package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	golangLog "log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"go.uber.org/zap"
)

// TimeEntry records a period of time when work was completed
type TimeEntry struct {
	/// StartTime is the date and time when the period started
	StartTime time.Time

	// EndTime is the date and time when the period ended
	EndTime time.Time

	// Command is an optional comment explaining what work was completed during the period
	Comment string
}

// Hash returns a checksum of the contents of TimeEntry, can be used to find duplicate TimeEntry structs
func (e TimeEntry) Hash() (string, error) {
	hash, err := hashstructure.Hash(e, hashstructure.FormatV2, nil)
	if err != nil {
		return "", fmt.Errorf("failed to hash structure: %s", err)
	}

	return fmt.Sprintf("%d", hash), nil
}

// TimeFormat represents YYYY-MM-DD HH:MM:SS
const TimeFormat = "2006-1-2 15:04:05"

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

	var columnComment string
	flag.StringVar(&columnComment, "column-comment", "comment", "Name of column which contains an optional description of what happened during the time period")

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
		logger.Fatal("failed to read input directory", zap.Error(err), zap.String("in-dir", inDir))
	}

	timeEntries := make(map[string]TimeEntry) // keys are hashes of the values
	for _, fileEntry := range inFileNames {
		filePath := filepath.Join(inDir, fileEntry.Name())
		// Open file
		f, err := os.Open(filePath)
		if err != nil {
			logger.Fatal("failed to open input file", zap.Error(err), zap.String("file", filePath))
		}

		reader := csv.NewReader(f)

		// Read headers
		headers, err := reader.Read()
		if err != nil {
			logger.Fatal("failed to read CSV headers", zap.Error(err), zap.String("file", filePath))
		}
		headerMap := make(map[string]int)
		for i, key := range headers {
			headerMap[key] = i
		}

		// Check for required columns
		if _, ok := headerMap[columnStartTime]; !ok {
			logger.Fatal("start time column not found in CSV file", zap.String("file", filePath), zap.String("column-start-time", columnStartTime))
		}

		if _, ok := headerMap[columnEndTime]; !ok {
			logger.Fatal("end time column not found in CSV file", zap.String("file", filePath), zap.String("column-end-time", columnEndTime))
		}

		if _, ok := headerMap[columnComment]; !ok {
			logger.Fatal("comment column not found in CSV file", zap.String("file", filePath), zap.String("column-comment", columnComment))
		}

		// Parse rows into TimeEntry structs
		rows, err := reader.ReadAll()
		if err != nil {
			logger.Fatal("failed to read rows of CSV", zap.String("file", filePath), zap.Error(err))
		}

		for rowI, row := range rows {
			// Parse date times
			startTimeStr := row[headerMap[columnStartTime]]
			startTime, err := time.Parse(TimeFormat, startTimeStr)
			if err != nil {
				logger.Fatal("failed to parse start time", zap.String("file", filePath), zap.Int("row", rowI), zap.Error(err), zap.String("raw start time", startTimeStr))
			}

			endTimeStr := row[headerMap[columnEndTime]]
			endTime, err := time.Parse(TimeFormat, endTimeStr)
			if err != nil {
				logger.Fatal("failed to parse end time", zap.String("file", filePath), zap.Int("row", rowI), zap.Error(err), zap.String("raw end time", endTimeStr))
			}

			// Create entry
			entry := TimeEntry{
				StartTime: startTime,
				EndTime:   endTime,
				Comment:   row[headerMap[columnComment]],
			}
			entryHash, err := entry.Hash()
			if err != nil {
				logger.Fatal("failed to hash time entry", zap.String("file", filePath), zap.Int("row", rowI), zap.Error(err))
			}
			timeEntries[entryHash] = entry
		}
	}

	logger.Debug("time entries", zap.String("timeEntries", fmt.Sprintf("%#v", timeEntries)))
}
