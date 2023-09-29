package models

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/mitchellh/hashstructure/v2"
)

// TimeEntry records a period of time when work was completed
type TimeEntry struct {
	/// StartTime is the date and time when the period started
	StartTime time.Time `json:"start_time"`

	// EndTime is the date and time when the period ended
	EndTime time.Time `json:"end_time"`

	// Command is an optional comment explaining what work was completed during the period
	Comment string `json:"comment"`
}

// Hash returns a checksum of the contents of TimeEntry, can be used to find duplicate TimeEntry structs
func (e TimeEntry) Hash() (string, error) {
	hash, err := hashstructure.Hash(e, hashstructure.FormatV2, nil)
	if err != nil {
		return "", fmt.Errorf("failed to hash structure: %s", err)
	}

	return fmt.Sprintf("%d", hash), nil
}

// Duration of the time entry
func (e TimeEntry) Duration() time.Duration {
	return e.EndTime.Sub(e.StartTime)
}

// TimeEntryRepo are functions to query and retrieve TimeEntries
type TimeEntryRepo interface {
	// List time entries sorted earliest to latest
	List(opts ListTimeEntriesOpts) ([]TimeEntry, error)
}

type ListTimeEntriesOpts struct {
	// StartDate indicates only time entries which started after (inclusive) this date should be returned
	StartDate *time.Time

	// EndDate indicates only time entries which started before (inclusive) this date should be returned
	EndDate *time.Time
}

// CSVTimeEntryRepo implements TimeEntryRepo by loading CSV files from a directory
type CSVTimeEntryRepo struct {
	// inDir is the directory in which CSV files will be located
	inDir string

	// timezone in which times are in
	timezone string

	// columnStartTime is the name of the start time column
	columnStartTime string

	// columnEndTime is the name of the end time column
	columnEndTime string

	// columnComment is the name of the comment column
	columnComment string
}

// CSVInputTimeFormat represents YYYY-MM-DD HH:MM:SS
const CSVInputTimeFormat = "2006-01-02 15:04:05 MST"

// NewCSVTimeEntryRepoOpts are options for creating a CSVTimeEntryRepo
type NewCSVTimeEntryRepoOpts struct {
	// InDir is the input directory
	InDir string

	// Timezone in which times are in
	Timezone string

	// ColumnStartTime is the name of the start time column
	ColumnStartTime string

	// ColumnEndTime is the name of the end time column
	ColumnEndTime string

	// ColumnComment is the name of the comment column
	ColumnComment string
}

// NewCSVTimeEntryRepo creates a new CSVTimeEntryRepo
func NewCSVTimeEntryRepo(opts NewCSVTimeEntryRepoOpts) CSVTimeEntryRepo {
	return CSVTimeEntryRepo{
		inDir:           opts.InDir,
		timezone:        opts.Timezone,
		columnStartTime: opts.ColumnStartTime,
		columnEndTime:   opts.ColumnEndTime,
		columnComment:   opts.ColumnComment,
	}
}

// dateCompare only compares the date (year, month, day) component of a time. Returns less than 0 if compare is before base, greater than 0 if compare is after base, 0 if the same.
func dateCompare(base time.Time, compare time.Time) int32 {
	// Year
	if compare.Year() > base.Year() {
		return 1
	} else if compare.Year() < base.Year() {
		return -1
	}

	// Month
	if compare.Month() > base.Month() {
		return 1
	} else if compare.Month() < base.Month() {
		return -1
	}

	// Day
	if compare.Day() > base.Day() {
		return 1
	} else if compare.Day() < base.Day() {
		return -1
	}

	return 0
}

func (r CSVTimeEntryRepo) List(opts ListTimeEntriesOpts) ([]TimeEntry, error) {
	// Read in data
	inFileNames, err := os.ReadDir(r.inDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read input directory: %s", err)
	}

	timeEntries := make(map[string]TimeEntry) // keys are hashes of the values
	for _, fileEntry := range inFileNames {
		filePath := filepath.Join(r.inDir, fileEntry.Name())
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
			r.columnStartTime,
			r.columnEndTime,
			r.columnComment,
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
			startTimeStr := fmt.Sprintf("%s %s", row[headerMap[r.columnStartTime]], r.timezone)
			startTime, err := time.Parse(CSVInputTimeFormat, startTimeStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse start time '%s' in row %d: %s", startTimeStr, rowI, err)
			}

			endTimeStr := fmt.Sprintf("%s %s", row[headerMap[r.columnEndTime]], r.timezone)
			endTime, err := time.Parse(CSVInputTimeFormat, endTimeStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse end time '%s' in row %d: %s", endTimeStr, rowI, err)
			}

			entries := []TimeEntry{
				{
					StartTime: startTime,
					EndTime:   endTime,
					Comment:   row[headerMap[r.columnComment]],
				},
			}

			// Check if time passes over day boundary
			if startTime.Day() != endTime.Day() || startTime.Month() != endTime.Month() || startTime.Year() != endTime.Year() {
				// Make two entries, one that goes from the start time to the end of the day, and another which goes from midnight the following day to the following end time
				totalDuration := endTime.Sub(startTime)

				// Start day entry goes from original start time to end of the day
				endOfStartDay := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 23, 59, 59, 999999999, startTime.Location())
				startDayDuration := endOfStartDay.Sub(endOfStartDay)

				// End day entry goes from midnight of the following day to whatever duration is needed to reach the original duration
				startOfEndDay := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 0, 0, 0, 0, endTime.Location())
				endDayDuration := totalDuration - startDayDuration
				endOfEndDay := startOfEndDay.Add(endDayDuration)

				entries = []TimeEntry{
					{
						StartTime: startTime,
						EndTime:   endOfStartDay,
						Comment:   row[headerMap[r.columnComment]],
					},
					{
						StartTime: startOfEndDay,
						EndTime:   endOfEndDay,
						Comment:   row[headerMap[r.columnComment]],
					},
				}
			}

			// Save entry(s)
			for _, entry := range entries {
				entryHash, err := entry.Hash()
				if err != nil {
					return nil, fmt.Errorf("failed to hash time entry: %s", err)
				}
				timeEntries[entryHash] = entry
			}
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

	// Filter
	filteredTimeEntries := []TimeEntry{}
	for _, entry := range timeEntriesList {
		if opts.StartDate != nil && dateCompare(*opts.StartDate, entry.StartTime) < 0 {
			continue
		}

		if opts.EndDate != nil && dateCompare(*opts.EndDate, entry.StartTime) > 0 {
			continue
		}

		filteredTimeEntries = append(filteredTimeEntries, entry)
	}

	// Done
	return filteredTimeEntries, nil
}
