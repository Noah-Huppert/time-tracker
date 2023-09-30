package server

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Noah-Huppert/gointerrupt"
	"github.com/Noah-Huppert/time-tracker/api/models"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/zap"
)

// Server is the HTTP server
type Server struct {
	// logger is used to output messages
	logger *zap.Logger

	// validate is a validator instance
	validate *validator.Validate

	// repos are model repositories
	repos models.Repos
}

// NewServerOpts are options to create a new server
type NewServerOpts struct {
	// Logger for server
	Logger *zap.Logger

	// Repos are model repositories
	Repos models.Repos
}

// NewServer creates a new Server
func NewServer(opts NewServerOpts) Server {
	return Server{
		logger:   opts.Logger,
		validate: validator.New(),
		repos:    opts.Repos,
	}
}

// Listen starts the HTTP server
func (s Server) Listen(ctxPair gointerrupt.CtxPair, addr string) error {
	// Setup Fiber
	app := fiber.New(fiber.Config{
		ErrorHandler: s.errorHandler,
	})

	app.Use(fiberzap.New(fiberzap.Config{
		Logger: s.logger,
	}))

	app.Use(cors.New())

	// Setup routes
	app.Get("/api/v0/health", s.EPHealth)

	app.Get("/api/v0/time-entries", s.EPTimeEntriesList)
	app.Post("/api/v0/time-entries/upload-csv", s.EPTimeEntriesUploadCSV)

	app.Get("/api/v0/invoice-settings", s.EPInvoiceSettingsGet)
	app.Put("/api/v0/invoice-settings", s.EPInvoiceSettingsSet)

	// Setup server graceful shutdown
	shutdownErr := make(chan error, 1)
	go func() {
		<-ctxPair.Graceful().Done()
		if err := app.ShutdownWithContext(ctxPair.Harsh()); err != nil {
			shutdownErr <- fmt.Errorf("failed to shutdown server: %s", err)
			return
		}

		shutdownErr <- nil
	}()

	// Start server
	s.logger.Debug("starting to listen for HTTP traffic", zap.String("address", addr))
	if err := app.Listen(addr); err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}

	return <-shutdownErr
}

// ServerErrorResp is a generic error response body
type ServerErrorResp struct {
	// Error is the error text
	Error string `json:"error"`
}

// errorHandler is the Fiber server error handler
func (s Server) errorHandler(c *fiber.Ctx, err error) error {
	// Status code defaults to 500
	code := fiber.StatusInternalServerError

	// Retrieve the custom status code if it's a *fiber.Error
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	errTxt := err.Error()
	if code >= 500 {
		errTxt = "internal error"
	}

	s.logger.Error("endpoint error", zap.Error(err))

	return c.Status(code).JSON(ServerErrorResp{
		Error: errTxt,
	})
}

func (s Server) parseBody(c *fiber.Ctx, out interface{}) error {
	if err := c.BodyParser(out); err != nil {
		return fmt.Errorf("failed to parse body: %s", err)
	}

	if err := s.validate.Struct(out); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid body: %s", err))
	}

	return nil
}

// EPHealth is the health check endpoint
func (s Server) EPHealth(c *fiber.Ctx) error {
	return c.JSON(EPHealthResp{
		OK: true,
	})
}

// EPHealthResp is the health check endpoint response
type EPHealthResp struct {
	// OK indicates if the server is functioning correctly
	OK bool `json:"ok"`
}

// EPTimeEntriesList lists time entries.
// The start_time and end_time query params can be used to filter the range of time entries returned.
// Times should be in ISO-8601 format.
func (s Server) EPTimeEntriesList(c *fiber.Ctx) error {
	// Query params
	listOpts := models.ListTimeEntriesOpts{
		StartDate: nil,
		EndDate:   nil,
	}

	if startTimeQuery, ok := c.Queries()["start_date"]; ok {
		startTime, err := time.Parse(time.RFC3339, startTimeQuery)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("failed to parse start_time '%s' as ISO-8601 date: %s", startTimeQuery, err))
		}
		listOpts.StartDate = &startTime
	}

	if endTimeQuery, ok := c.Queries()["end_date"]; ok {
		endTime, err := time.Parse(time.RFC3339, endTimeQuery)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("failed to parse end_time '%s' as ISO-8601 date: %s", endTimeQuery, err))
		}
		listOpts.EndDate = &endTime
	}

	// Get time entries
	timeEntries, err := s.repos.TimeEntry.List(listOpts)
	if err != nil {
		return fmt.Errorf("failed to list time entries: %s", err)
	}

	// Add additional fields onto time entries
	listItems := []TimeEntryListItem{}
	var totalDuration time.Duration

	for _, timeEntry := range timeEntries {
		duration := timeEntry.Duration()
		totalDuration += duration

		listItems = append(listItems, TimeEntryListItem{
			TimeEntry: timeEntry,
			Duration:  duration,
		})
	}

	return c.JSON(EPTimeEntriesListResp{
		TimeEntries:   listItems,
		TotalDuration: totalDuration,
	})
}

// EPTimeEntriesUploadCSV creates new time entries via CSV
func (s Server) EPTimeEntriesUploadCSV(c *fiber.Ctx) error {
	// Parse body
	var body EPTimeEntriesUploadCSVReq
	if err := s.parseBody(c, &body); err != nil {
		return err
	}

	// Parse entries from each file
	csvParser := models.NewCSVTimeEntryParser(models.NewCSVTimeEntryParserOpts{
		Timezone:        "EST",
		ColumnStartTime: "time started",
		ColumnEndTime:   "time ended",
		ColumnComment:   "comment",
	})

	csvImports := []models.CSVImport{}
	existingTimeEntries := []models.TimeEntry{}
	newTimeEntries := []models.TimeEntry{}

	for _, file := range body.CSVFiles {
		// Parse time entries
		parsedEntries, err := csvParser.Parse(strings.NewReader(file.Content))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("failed to parse '%s': %s", file.Name, err))
		}

		// Create CSV import
		csvImport := models.CSVImport{
			FileName:     file.Name,
			FileContents: file.Content,
		}
		if err := s.repos.CSVImport.Create(&csvImport); err != nil {
			return fmt.Errorf("failed to create CSV import for '%s' file: %s", file.Name, err)
		}
		csvImports = append(csvImports, csvImport)

		// Add new entries into db
		insertRes, err := s.repos.TimeEntry.Create(csvImport, parsedEntries)
		if err != nil {
			return fmt.Errorf("failed to insert entries: %s", err)
		}

		existingTimeEntries = append(existingTimeEntries, insertRes.ExistingEntries...)
		newTimeEntries = append(newTimeEntries, insertRes.NewEntries...)
	}

	return c.JSON(EPTimeEntriesUploadCSVResp{
		ExistingTimeEntries: existingTimeEntries,
		NewTimeEntries:      newTimeEntries,
		CSVImports:          csvImports,
	})
}

// EPTimeEntriesUploadCSVReq is the upload time entries CSV request
type EPTimeEntriesUploadCSVReq struct {
	CSVFiles []EPTimeEntriesUploadCSVReqFile `json:"csv_files"`
}

// EPTimeEntriesUploadCSVReqFile is a file in an upload request
type EPTimeEntriesUploadCSVReqFile struct {
	// Name of file
	Name string `json:"name"`

	// Content of file
	Content string `json:"content"`
}

// EPTimeEntriesUploadCSVResp is the response for the upload time entries CSV endpoint
type EPTimeEntriesUploadCSVResp struct {
	// ExistingTimeEntries are time entries which already existed
	ExistingTimeEntries []models.TimeEntry `json:"existing_time_entries"`

	// NewTimeEntries are time entries which were newly created
	NewTimeEntries []models.TimeEntry `json:"new_time_entries"`

	// CSVImports are the CSV import records created during the import
	CSVImports []models.CSVImport
}

// EPTimeEntriesListResp is the list time entries endpoint response
type EPTimeEntriesListResp struct {
	// TimeEntries is the list of time entries
	TimeEntries []TimeEntryListItem `json:"time_entries"`

	// TotalDuration is the duration of each time entry added up
	TotalDuration time.Duration `json:"total_duration"`
}

// TimeEntryListItem is a time entry item in a list endpoint response
type TimeEntryListItem struct {
	models.TimeEntry

	// Duration of the time entry
	Duration time.Duration `json:"duration"`
}

// EPInvoiceSettingsGet gets invoice settings
func (s Server) EPInvoiceSettingsGet(c *fiber.Ctx) error {
	settings, err := s.repos.InvoiceSettings.Get()
	if err != nil {
		return fmt.Errorf("failed to get invoice settings: %s", err)
	}

	return c.JSON(settings)
}

// EPInvoiceSettingsSet sets invoice settings
func (s Server) EPInvoiceSettingsSet(c *fiber.Ctx) error {
	// Parse body
	var body EPInvoiceSettingsSetReq
	if err := s.parseBody(c, &body); err != nil {
		return err
	}

	// Set
	newSettings := models.InvoiceSettings{
		HourlyRate: body.HourlyRate,
		Recipient:  body.Recipient,
		Sender:     body.Sender,
	}
	s.logger.Debug("EPInvoiceSettingsSet", zap.Any("newSettings", newSettings))
	if err := s.repos.InvoiceSettings.Set(&newSettings); err != nil {
		return fmt.Errorf("failed to set invoice settings: %s", err)
	}

	return c.JSON(newSettings)
}

// EPInvoiceSettingsSetReq is the set invoice settings request body
type EPInvoiceSettingsSetReq struct {
	// HourlyRate is the new hourly rate value
	HourlyRate float32 `json:"hourly_rate" validate:"required"`

	// Recipient is the new recipient value
	Recipient string `json:"recipient" validate:"required"`

	// Sender is the new sender value
	Sender string `json:"sender" validate:"required"`
}
