package models

import (
	"encoding/json"
	"fmt"
	"os"

	"gorm.io/gorm"
)

// InvoiceSettings records details about how the user is paid for their work
type InvoiceSettings struct {
	// ID is the unique identifier
	ID uint `gorm:"primarykey"`

	// HourlyRate is the number of currency the user makes per hour of work
	HourlyRate float32 `gorm:"not null" json:"hourly_rate"`

	// Recipient is information about the person who is receiving the invoice
	Recipient string `gorm:"not null" json:"recipient"`

	// Sender is information about the person who is sending the invoice
	Sender string `gorm:"not null" json:"sender"`
}

// InvoiceSettingsRepo are methods to query and modify compensation information
type InvoiceSettingsRepo interface {
	// Get the invoice settings information, if information doesn't exist then initialize to default settings
	Get() (*InvoiceSettings, error)

	// Set the invoice settings information
	Set(settings *InvoiceSettings) error
}

// JSONInvoiceSettingsRepo stores compensation information in a JSON data file
type JSONInvoiceSettingsRepo struct {
	// filePath is the path to the JSON file in which information will be stored
	filePath string
}

// NewJSONInvoiceSettingsRepoOpts are options to create a new JSONCompensationRepo
type NewJSONInvoiceSettingsRepoOpts struct {
	// FilePath is the path to the JSON file in which information will be stored
	FilePath string
}

// NewJSONInvoiceSettingsRepo creates a new JSONInvoiceSettingsRepo
func NewJSONInvoiceSettingsRepo(opts NewJSONInvoiceSettingsRepoOpts) JSONInvoiceSettingsRepo {
	return JSONInvoiceSettingsRepo{
		filePath: opts.FilePath,
	}
}

func (r JSONInvoiceSettingsRepo) Get() (*InvoiceSettings, error) {
	// Check if file exists
	if _, err := os.Stat(r.filePath); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to determine if data file: %s", err)
		}

		if err := r.Set(&InvoiceSettings{}); err != nil {
			return nil, fmt.Errorf("failed to set to default because file didn't exist: %s", err)
		}
	}

	// Get
	fileBytes, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s", err)
	}

	var settings InvoiceSettings
	if err := json.Unmarshal(fileBytes, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse file: %s", err)
	}

	return &settings, nil
}

func (r JSONInvoiceSettingsRepo) Set(settings *InvoiceSettings) error {
	fileBytes, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshall into JSON: %s", err)
	}

	if err := os.WriteFile(r.filePath, fileBytes, 0666); err != nil {
		return fmt.Errorf("failed to write file: %s", err)
	}

	return nil
}

// DBInvoiceSettingsRepo implements InvoiceSettingsRepo using a database
type DBInvoiceSettingsRepo struct {
	// db is the database client
	db *gorm.DB
}

func (r DBInvoiceSettingsRepo) Get() (*InvoiceSettings, error) {
	// Get all rows
	var settings []InvoiceSettings
	if res := r.db.Find(&settings); res.Error != nil {
		return nil, fmt.Errorf("failed to run get query: %s", res.Error)
	}

	// Initialize a settings if none exist
	if len(settings) == 0 {
		initSettings := InvoiceSettings{}
		if err := r.Set(&initSettings); err != nil {
			return nil, fmt.Errorf("failed to add initial settings: %s", err)
		}

		return &initSettings, nil
	}

	// Check no more than 1 row exists
	if len(settings) != 1 {
		return nil, fmt.Errorf("%d invoice settings row(s) found, only one row should ever exist", len(settings))
	}

	return &settings[0], nil
}

func (r DBInvoiceSettingsRepo) Set(settings *InvoiceSettings) error {
	// Get all rows
	var allSettings []InvoiceSettings
	if res := r.db.Find(&allSettings); res.Error != nil {
		return fmt.Errorf("failed to run get query: %s", res.Error)
	}

	// If no rows insert first row
	if len(allSettings) == 0 {
		if res := r.db.Create(settings); res.Error != nil {
			return fmt.Errorf("failed to insert new settings: %s", res.Error)
		}
		return nil
	}

	// Check no more than 1 row exists
	if len(allSettings) != 1 {
		return fmt.Errorf("%d invoice settings row(s) found, only one row should ever exist", len(allSettings))
	}

	// Update
	settings.ID = allSettings[0].ID
	if res := r.db.Save(settings); res.Error != nil {
		return fmt.Errorf("failed to update existing settings: %s", res.Error)
	}

	return nil
}
