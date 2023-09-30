package models

import (
	"fmt"

	"github.com/lib/pq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CSVImport records an upload of a CSV file to create time entries
type CSVImport struct {
	// ID is a unique identifier
	ID uint `gorm:"primary key" json:"id"`

	// FileName is the name of the CSV file which was uploaded
	FileName string `gorm:"not null" json:"file_name"`

	// FileContents is the text contents of the imported file
	FileContents string `gorm:"not null" json:"file_contents"`

	// DuplicateTimeEntryIDs is a list of time entries found in this FileContents which were already uploaded
	DuplicateTimeEntryIDs pq.Int64Array `gorm:"not null;type:int[]" json:"duplicate_time_entry_ids"`

	// TimeEntries created by the import
	TimeEntries []TimeEntry `gorm:"-" json:"-"`
}

// CSVImportRepo is a repository for a querying csv imports
type CSVImportRepo interface {
	// Create a CSV import
	Create(csvImport *CSVImport) error

	// Update a csv import
	Update(csvImport *CSVImport) error
}

// DBCSVImportRepo implements a CSVImportRepo using a database client
type DBCSVImportRepo struct {
	// db client
	db *gorm.DB

	// logger used to output runtime entries
	logger *zap.Logger
}

func (r DBCSVImportRepo) Create(csvImport *CSVImport) error {
	if res := r.db.Create(csvImport); res.Error != nil {
		return fmt.Errorf("failed to run create query: %s", res.Error)
	}

	return nil
}

func (r DBCSVImportRepo) Update(csvImport *CSVImport) error {
	if res := r.db.Save(csvImport); res.Error != nil {
		return fmt.Errorf("failed to run update query: %s", res.Error)
	}

	return nil
}
