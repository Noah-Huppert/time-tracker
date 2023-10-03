package timeutil

import (
	"database/sql/driver"
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

	storeTime := inTime.In(time.UTC)
	storeTime = storeTime.Round(time.Millisecond)

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

func (t *APITime) Value() (driver.Value, error) {
	return t.Time, nil
}

func (t *APITime) Scan(value interface{}) error {
	inTime, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("could not parse")
	}

	newTime, err := NewAPITime(inTime)
	if err != nil {
		return err
	}

	*t = *newTime
	return nil
	/* if strValue, ok := value.(string); ok {
		outTime, err := pq.ParseTimestamp(time.UTC, strValue)
		if err != nil {
			return fmt.Errorf("failed to parse time: %s", err)
		}

		newTime, err := NewAPITime(outTime)
		if err != nil {
			return fmt.Errorf("failed to create APITime: %s", err)
		}

		*t = *newTime
	}

	return fmt.Errorf("value must be a string") */
}
