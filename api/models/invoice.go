package models

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// HOUR_IN_NANOSECONDS is the number of nanoseconds in an hour
const HOUR_AS_NANOSECONDS = 3.6e+12

// Invoice is a record of time entries over a certain time period
type Invoice struct {
	// ID is a unique ID
	ID uint `gorm:"primarykey" json:"id"`

	// InvoiceSettings are the settings used by the invoice
	InvoiceSettingsID uint `gorm:"not null" json:"invoice_settings_id"`

	// StartDate is the date on which the invoice time entries start
	StartDate time.Time `gorm:"not null" json:"start_date"`

	// EndDate is the date on which the invoice time entries ended
	EndDate time.Time `gorm:"not null" json:"end_date"`

	// Duration of time of time entries in invoice
	Duration time.Duration `gorm:"not null" json:"duration"`

	// AmountDue is the amount of money due for the duration of time
	AmountDue float64 `gorm:"not null" json:"amount_due"`

	// SentToClient if not null the invoice has been sent to a client on that date
	SentToClient *time.Time `json:"sent_to_client"`

	// PaidByClient if not null the invoice was paid by the client on that date
	PaidByClient *time.Time `json:"paid_by_client"`

	// Archived is used to soft delete invoices
	Archived bool `gorm:"not null" json:"archived"`

	// InvoiceSettings is an ORM filled field based on the InvoiceSettingsID foreign key
	InvoiceSettings InvoiceSettings `json:"invoice_settings"`

	// InvoiceTimeEntries are time entry associations for the invoice
	InvoiceTimeEntries []InvoiceTimeEntry `json:"invoice_time_entries"`
}

// InvoiceTimeEntry associates a TimeEntry with an Invoice
type InvoiceTimeEntry struct {
	// ID is a unique ID
	ID uint `gorm:"primarykey" json:"id"`

	// InvoiceID is the ID of the Invoice
	InvoiceID uint `gorm:"not null" json:"invoice_id"`

	// TimeEntryID is the ID of the TimeEntry
	TimeEntryID uint `gorm:"not null" json:"time_entry_id"`

	// Invoice is an ORM filled field based on the InvoiceID foreign key
	Invoice Invoice `json:"-"`

	// TimeEntry is an ORM filled field based on the TimeEntryID foreign key
	TimeEntry TimeEntry `json:"time_entry"`
}

// InvoiceRepo queries invoices
type InvoiceRepo interface {
	// Create an invoice
	Create(opts CreateInvoiceOpts) (*CreateInvoiceRes, error)

	// List invoices
	List(opts ListInvoicesOpts) ([]Invoice, error)

	// Update an invoice, can't update most fields as they are immutable
	Update(opts UpdateInvoiceOpts) (*Invoice, error)
}

// CreateInvoiceOpts are options for creating a new invoice
type CreateInvoiceOpts struct {
	// InvoiceSettings are the settings to use for this invoice
	InvoiceSettings InvoiceSettings

	// StartDate of invoice
	StartDate time.Time

	// EndDate of invoice
	EndDate time.Time

	// TimeEntries for invoice
	TimeEntries []TimeEntry
}

// CreateInvoiceRes is the result of creating an invoice
type CreateInvoiceRes struct {
	// Invoice which was created
	Invoice Invoice `json:"invoice"`

	// InvoiceTimeEntries are InvoiceTimeEntries which were created
	InvoiceTimeEntries []InvoiceTimeEntry `json:"invoice_time_entries"`
}

// ListInvoicesOpts are options for listing invoices
type ListInvoicesOpts struct {
	// IDs is a list of invoice IDs which should be retrieved
	IDs []uint64

	// Archived filters invoices which are archived or not archived, if nil no filtering done
	Archived *bool
}

// UpdateInvoiceOpts specify the new values of the parts of an invoice which can be updated
type UpdateInvoiceOpts struct {
	// ID of the invoice to update
	ID uint

	// SentToClient if provided updates the SentToClient field
	SentToClient *time.Time

	// PaidByClient if provided updates the PaidByClient field
	PaidByClient *time.Time
}

// DBInvoiceRepo implements a InvoiceRepo using a database client
type DBInvoiceRepo struct {
	// db client
	db *gorm.DB

	// logger used to output runtime information
	logger *zap.Logger
}

func (r DBInvoiceRepo) Create(opts CreateInvoiceOpts) (*CreateInvoiceRes, error) {
	// Calculate total duration
	duration := time.Duration(0)
	for _, entry := range opts.TimeEntries {
		duration += entry.Duration
	}

	// Calculate amount due
	amountDue := duration.Hours() * opts.InvoiceSettings.HourlyRate

	// Create invoice
	invoice := Invoice{
		InvoiceSettingsID: opts.InvoiceSettings.ID,
		StartDate:         opts.StartDate,
		EndDate:           opts.EndDate,
		SentToClient:      nil,
		PaidByClient:      nil,
		Duration:          duration,
		AmountDue:         amountDue,
	}
	if res := r.db.Create(&invoice); res.Error != nil {
		return nil, fmt.Errorf("failed to run create query: %s", res.Error)
	}

	// Create invoice time entries
	invoiceTimeEntries := []InvoiceTimeEntry{}
	for _, entry := range opts.TimeEntries {
		invoiceTimeEntries = append(invoiceTimeEntries, InvoiceTimeEntry{
			InvoiceID:   invoice.ID,
			TimeEntryID: entry.ID,
		})
	}
	if res := r.db.Create(&invoiceTimeEntries); res.Error != nil {
		return nil, fmt.Errorf("failed to run create invoice time entries query: %s", res.Error)
	}
	invoice.InvoiceTimeEntries = invoiceTimeEntries
	invoice.InvoiceSettings = opts.InvoiceSettings

	return &CreateInvoiceRes{
		Invoice:            invoice,
		InvoiceTimeEntries: invoiceTimeEntries,
	}, nil
}

func (r DBInvoiceRepo) List(opts ListInvoicesOpts) ([]Invoice, error) {
	// Filter options
	tx := r.db.Model(&Invoice{}).Preload("InvoiceTimeEntries").Preload("InvoiceTimeEntries.TimeEntry").Preload("InvoiceSettings")
	if len(opts.IDs) > 0 {
		tx.Where("id IN ?", opts.IDs)
	}

	if opts.Archived != nil {
		tx.Where("archived = ?", *opts.Archived)
	}

	// Get
	var invoices []Invoice
	if res := tx.Find(&invoices); res.Error != nil {
		return nil, fmt.Errorf("failed to run list query: %s", res.Error)
	}

	return invoices, nil
}

func (r DBInvoiceRepo) Update(opts UpdateInvoiceOpts) (*Invoice, error) {
	// Get invoice
	invoices, err := r.List(ListInvoicesOpts{
		IDs: []uint64{uint64(opts.ID)},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list invoice: %s", err)
	}
	if len(invoices) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	invoice := invoices[0]

	// Perform updates
	updates := map[string]interface{}{}
	updates["sent_to_client"] = opts.SentToClient
	updates["paid_by_client"] = opts.PaidByClient

	if res := r.db.Model(&invoice).Updates(updates); res.Error != nil {
		return nil, fmt.Errorf("failed to run update query: %s", res.Error)
	}

	return &invoice, nil
}
