package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	golangLog "log"
	"os"
	"path/filepath"
	"sort"
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
const TimeFormat = "2006-01-02 15:04:05"

// PeriodWeekly is a week duration
const PeriodWeekly = time.Hour * 24

// PeriodBiWeekly is a 2 week duration
const PeriodBiWeekly = PeriodWeekly * 2

// PeriodMonthly is a month duration
const PeriodMonthly = PeriodBiWeekly * 2

// PeriodWeeklyStr is the value a user must type to refer to PeriodWeekly
const PeriodWeeklyStr = "weekly"

// PeriodBiWeeklyStr is the value a user must type to refer to PeriodBiWeekly
const PeriodBiWeeklyStr = "bi-weekly"

// PeriodMonthlyStr is the value a user must type to refer to PeriodMonthly
const PeriodMonthlyStr = "monthly"

// ValidPeriodStrs is a list of all valid values a user could type to refer to a period
var ValidPeriodStrs = []string{
	PeriodWeeklyStr,
	PeriodBiWeeklyStr,
	PeriodMonthlyStr,
}

// ValidPeriodStrsJoined is a user friendly string listing values from ValidPeriodStrs
var ValidPeriodStrJoined = strings.Join(ValidPeriodStrs, ", ")

// ParsePeriod takes a str and converts it to a duration. See Period... constants
func ParsePeriod(str string) (time.Duration, error) {
	if str == PeriodWeeklyStr {
		return PeriodWeekly, nil
	} else if str == PeriodBiWeeklyStr {
		return PeriodBiWeekly, nil
	} else if str == PeriodMonthlyStr {
		return PeriodMonthly, nil
	}

	return time.Hour * 0, fmt.Errorf("invalid period '%s', valid values: %s", str, ValidPeriodStrJoined)
}

func main() {
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

	// Parse options
	var inDir string
	flag.StringVar(&inDir, "in-dir", "times", "Directory containing time tracking CSV files")

	var outDir string
	flag.StringVar(&outDir, "out-dir", "reports", "Directory where reports for each billing period will be written")

	var billingPeriod string
	flag.StringVar(&billingPeriod, "billing-period", "bi-weekly", fmt.Sprintf("How often bills will be issued, valid values: %s", ValidPeriodStrJoined))

	var columnStartTime string
	flag.StringVar(&columnStartTime, "column-start-time", "time started", "Name of column which contains start time in format (24 hour time): YYYY-MM-DD HH:MM:SS")

	var columnEndTime string
	flag.StringVar(&columnEndTime, "column-end-time", "time ended", "Name of column which contains end time in format (24 hour time): YYYY-MM-DD HH:MM:SS")

	var columnComment string
	flag.StringVar(&columnComment, "column-comment", "comment", "Name of column which contains an optional description of what happened during the time period")

	flag.Parse()

	// Convert options into programmatic values
	billingPeriodDuration, err := ParsePeriod(billingPeriod)
	if err != nil {
		logger.Fatal("failed to parse billing-period option", zap.Error(err))
	}

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

	timeEntriesList := []TimeEntry{}
	for _, entry := range timeEntries {
		timeEntriesList = append(timeEntriesList, entry)
	}

	sort.Slice(timeEntriesList, func(i, j int) bool {
		return timeEntriesList[i].StartTime.Before(timeEntriesList[j].StartTime)
	})

	if len(timeEntriesList) == 0 {
		logger.Fatal("no time entries")
	}

	// Roll-up into reports
	firstStartDate := timeEntriesList[0].StartTime

	periodStart := time.Date(firstStartDate.Year(), firstStartDate.Month(), 1, 0, 0, 0, 0, firstStartDate.Location())
	periodEnd := periodStart.Add(billingPeriodDuration)

	billingPeriods := [][]time.Time{
		{periodStart, periodEnd},
	}
	timeEntriesByPeriodStart := make(map[time.Time][]TimeEntry)
	timeEntriesByPeriodStart[periodStart] = []TimeEntry{}

	for _, entry := range timeEntriesList {
		// Check if starting a new period
		if entry.StartTime.After(periodEnd) {
			periodStart = periodStart.Add(billingPeriodDuration)
			periodEnd = periodEnd.Add(billingPeriodDuration)

			billingPeriods = append(billingPeriods, []time.Time{periodStart, periodEnd})
			timeEntriesByPeriodStart[periodStart] = []TimeEntry{}
		}

		// Add time entry
		timeEntriesByPeriodStart[periodStart] = append(timeEntriesByPeriodStart[periodStart], entry)
	}
}
