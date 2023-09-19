package models

import (
	"encoding/json"
	"fmt"
	"os"
)

// Compensation records details about how the user is paid for their work
type Compensation struct {
	// HourlyRate is the number of currency the user makes per hour of work
	HourlyRate float32 `json:"hourly_rate"`
}

// CompensationRepo are methods to query and modify compensation information
type CompensationRepo interface {
	// Get the compensation information
	Get() (*Compensation, error)

	// Set the compensation information
	Set(compensation Compensation) error
}

// JSONCompensationRepo stores compensation information in a JSON data file
type JSONCompensationRepo struct {
	// filePath is the path to the JSON file in which compensation information will be stored
	filePath string
}

// NewJSONCompensationRepoOpts are options to create a new JSONCompensationRepo
type NewJSONCompensationRepoOpts struct {
	// FilePath is the path to the JSON file in which compensation information will be stored
	FilePath string
}

// NewJSONCompensationRepo creates a new JSONCompensationRepo
func NewJSONCompensationRepo(opts NewJSONCompensationRepoOpts) JSONCompensationRepo {
	return JSONCompensationRepo{
		filePath: opts.FilePath,
	}
}

func (r JSONCompensationRepo) Get() (*Compensation, error) {
	// Check if file exists
	if _, err := os.Stat(r.filePath); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to determine if data file: %s", err)
		}

		if err := r.Set(Compensation{}); err != nil {
			return nil, fmt.Errorf("failed to set compensation to default because file didn't exist: %s", err)
		}
	}

	// Get
	fileBytes, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s", err)
	}

	var comp Compensation
	if err := json.Unmarshal(fileBytes, &comp); err != nil {
		return nil, fmt.Errorf("failed to parse file: %s", err)
	}

	return &comp, nil
}

func (r JSONCompensationRepo) Set(comp Compensation) error {
	fileBytes, err := json.Marshal(comp)
	if err != nil {
		return fmt.Errorf("failed to marshall into JSON: %s", err)
	}

	if err := os.WriteFile(r.filePath, fileBytes, 0666); err != nil {
		return fmt.Errorf("failed to write file: %s", err)
	}

	return nil
}
