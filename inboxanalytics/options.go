package inboxanalytics

import "github.com/mrz1836/go-instantly"

// ListOption customizes a List request.
//
// Options are typed per resource, so passing an option from another resource is
// a compile error. Only the options actually supplied are sent.
//
// The recipient filters are string typed because the API accepts them here as
// comma-joined lists of integer codes on the query string; the typed enum slices
// live on the stats POST bodies instead.
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

// WithDateFrom restricts results to events on or after the given date.
func WithDateFrom(date string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("date_from", date)
	}
}

// WithDateTo restricts results to events on or before the given date.
func WithDateTo(date string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("date_to", date)
	}
}

// WithRecipientGeo restricts results to the given recipient regions, sent as a
// comma-separated list of integer codes.
func WithRecipientGeo(geos string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("recipient_geo", geos)
	}
}

// WithRecipientType restricts results to the given recipient types, sent as a
// comma-separated list of integer codes.
func WithRecipientType(types string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("recipient_type", types)
	}
}

// WithRecipientESP restricts results to the given recipient providers, sent as a
// comma-separated list of integer codes.
func WithRecipientESP(esps string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("recipient_esp", esps)
	}
}

// WithSenderEmail restricts results to a single sender address.
func WithSenderEmail(email string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("sender_email", email)
	}
}
