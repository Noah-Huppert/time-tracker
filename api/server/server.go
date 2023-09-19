package server

import (
	"errors"
	"fmt"

	"github.com/Noah-Huppert/gointerrupt"
	"github.com/Noah-Huppert/time-tracker/api/models"
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Server is the HTTP server
type Server struct {
	// logger is used to output messages
	logger *zap.Logger

	// timeEntryRepo is the time entry repository
	timeEntryRepo models.TimeEntryRepo
}

// NewServerOpts are options to create a new server
type NewServerOpts struct {
	// Logger for server
	Logger *zap.Logger

	// TimeEntryRepo is the time entry repository
	TimeEntryRepo models.TimeEntryRepo
}

// NewServer creates a new Server
func NewServer(opts NewServerOpts) Server {
	return Server{
		logger:        opts.Logger,
		timeEntryRepo: opts.TimeEntryRepo,
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

	// Setup routes
	app.Get("/api/v0/health", s.EPHealth)
	app.Get("/api/v0/time-entries", s.EPTimeEntriesList)

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

// EPTimeEntriesList lists time entries
func (s Server) EPTimeEntriesList(c *fiber.Ctx) error {
	timeEntries, err := s.timeEntryRepo.List()
	if err != nil {
		return fmt.Errorf("failed to list time entries: %s", err)
	}

	return c.JSON(EPTimeEntriesListResp{
		TimeEntries: timeEntries,
	})
}

// EPTimeEntriesListResp is the list time entries endpoint response
type EPTimeEntriesListResp struct {
	// TimeEntries is the list of time entries
	TimeEntries []models.TimeEntry `json:"time_entries"`
}
