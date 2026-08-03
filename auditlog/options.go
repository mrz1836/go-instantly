package auditlog

import (
	"time"

	"github.com/mrz1836/go-instantly"
)

// ListOption customizes a List request.
//
// Options are typed per resource, so passing an option from another resource is
// a compile error. Only the options actually supplied are sent.
type ListOption func(*instantly.Query)

// WithLimit sets the maximum number of records returned in a single page.
func WithLimit(limit int) ListOption {
	return func(q *instantly.Query) {
		q.SetInt("limit", limit)
	}
}

// WithStartingAfter sets the pagination cursor to resume from, which is the
// NextStartingAfter value of a previous page.
func WithStartingAfter(cursor string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("starting_after", cursor)
	}
}

// WithActivityType restricts results to records of a single activity type.
func WithActivityType(activityType ActivityType) ListOption {
	return func(q *instantly.Query) {
		q.SetInt("activity_type", int(activityType))
	}
}

// WithSearch restricts results to records matching a search term.
func WithSearch(term string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("search", term)
	}
}

// WithStartDate restricts results to records on or after the given date. The
// date is sent as a YYYY-MM-DD wire value.
func WithStartDate(date time.Time) ListOption {
	return func(q *instantly.Query) {
		q.SetString("start_date", date.Format(time.DateOnly))
	}
}

// WithEndDate restricts results to records on or before the given date. The
// date is sent as a YYYY-MM-DD wire value.
func WithEndDate(date time.Time) ListOption {
	return func(q *instantly.Query) {
		q.SetString("end_date", date.Format(time.DateOnly))
	}
}
