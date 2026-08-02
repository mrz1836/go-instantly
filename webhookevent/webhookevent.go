package webhookevent

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Webhook Event API.
const basePath = "/api/v2/webhook-events"

// Service provides access to the Instantly.ai V2 Webhook Event API.
type Service struct {
	client *instantly.Client
}

// New builds a Webhook Event API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Event is a single webhook delivery event returned by the Instantly.ai V2 API.
//
// Fields the API declares as nullable are pointers so an absent value stays
// distinguishable from a zero value. The delivered payload is free-form,
// preserved verbatim as json.RawMessage.
type Event struct {
	// ID is the unique identifier of the event.
	ID string `json:"id"`

	// OrganizationID identifies the organization the event belongs to.
	OrganizationID string `json:"organization_id"`

	// WebhookURL is the URL the event was delivered to.
	WebhookURL string `json:"webhook_url"`

	// TimestampCreated is when the event was created.
	TimestampCreated string `json:"timestamp_created"`

	// TimestampCreatedDate is the calendar date the event was created on.
	TimestampCreatedDate string `json:"timestamp_created_date"`

	// Success reports whether the delivery succeeded.
	Success bool `json:"success"`

	// RetryCount is how many times delivery has been retried.
	RetryCount float64 `json:"retry_count"`

	// WillRetry reports whether the delivery will be retried.
	WillRetry bool `json:"will_retry"`

	// LeadEmail is the address of the lead the event relates to.
	LeadEmail *string `json:"lead_email,omitempty"`

	// StatusCode is the HTTP status the target responded with.
	StatusCode *float64 `json:"status_code,omitempty"`

	// ResponseTimeMS is how long the target took to respond, in milliseconds.
	ResponseTimeMS *float64 `json:"response_time_ms,omitempty"`

	// ErrorMessage is the error detail when the delivery failed.
	ErrorMessage *string `json:"error_message,omitempty"`

	// RetryGroupID groups the retries of a single delivery.
	RetryGroupID *string `json:"retry_group_id,omitempty"`

	// RetrySuccessful reports whether a retry ultimately succeeded.
	RetrySuccessful *bool `json:"retry_successful,omitempty"`

	// TimestampNextRetry is when the next retry is scheduled.
	TimestampNextRetry *string `json:"timestamp_next_retry,omitempty"`

	// Payload carries the delivered event payload, which the API models as a
	// free-form object, so it is preserved verbatim.
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded event re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (e *Event) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, e.TimestampCreated)
}

// ListResponse is a single page of webhook events.
//
// It aliases instantly.Page[Event], the cursor-paginated envelope every resource
// shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[Event]

// Summary is the aggregate delivery outcome across webhook events.
type Summary struct {
	// TotalEvents is the total number of events.
	TotalEvents float64 `json:"total_events"`

	// SuccessfulEvents is the number of successful deliveries.
	SuccessfulEvents float64 `json:"successful_events"`

	// FailedEvents is the number of failed deliveries.
	FailedEvents float64 `json:"failed_events"`

	// SuccessRate is the fraction of deliveries that succeeded.
	SuccessRate float64 `json:"success_rate"`

	// FailureRate is the fraction of deliveries that failed.
	FailureRate float64 `json:"failure_rate"`
}

// DateSummary is the aggregate delivery outcome for a single calendar date.
type DateSummary struct {
	// Date is the calendar date the summary is for.
	Date string `json:"date"`

	// TotalEvents is the total number of events that day.
	TotalEvents float64 `json:"total_events"`

	// SuccessfulEvents is the number of successful deliveries that day.
	SuccessfulEvents float64 `json:"successful_events"`

	// FailedEvents is the number of failed deliveries that day.
	FailedEvents float64 `json:"failed_events"`

	// SuccessRate is the fraction of deliveries that succeeded that day.
	SuccessRate float64 `json:"success_rate"`
}

// summaryByDateResponse is the {"items":[...]} wrapper the summary-by-date
// endpoint returns, unwrapped to a plain slice for the caller.
type summaryByDateResponse struct {
	Items []DateSummary `json:"items"`
}

// List returns a single page of webhook events filtered by the supplied options.
//
// Pagination is cursor based: pass the returned NextStartingAfter back with
// WithStartingAfter to fetch the following page.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}

// Get returns a single webhook event by its unique identifier.
func (s *Service) Get(ctx context.Context, id string) (*Event, error) {
	return instantly.GetResult[Event](ctx, s.client, instantly.JoinPath(basePath, id))
}

// Summary returns the aggregate delivery outcome across webhook events.
//
// from and to bound the reporting window as YYYY-MM-DD dates; an empty value
// leaves that bound open.
func (s *Service) Summary(ctx context.Context, from, to string) (*Summary, error) {
	return instantly.GetResult[Summary](ctx, s.client, window(from, to).Path(basePath+"/summary"))
}

// SummaryByDate returns the aggregate delivery outcome broken down by date.
//
// from and to bound the reporting window as YYYY-MM-DD dates; an empty value
// leaves that bound open.
func (s *Service) SummaryByDate(ctx context.Context, from, to string) ([]DateSummary, error) {
	path := window(from, to).Path(basePath + "/summary-by-date")

	out, err := instantly.GetResult[summaryByDateResponse](ctx, s.client, path)
	if err != nil {
		return nil, err
	}

	return out.Items, nil
}

// window builds the from/to query shared by the summary endpoints, setting only
// the bounds that were supplied.
func window(from, to string) *instantly.Query {
	q := instantly.NewQuery()
	if from != "" {
		q.SetString("from", from)
	}
	if to != "" {
		q.SetString("to", to)
	}

	return q
}
