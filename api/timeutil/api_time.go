package timeutil

import (
	"fmt"
	"time"
)

// APITime is a time structure used for all time operations in the system
// This takes care of a few concerns:
// - Always storing data as UTC
// - Rounding times to the precision used by the DB
type APITime struct {
	time.Time
}

// NewAPITime creates a new APITime
func NewAPITime(inTime time.Time) (*APITime, error) {
	if inTime.Location() == nil {
		return nil, fmt.Errorf("input time does not have a location set")
	}

	storeTime := RoundForPG(inTime.In(time.UTC))

	return &APITime{
		Time: storeTime,
	}, nil
}

// ParseFormat parses a time string based on a time format
func ParseFormat(format string, inVal string) (*APITime, error) {
	inTime, err := time.Parse(format, inVal)
	if err != nil {
		return nil, err
	}

	return NewAPITime(inTime)
}

func (t *APITime) MakeDateOnly() *APITime {
	return &APITime{
		Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()),
	}
}

func RoundForPG(inTime time.Time) time.Time {
	return inTime.Round(time.Millisecond)
}
