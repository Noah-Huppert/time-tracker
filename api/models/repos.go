package models

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Repos is a collection of model repositories
type Repos struct {
	// TimeEntry repository
	TimeEntry TimeEntryRepo

	// InvoiceSettings repository
	InvoiceSettings InvoiceSettingsRepo

	// CSVImport repository
	CSVImport CSVImportRepo
}

// NewReposOpts are options to create a new repos
type NewReposOpts struct {
	// DB is the database client
	DB *gorm.DB

	// Logger used by repositories
	Logger *zap.Logger
}

// NewRepos creates a new Repos
func NewRepos(opts NewReposOpts) Repos {
	return Repos{
		TimeEntry: DBTimeEntryRepo{
			db:     opts.DB,
			logger: opts.Logger.With(zap.String("repo", "DBTimeEntryRepo")),
		},
		InvoiceSettings: DBInvoiceSettingsRepo{
			db:     opts.DB,
			logger: opts.Logger.With(zap.String("repo", "DBInvoiceSettingsRepo")),
		},
		CSVImport: DBCSVImportRepo{
			db:     opts.DB,
			logger: opts.Logger.With(zap.String("repo", "DBCSVImportRepo")),
		},
	}
}
