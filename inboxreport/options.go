package inboxreport

import "github.com/mrz1836/go-instantly"

// ListOption customizes a List request.
//
// Options are typed per resource, so passing an option from another resource is
// a compile error. Only the options actually supplied are sent.
type ListOption func(*instantly.Query)

// WithLimit sets the maximum number of reports returned in a single page.
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

// WithDateFrom restricts results to reports on or after the given date.
func WithDateFrom(date string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("date_from", date)
	}
}

// WithDateTo restricts results to reports on or before the given date.
func WithDateTo(date string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("date_to", date)
	}
}

// WithSkipSpamAssassinReport trims the spam_assassin_report payload from each
// report when true, which keeps the response smaller.
func WithSkipSpamAssassinReport(skip bool) ListOption {
	return func(q *instantly.Query) {
		q.SetBool("skip_spam_assassin_report", skip)
	}
}

// WithSkipBlacklistReport trims the blacklist_report payload from each report
// when true, which keeps the response smaller.
func WithSkipBlacklistReport(skip bool) ListOption {
	return func(q *instantly.Query) {
		q.SetBool("skip_blacklist_report", skip)
	}
}
