package models

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TimeEntry records a period of time when work was completed
type TimeEntry struct {
	// ID is the unique identifier
	ID uint `gorm:"primarykey" json:"id"`

	/// StartTime is the date and time when the period started
	StartTime time.Time `gorm:"not null" json:"start_time"`

	// EndTime is the date and time when the period ended
	EndTime time.Time `gorm:"not null" json:"end_time"`

	// Command is an optional comment explaining what work was completed during the period
	Comment string `gorm:"not null" json:"comment"`
}

// Duration of the time entry
func (e TimeEntry) Duration() time.Duration {
	return e.EndTime.Sub(e.StartTime)
}

// TimeEntryRepo are functions to query and retrieve TimeEntries
type TimeEntryRepo interface {
	// List time entries sorted earliest to latest
	List(opts ListTimeEntriesOpts) ([]TimeEntry, error)

	// Create time entries, should not insert duplicate time entries (duplicate meaning all fields except ID are the same)
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

	// logger used to output runtime information
	logger *zap.Logger
}

func (r DBTimeEntryRepo) List(opts ListTimeEntriesOpts) ([]TimeEntry, error) {
	// Base query
	var timeEntries []TimeEntry
	queryTx := r.db.Order("start_time DESC")

	// Filter options
	if opts.StartDate != nil {
		queryTx.Where("start_time >= ?-?-?", opts.StartDate.Year(), opts.StartDate.Month(), opts.StartDate.Day())
	}

	if opts.EndDate != nil {
		queryTx.Where("start_time <= ?-?-?", opts.EndDate.Year(), opts.EndDate.Month(), opts.EndDate.Day())
	}

	// Run query
	queryRes := queryTx.Find(&timeEntries)
	if queryRes.Error != nil {
		return nil, fmt.Errorf("failed to run list query: %s", queryRes.Error)
	}

	return timeEntries, nil
}

func (r DBTimeEntryRepo) Create(timeEntries []TimeEntry) ([]TimeEntry, error) {
	res := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "hash"},
		},
		DoNothing: true,
	}).Create(&timeEntries)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to run batch insert query: %s", res.Error)
	}

	return timeEntries, nil
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

// Parse CSV file contents into time entries, does not perform de-duplication
func (r CSVTimeEntryParser) Parse(csvIn io.Reader) ([]TimeEntry, error) {
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

	timeEntries := []TimeEntry{}
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
		timeEntries = append(timeEntries, entries...)
	}

	return timeEntries, nil
}
