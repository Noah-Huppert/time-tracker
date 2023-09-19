package models

import (
	"encoding/json"
	"fmt"
	"os"
)

// InvoiceSettings records details about how the user is paid for their work
type InvoiceSettings struct {
	// HourlyRate is the number of currency the user makes per hour of work
	HourlyRate float32 `json:"hourly_rate"`
}

// InvoiceSettingsRepo are methods to query and modify compensation information
type InvoiceSettingsRepo interface {
	// Get the invoice settings information
	Get() (*InvoiceSettings, error)

	// Set the invoice settings information
	Set(settings InvoiceSettings) error
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

		if err := r.Set(InvoiceSettings{}); err != nil {
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

func (r JSONInvoiceSettingsRepo) Set(settings InvoiceSettings) error {
	fileBytes, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshall into JSON: %s", err)
	}

	if err := os.WriteFile(r.filePath, fileBytes, 0666); err != nil {
		return fmt.Errorf("failed to write file: %s", err)
	}

	return nil
}
