package leadlist

import (
	"github.com/mrz1836/go-instantly"
)

// ListOption customizes a List request.
//
// Options are typed per resource, so passing an option from another resource is
// a compile error. Only the options actually supplied are sent.
type ListOption func(*instantly.Query)

// WithLimit sets the maximum number of lists returned in a single page.
func WithLimit(limit int) ListOption {
	return func(q *instantly.Query) {
		q.SetInt("limit", limit)
	}
}

// WithStartingAfter sets the pagination cursor to resume from.
func WithStartingAfter(cursor string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("starting_after", cursor)
	}
}

// WithSearch restricts results to lists matching a search term.
func WithSearch(term string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("search", term)
	}
}

// WithHasEnrichmentTask restricts results to lists that have (or lack) an
// enrichment task.
func WithHasEnrichmentTask(hasTask bool) ListOption {
	return func(q *instantly.Query) {
		q.SetBool("has_enrichment_task", hasTask)
	}
}
