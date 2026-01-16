package valueobject

import (
	"fmt"
	"time"
)

// DateRange represents a time period for backtesting
type DateRange struct {
	StartDate time.Time
	EndDate   time.Time
}

// NewDateRange creates a new DateRange with validation
func NewDateRange(start, end time.Time) (DateRange, error) {
	if end.Before(start) {
		return DateRange{}, fmt.Errorf("end date must be after start date: start=%v, end=%v", start, end)
	}
	return DateRange{StartDate: start, EndDate: end}, nil
}

// MustNewDateRange creates a DateRange and panics on error
func MustNewDateRange(start, end time.Time) DateRange {
	dr, err := NewDateRange(start, end)
	if err != nil {
		panic(err)
	}
	return dr
}

// Duration returns the duration of the date range
func (dr DateRange) Duration() time.Duration {
	return dr.EndDate.Sub(dr.StartDate)
}

// Days returns the number of days in the date range
func (dr DateRange) Days() int {
	return int(dr.Duration().Hours() / 24)
}

// Contains checks if a date falls within the range
func (dr DateRange) Contains(date time.Time) bool {
	return (date.Equal(dr.StartDate) || date.After(dr.StartDate)) &&
		(date.Equal(dr.EndDate) || date.Before(dr.EndDate))
}

// String returns a string representation of the date range
func (dr DateRange) String() string {
	return fmt.Sprintf("%s to %s", dr.StartDate.Format("2006-01-02"), dr.EndDate.Format("2006-01-02"))
}
