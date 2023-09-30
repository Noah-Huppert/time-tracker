package models

import (
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InvoiceSettings records details about how the user is paid for their work
type InvoiceSettings struct {
	// ID is the unique identifier
	ID uint `gorm:"primarykey" json:"id"`

	// Slot is used to ensure that only one InvoiceSettings exists in the database
	Slot string `gorm:"not null;unique" json:"-"`

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

	// logger used to output runtime information
	logger *zap.Logger
}

const DB_INVOICE_SETTINGS_SLOT = "primary"

func (r DBInvoiceSettingsRepo) getRow() (*InvoiceSettings, error) {
	var allSettings []InvoiceSettings
	if res := r.db.Where("slot = ?", DB_INVOICE_SETTINGS_SLOT).Find(&allSettings); res.Error != nil {
		return nil, fmt.Errorf("failed to run get row query: %s", res.Error)
	}

	if len(allSettings) == 0 {
		return nil, nil
	}

	if len(allSettings) > 1 {
		return nil, fmt.Errorf("%d row(s) found, only one row should exist", len(allSettings))
	}

	return &allSettings[0], nil
}

func (r DBInvoiceSettingsRepo) Get() (*InvoiceSettings, error) {
	//  Get settings row
	settings, err := r.getRow()
	if err != nil {
		return nil, fmt.Errorf("failed to get single settings rows: %s", err)
	}

	// Return default settings if none exist
	if settings == nil {
		initSettings := InvoiceSettings{}
		return &initSettings, nil
	}

	return settings, nil
}

func (r DBInvoiceSettingsRepo) Set(newSettings *InvoiceSettings) error {
	//  Get settings row
	settings, err := r.getRow()
	if err != nil {
		return fmt.Errorf("failed to get single settings rows: %s", err)
	}

	// If no rows insert first row
	if settings == nil {
		newSettings.Slot = DB_INVOICE_SETTINGS_SLOT
		if res := r.db.Create(newSettings); res.Error != nil {
			return fmt.Errorf("failed to insert new settings: %s", res.Error)
		}
		return nil
	}

	// Update
	if res := r.db.Model(settings).UpdateColumns(newSettings); res.Error != nil {
		return fmt.Errorf("failed to update existing settings: %s", res.Error)
	}

	return nil
}
