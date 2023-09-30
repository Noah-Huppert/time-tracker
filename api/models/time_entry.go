package models

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TimeEntry records a period of time when work was completed
type TimeEntry struct {
	// ID is the unique identifier
	ID uint `gorm:"primarykey" json:"id"`

	// Hash of fields which define the main component of a time entry, see TimeEntry.IdentityFields()
	Hash string `gorm:"not null" json:"hash"`

	/// StartTime is the date and time when the period started
	StartTime time.Time `gorm:"not null" json:"start_time"`

	// EndTime is the date and time when the period ended
	EndTime time.Time `gorm:"not null" json:"end_time"`

	// Command is an optional comment explaining what work was completed during the period
	Comment string `gorm:"not null" json:"comment"`
}

// IdentityFields returns a map of the StartTime, EndTime, and Comment fields. These fields define what makes a time entry unique.
func (e TimeEntry) IdentityFields() map[string]interface{} {
	return map[string]interface{}{
		"StartTime": e.StartTime,
		"EndTime":   e.EndTime,
		"Comment":   e.Comment,
	}
}

// ComputeIdentityHash returns a checksum of the TimeEntry.IdentityFields(), can be used to find duplicate TimeEntry structs
func (e TimeEntry) ComputeIdentityHash() (string, error) {
	hash, err := hashstructure.Hash(e.IdentityFields(), hashstructure.FormatV2, nil)
	if err != nil {
		return "", fmt.Errorf("failed to hash identity fields: %s", err)
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

	// Create time entries
	Create(timeEntries []TimeEntry) ([]TimeEntry, error)
}

// ListTimeEntriesOpts are options for listing time entries
type ListTimeEntriesOpts struct {
	// StartDate indicates only time entries which started after (inclusive) this date should be returned
	StartDate *time.Time

	// EndDate indicates only time entries which started before (inclusive) this date should be returned
	EndDate *time.Time
}

// DBTimeEntryRepo implements a TimeEntryRepo using a database
type DBTimeEntryRepo struct {
	// db client
	db *gorm.DB
}

func (r DBTimeEntryRepo) List(opts ListTimeEntriesOpts) ([]TimeEntry, error) {
	// Base query
	var timeEntries []TimeEntry
	queryTx := r.db.Order("StartTime DESC")

	// Filter options
	if opts.StartDate != nil {
		queryTx.Where("StartTime >= ?-?-?", opts.StartDate.Year(), opts.StartDate.Month(), opts.StartDate.Day())
	}

	if opts.EndDate != nil {
		queryTx.Where("StartTime <= ?-?-?", opts.EndDate.Year(), opts.EndDate.Month(), opts.EndDate.Day())
	}

	// Run query
	queryRes := queryTx.Find(&timeEntries)
	if queryRes.Error != nil {
		return nil, fmt.Errorf("failed to run list query: %s", queryRes.Error)
	}

	return timeEntries, nil
}

func (r DBTimeEntryRepo) Create(timeEntries []TimeEntry) ([]TimeEntry, error) {
	r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "hash"},
		},
		DoUpdates: clause.AssignmentColumns({
			""
		}),
	})
}

// CSVTimeEntryParser reads a CSV file's contents and creates time entries
type CSVTimeEntryParser struct {
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

// NewCSVTimeEntryParserOpts are options for creating a CSVTimeEntryParser
type NewCSVTimeEntryParserOpts struct {
	// Timezone in which times are in
	Timezone string

	// ColumnStartTime is the name of the start time column
	ColumnStartTime string

	// ColumnEndTime is the name of the end time column
	ColumnEndTime string

	// ColumnComment is the name of the comment column
	ColumnComment string
}

// NewCSVTimeEntryParser creates a new CSVTimeEntryRepo
func NewCSVTimeEntryParser(opts NewCSVTimeEntryParserOpts) CSVTimeEntryParser {
	return CSVTimeEntryParser{
		timezone:        opts.Timezone,
		columnStartTime: opts.ColumnStartTime,
		columnEndTime:   opts.ColumnEndTime,
		columnComment:   opts.ColumnComment,
	}
}

// DateCompare only compares the date (year, month, day) component of a time. Returns less than 0 if compare is before base, greater than 0 if compare is after base, 0 if the same.
func DateCompare(base time.Time, compare time.Time) int32 {
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

func (r CSVTimeEntryParser) Parse(csvIn io.Reader) ([]TimeEntry, error) {
	timeEntries := make(map[string]TimeEntry) // keys are hashes of the values

	csvReader := csv.NewReader(csvIn)

	// Read headers
	headers, err := csvReader.Read()
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
			return nil, fmt.Errorf("missing column '%s'", requiredCol)
		}
	}

	// Parse rows into TimeEntry structs
	rows, err := csvReader.ReadAll()
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
			entryHash, err := entry.ComputeIdentityHash()
			if err != nil {
				return nil, fmt.Errorf("failed to hash time entry: %s", err)
			}

			entry.Hash = entryHash
			timeEntries[entryHash] = entry
		}
	}

	// Convert map into list
	timeEntriesList := []TimeEntry{}

	for _, timeEntry := range timeEntries {
		timeEntriesList = append(timeEntriesList, timeEntry)
	}

	return timeEntriesList, nil
}
