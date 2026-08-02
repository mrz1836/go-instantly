package customtag

import "github.com/mrz1836/go-instantly"

// ListOption customizes a List request.
//
// Options are typed per resource, so passing an option from another resource is
// a compile error. Only the options actually supplied are sent.
type ListOption func(*instantly.Query)

// WithLimit sets the maximum number of tags returned in a single page.
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

// WithSearch restricts results to tags matching a search term.
func WithSearch(term string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("search", term)
	}
}

// WithResourceIDs restricts results to tags assigned to the given resources,
// sent as a comma-separated list.
func WithResourceIDs(resourceIDs string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("resource_ids", resourceIDs)
	}
}

// WithTagIDs restricts results to the given tags, sent as a comma-separated list.
func WithTagIDs(tagIDs string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("tag_ids", tagIDs)
	}
}
