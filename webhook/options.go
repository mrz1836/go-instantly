package webhook

import "github.com/mrz1836/go-instantly"

// ListOption customizes a List request.
//
// Options are typed per resource, so passing an option from another resource is
// a compile error. Only the options actually supplied are sent.
type ListOption func(*instantly.Query)

// WithLimit sets the maximum number of webhooks returned in a single page.
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

// WithCampaign restricts results to webhooks scoped to a campaign.
func WithCampaign(campaignID string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("campaign", campaignID)
	}
}

// WithEventType restricts results to webhooks subscribed to an event type.
func WithEventType(eventType EventType) ListOption {
	return func(q *instantly.Query) {
		instantly.SetEnum(q, "event_type", eventType)
	}
}
