package models

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/Noah-Huppert/time-tracker/api/timeutil"
	"github.com/mitchellh/hashstructure/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TimeEntry records a period of time when work was completed
type TimeEntry struct {
	// ID is the unique identifier
	ID uint `gorm:"primarykey" json:"id"`

	/// StartTime is the date and time when the period started
	StartTime timeutil.APITime `gorm:"not null;index:time_entry_identity_unique,unique" json:"start_time"`

	// EndTime is the date and time when the period ended
	EndTime timeutil.APITime `gorm:"not null;index:time_entry_identity_unique,unique" json:"end_time"`

	// Duration is the amount of time between the start and end time
	Duration time.Duration `gorm:"not null" json:"duration"`

	// Command is an optional comment explaining what work was completed during the period
	Comment string `gorm:"not null;index:time_entry_identity_unique,unique" json:"comment"`

	// CSVImportID is the ID of the CSVImport which created the time entry
	CSVImportID uint `gorm:"not null" json:"csv_import_id"`

	// CSVImport is an ORM filled reference filled by the CSVImportID foreign key
	CSVImport CSVImport `json:"-"`
}

// Duration of the time entry
func (e TimeEntry) ComputeDuration() time.Duration {
	return e.EndTime.Sub(e.StartTime.Time)
}

// IdentityFields returns a tuple of the StartTime, EndTime, and Comment fields. These fields define what makes a time entry unique.
func (e TimeEntry) IdentityFields() []interface{} {
	return []interface{}{
		e.StartTime,
		e.EndTime,
		e.Comment,
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

// TimeEntryRepo are functions to query and retrieve TimeEntries
type TimeEntryRepo interface {
	// List time entries sorted earliest to latest
	List(opts ListTimeEntriesOpts) ([]TimeEntry, error)

	// Create time entries, should not insert duplicate time entries (duplicate meaning all fields except ID are the same)
	Create(csvImport CSVImport, timeEntries []TimeEntry) (*CreateTimeEntriesRes, error)
}

// ListTimeEntriesOpts are options for listing time entries
type ListTimeEntriesOpts struct {
	// StartDate indicates only time entries which started after (inclusive) this date should be returned
	StartDate *timeutil.APITime

	// EndDate indicates only time entries which started before (inclusive) this date should be returned
	EndDate *timeutil.APITime
}

// CreateTimeEntriesRes is a create time entry result
type CreateTimeEntriesRes struct {
	// ExistingEntries are time entries which already existed in the database
	ExistingEntries []TimeEntry

	// NewEntries are time entries which didn't exist
	NewEntries []TimeEntry
}

// DBTimeEntryRepo implements a TimeEntryRepo using a database
type DBTimeEntryRepo struct {
	// db client
	db *gorm.DB

	// logger used to output runtime information
	logger *zap.Logger
}

func makeDateOnlyTime(in time.Time) time.Time {
	return time.Date(in.Year(), in.Month(), in.Day(), 0, 0, 0, 0, in.Location())
}

func (r DBTimeEntryRepo) List(opts ListTimeEntriesOpts) ([]TimeEntry, error) {
	// Base query
	var timeEntries []TimeEntry
	queryTx := r.db.Order("start_time ASC")

	// Filter options
	if opts.StartDate != nil && opts.EndDate != nil {
		queryTx.Where("start_time >= ? AND start_time <= ?", opts.StartDate.MakeDateOnly(), opts.EndDate.MakeDateOnly())
	} else if opts.StartDate != nil {
		queryTx.Or("start_time >= ?", opts.StartDate.MakeDateOnly())
	} else if opts.EndDate != nil {
		queryTx.Or("start_time <= ?", opts.EndDate.MakeDateOnly())
	}

	// Run query
	queryRes := queryTx.Find(&timeEntries)
	if queryRes.Error != nil {
		return nil, fmt.Errorf("failed to run list query: %s", queryRes.Error)
	}

	return timeEntries, nil
}

func (r DBTimeEntryRepo) Create(csvImport CSVImport, timeEntries []TimeEntry) (*CreateTimeEntriesRes, error) {
	// Find existing entries
	var existingEntries []TimeEntry
	existingEntriesWhereTuple := [][]interface{}{}
	for _, entry := range timeEntries {
		existingEntriesWhereTuple = append(existingEntriesWhereTuple, entry.IdentityFields())
	}

	res := r.db.Where("(start_time, end_time, comment) IN ?", existingEntriesWhereTuple).Find(&existingEntries)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to find existing time entries: %s", res.Error)
	}

	existingEntriesHashes := map[string]interface{}{}
	for _, entry := range existingEntries {
		hash, err := entry.ComputeIdentityHash()
		if err != nil {
			return nil, fmt.Errorf("failed to compute identity hash of time entry with ID '%d': %s", entry.ID, err)
		}

		existingEntriesHashes[hash] = nil
	}

	// Find entries to create
	toCreateEntries := []TimeEntry{}
	for entryI, entry := range timeEntries {
		hash, err := entry.ComputeIdentityHash()
		if err != nil {
			return nil, fmt.Errorf("failed to compute identity hash of time entry with index %d in timeEntries argument: %s", entryI, err)
		}

		if _, ok := existingEntriesHashes[hash]; !ok {
			// Associate with CSV import
			entry.CSVImportID = csvImport.ID
			toCreateEntries = append(toCreateEntries, entry)
		}
	}

	if len(toCreateEntries) > 0 {
		res = r.db.Create(&toCreateEntries)
		if res.Error != nil {
			return nil, fmt.Errorf("failed to run batch insert query: %s", res.Error)
		}
	}

	createRes := CreateTimeEntriesRes{
		ExistingEntries: existingEntries,
		NewEntries:      toCreateEntries,
	}

	return &createRes, nil
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
		startTime, err := timeutil.ParseFormat(CSVInputTimeFormat, startTimeStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start time '%s' in row %d: %s", startTimeStr, rowI, err)
		}

		endTimeStr := fmt.Sprintf("%s %s", row[headerMap[r.columnEndTime]], r.timezone)
		endTime, err := timeutil.ParseFormat(CSVInputTimeFormat, endTimeStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end time '%s' in row %d: %s", endTimeStr, rowI, err)
		}

		entries := []TimeEntry{
			{
				StartTime: *startTime,
				EndTime:   *endTime,
				Comment:   row[headerMap[r.columnComment]],
			},
		}

		// Check if time passes over day boundary
		if startTime.Day() != endTime.Day() || startTime.Month() != endTime.Month() || startTime.Year() != endTime.Year() {
			// Make two entries, one that goes from the start time to the end of the day, and another which goes from midnight the following day to the following end time
			totalDuration := endTime.Sub(startTime.Time)

			// Start day entry goes from original start time to end of the day
			endOfStartDay, err := timeutil.NewAPITime(time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 23, 59, 59, 999999999, startTime.Time.Location()))
			if err != nil {
				return nil, fmt.Errorf("failed to make end of first day APITime: %s", err)
			}
			startDayDuration := endOfStartDay.Sub(endOfStartDay.Time)

			// End day entry goes from midnight of the following day to whatever duration is needed to reach the original duration
			startOfEndDay, err := timeutil.NewAPITime(time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 0, 0, 0, 0, endTime.Location()))
			if err != nil {
				return nil, fmt.Errorf("failed to make start of second day APITime: %s", err)
			}
			endDayDuration := totalDuration - startDayDuration
			endOfEndDay, err := timeutil.NewAPITime(startOfEndDay.Add(endDayDuration))
			if err != nil {
				return nil, fmt.Errorf("failed to make end of second day APITime: %s", err)
			}

			entries = []TimeEntry{
				{
					StartTime: *startTime,
					EndTime:   *endOfStartDay,
					Comment:   row[headerMap[r.columnComment]],
				},
				{
					StartTime: *startOfEndDay,
					EndTime:   *endOfEndDay,
					Comment:   row[headerMap[r.columnComment]],
				},
			}
		}

		// Save entry(s)
		timeEntries = append(timeEntries, entries...)
	}

	// Computed fields
	computedTimeEntries := []TimeEntry{}
	for _, entry := range timeEntries {
		entry.Duration = entry.ComputeDuration()
		computedTimeEntries = append(computedTimeEntries, entry)
	}

	return computedTimeEntries, nil
}
