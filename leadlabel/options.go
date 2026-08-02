package leadlabel

import (
	"github.com/mrz1836/go-instantly"
)

// InterestFilter filters labels by the interest category they map to.
type InterestFilter string

// The interest categories a list request can filter labels by.
const (
	// InterestPositive restricts results to positive-interest labels.
	InterestPositive InterestFilter = "positive"

	// InterestNeutral restricts results to neutral-interest labels.
	InterestNeutral InterestFilter = "neutral"

	// InterestNegative restricts results to negative-interest labels.
	InterestNegative InterestFilter = "negative"
)

// ListOption customizes a List request.
//
// Options are typed per resource, so passing an option from another resource is
// a compile error. Only the options actually supplied are sent.
type ListOption func(*instantly.Query)

// WithLimit sets the maximum number of labels returned in a single page.
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

// WithSearch restricts results to labels matching a search term.
func WithSearch(term string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("search", term)
	}
}

// WithInterestStatus restricts results to labels in an interest category.
func WithInterestStatus(status InterestFilter) ListOption {
	return func(q *instantly.Query) {
		instantly.SetEnum(q, "interest_status", status)
	}
}
