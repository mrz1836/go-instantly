package webhookevent

import "github.com/mrz1836/go-instantly"

// ListOption customizes a List request.
//
// Options are typed per resource, so passing an option from another resource is
// a compile error. Only the options actually supplied are sent.
type ListOption func(*instantly.Query)

// WithLimit sets the maximum number of events returned in a single page.
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

// WithSuccess restricts results to successful deliveries when true, and to
// failed deliveries when false.
func WithSuccess(success bool) ListOption {
	return func(q *instantly.Query) {
		q.SetBool("success", success)
	}
}

// WithFrom restricts results to events on or after the given date, sent as a
// YYYY-MM-DD wire value.
func WithFrom(date string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("from", date)
	}
}

// WithTo restricts results to events on or before the given date, sent as a
// YYYY-MM-DD wire value.
func WithTo(date string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("to", date)
	}
}

// WithSearch restricts results to events matching an exact webhook URL or lead
// email.
func WithSearch(term string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("search", term)
	}
}
