package inboxtest

import "github.com/mrz1836/go-instantly"

// ListOption customizes a List request.
//
// Options are typed per resource, so passing an option from another resource is
// a compile error. Only the options actually supplied are sent.
type ListOption func(*instantly.Query)

// WithLimit sets the maximum number of tests returned in a single page.
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

// WithSearch restricts results to tests matching a search term.
func WithSearch(term string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("search", term)
	}
}

// WithStatus restricts results to tests with the given status.
func WithStatus(status Status) ListOption {
	return func(q *instantly.Query) {
		q.SetInt("status", int(status))
	}
}

// WithSortOrder sets the direction results are sorted in.
func WithSortOrder(order instantly.SortOrder) ListOption {
	return func(q *instantly.Query) {
		instantly.SetEnum(q, "sort_order", order)
	}
}

// GetOption customizes a Get request.
type GetOption func(*instantly.Query)

// WithMetadata includes the associated campaign and tag details in the test's
// Metadata field.
func WithMetadata() GetOption {
	return func(q *instantly.Query) {
		q.SetBool("with_metadata", true)
	}
}
