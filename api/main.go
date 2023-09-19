package main

import (
	"context"
	"encoding/csv"
	"fmt"
	golangLog "log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Noah-Huppert/gointerrupt"
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

func (e TimeEntry) Duration() time.Duration {
	return e.EndTime.Sub(e.StartTime)
}

type BillingPeriod struct {
	StartTime time.Time
	EndTime   time.Time
	Entries   []TimeEntry
}

func (p BillingPeriod) Duration() time.Duration {
	var total time.Duration = 0

	for _, entry := range p.Entries {
		total += entry.Duration()
	}

	return total
}

// InputTimeFormat represents YYYY-MM-DD HH:MM:SS
const InputTimeFormat = "2006-01-02 15:04:05 MST"

type ReadInTimeEntriesOpts struct {
	InDir           string
	Timezone        string
	ColumnStartTime string
	ColumnEndTime   string
	ColumnComment   string
}

func ReadInTimeEntries(opts ReadInTimeEntriesOpts) ([]TimeEntry, error) {
	// Read in data
	inFileNames, err := os.ReadDir(opts.InDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read input directory: %s", err)
	}

	timeEntries := make(map[string]TimeEntry) // keys are hashes of the values
	for _, fileEntry := range inFileNames {
		filePath := filepath.Join(opts.InDir, fileEntry.Name())
		// Open file
		f, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open input file: %s", err)
		}

		reader := csv.NewReader(f)

		// Read headers
		headers, err := reader.Read()
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV headers: %s", err)
		}
		headerMap := make(map[string]int)
		for i, key := range headers {
			headerMap[key] = i
		}

		// Check for required columns
		for _, requiredCol := range []string{
			opts.ColumnStartTime,
			opts.ColumnEndTime,
			opts.ColumnComment,
		} {
			if _, ok := headerMap[requiredCol]; !ok {
				return nil, fmt.Errorf("missing column '%s' from '%s' file", requiredCol, fileEntry)
			}
		}

		// Parse rows into TimeEntry structs
		rows, err := reader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("failed to read rows of CSV: %s", err)
		}

		for rowI, row := range rows {
			// Parse date times
			startTimeStr := fmt.Sprintf("%s %s", row[headerMap[opts.ColumnStartTime]], opts.Timezone)
			startTime, err := time.Parse(InputTimeFormat, startTimeStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse start time '%s' in row %d: %s", startTimeStr, rowI, err)
			}

			endTimeStr := fmt.Sprintf("%s %s", row[headerMap[opts.ColumnEndTime]], opts.Timezone)
			endTime, err := time.Parse(InputTimeFormat, endTimeStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse end time '%s' in row %d: %s", endTimeStr, rowI, err)
			}

			// Create entry
			entry := TimeEntry{
				StartTime: startTime,
				EndTime:   endTime,
				Comment:   row[headerMap[opts.ColumnComment]],
			}
			entryHash, err := entry.Hash()
			if err != nil {
				return nil, fmt.Errorf("failed to hash time entry: %s", err)
			}
			timeEntries[entryHash] = entry
		}
	}

	// Sort
	timeEntriesList := []TimeEntry{}
	for _, entry := range timeEntries {
		timeEntriesList = append(timeEntriesList, entry)
	}

	sort.Slice(timeEntriesList, func(i, j int) bool {
		return timeEntriesList[i].StartTime.Before(timeEntriesList[j].StartTime)
	})

	// Done
	return timeEntriesList, nil
}

func main() {
	ctxPair := gointerrupt.NewCtxPair(context.Background())
	var wg sync.WaitGroup

	// Setup logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		golangLog.Fatalf("failed to initialize logger: %s\n", err)
	}

	func() {
		wg.Add(1)
		<-ctxPair.Graceful().Done()
		if err := logger.Sync(); err != nil && !strings.Contains(err.Error(), "handle is invalid") {
			golangLog.Fatalf("failed flush logger: %s\n", err)
		}
	}()

	wg.Done()
}
