package webhook

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Webhook API.
const basePath = "/api/v2/webhooks"

// Service provides access to the Instantly.ai V2 Webhook API.
type Service struct {
	client *instantly.Client
}

// New builds a Webhook API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Status is the delivery status of a webhook.
type Status int64

// The statuses a webhook can be in.
const (
	// StatusActive means the webhook is active.
	StatusActive Status = 1

	// StatusError means the webhook was disabled due to delivery failures.
	StatusError Status = -1
)

// EventType is the campaign event a webhook subscribes to.
type EventType string

// The events a webhook can subscribe to.
const (
	// EventAll subscribes to every event.
	EventAll EventType = "all_events"

	// EventEmailSent fires when an email is sent.
	EventEmailSent EventType = "email_sent"

	// EventEmailOpened fires when an email is opened.
	EventEmailOpened EventType = "email_opened"

	// EventEmailLinkClicked fires when a link in an email is clicked.
	EventEmailLinkClicked EventType = "email_link_clicked"

	// EventReplyReceived fires when a reply is received.
	EventReplyReceived EventType = "reply_received"

	// EventEmailBounced fires when an email bounces.
	EventEmailBounced EventType = "email_bounced"

	// EventLeadUnsubscribed fires when a lead unsubscribes.
	EventLeadUnsubscribed EventType = "lead_unsubscribed"

	// EventCampaignCompleted fires when a campaign completes.
	EventCampaignCompleted EventType = "campaign_completed"

	// EventAccountError fires when a sending account errors.
	EventAccountError EventType = "account_error"

	// EventLeadNeutral fires when a lead is marked neutral.
	EventLeadNeutral EventType = "lead_neutral"

	// EventLeadInterested fires when a lead is marked interested.
	EventLeadInterested EventType = "lead_interested"

	// EventLeadNotInterested fires when a lead is marked not interested.
	EventLeadNotInterested EventType = "lead_not_interested"

	// EventLeadMeetingBooked fires when a lead books a meeting.
	EventLeadMeetingBooked EventType = "lead_meeting_booked"

	// EventLeadMeetingCompleted fires when a lead completes a meeting.
	EventLeadMeetingCompleted EventType = "lead_meeting_completed"

	// EventLeadClosed fires when a lead is closed.
	EventLeadClosed EventType = "lead_closed"

	// EventLeadOutOfOffice fires when a lead replies out of office.
	EventLeadOutOfOffice EventType = "lead_out_of_office"

	// EventLeadWrongPerson fires when a lead is the wrong person.
	EventLeadWrongPerson EventType = "lead_wrong_person"

	// EventLeadNoShow fires when a lead is a no-show.
	EventLeadNoShow EventType = "lead_no_show"

	// EventSuperSearchEnrichmentCompleted fires when a SuperSearch enrichment
	// completes.
	EventSuperSearchEnrichmentCompleted EventType = "supersearch_enrichment_completed"
)

// Webhook is a single webhook returned by the Instantly.ai V2 API.
//
// Fields the API declares as nullable are pointers so an absent value stays
// distinguishable from a zero value. The custom headers are a free-form map,
// preserved verbatim as json.RawMessage.
type Webhook struct {
	// ID is the unique identifier of the webhook.
	ID string `json:"id"`

	// Organization identifies the organization the webhook belongs to.
	Organization string `json:"organization"`

	// TargetHookURL is the URL events are delivered to.
	TargetHookURL string `json:"target_hook_url"`

	// TimestampCreated is when the webhook was created.
	TimestampCreated string `json:"timestamp_created"`

	// Name is the display name of the webhook.
	Name *string `json:"name,omitempty"`

	// EventType is the event the webhook subscribes to.
	EventType *EventType `json:"event_type,omitempty"`

	// Campaign identifies the campaign the webhook is scoped to.
	Campaign *string `json:"campaign,omitempty"`

	// CustomInterestValue is the interest value the webhook filters on.
	CustomInterestValue *float64 `json:"custom_interest_value,omitempty"`

	// Status is the delivery status of the webhook.
	Status *Status `json:"status,omitempty"`

	// TimestampError is when the webhook was last disabled by a delivery failure.
	TimestampError *string `json:"timestamp_error,omitempty"`

	// Headers carries the custom headers the webhook sends, which the API models
	// as a free-form map, so they are preserved verbatim.
	Headers json.RawMessage `json:"headers,omitempty"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded webhook re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (w *Webhook) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, w.TimestampCreated)
}

// ListResponse is a single page of webhooks.
//
// It aliases instantly.Page[Webhook], the cursor-paginated envelope every
// resource shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[Webhook]

// EventTypeInfo describes an event type a webhook can subscribe to.
type EventTypeInfo struct {
	// ID is the identifier of the event type.
	ID string `json:"id"`

	// Label is the human-readable label of the event type.
	Label string `json:"label"`

	// Type is the event type value used in a webhook subscription.
	Type string `json:"type"`
}

// eventTypesResponse is the {"event_types":[...]} wrapper the event-types
// endpoint returns, unwrapped to a plain slice for the caller.
type eventTypesResponse struct {
	EventTypes []EventTypeInfo `json:"event_types"`
}

// TestResult is the outcome of a webhook test delivery.
//
// StatusCode and ResponseTimeMS are pointers so an absent value — for example
// when the target could not be reached at all — stays distinguishable from a
// zero value.
type TestResult struct {
	// Success reports whether the test delivery succeeded.
	Success bool `json:"success"`

	// StatusCode is the HTTP status the target responded with.
	StatusCode *float64 `json:"status_code,omitempty"`

	// ResponseTimeMS is how long the target took to respond, in milliseconds.
	ResponseTimeMS *float64 `json:"response_time_ms,omitempty"`

	// Message is a human-readable outcome message.
	Message string `json:"message,omitempty"`

	// Error is the error detail when the delivery failed.
	Error string `json:"error,omitempty"`
}

// CreateRequest is the body of a create-webhook request.
type CreateRequest struct {
	// TargetHookURL is the URL events are delivered to. Required.
	TargetHookURL string `json:"target_hook_url"`

	// Name is the display name of the webhook.
	Name *string `json:"name,omitempty"`

	// EventType is the event to subscribe to.
	EventType EventType `json:"event_type,omitempty"`

	// Campaign scopes the webhook to a campaign.
	Campaign *string `json:"campaign,omitempty"`

	// CustomInterestValue is the interest value to filter on.
	CustomInterestValue *float64 `json:"custom_interest_value,omitempty"`

	// Headers carries the custom headers to send, preserved verbatim.
	Headers json.RawMessage `json:"headers,omitempty"`
}

// UpdateRequest is the body of a patch-webhook request. No field is required; an
// omitted field leaves the current value unchanged.
type UpdateRequest struct {
	// TargetHookURL is the URL events are delivered to.
	TargetHookURL string `json:"target_hook_url,omitempty"`

	// Name is the display name of the webhook.
	Name *string `json:"name,omitempty"`

	// EventType is the event to subscribe to.
	EventType EventType `json:"event_type,omitempty"`

	// Campaign scopes the webhook to a campaign.
	Campaign *string `json:"campaign,omitempty"`

	// CustomInterestValue is the interest value to filter on.
	CustomInterestValue *float64 `json:"custom_interest_value,omitempty"`

	// Headers carries the custom headers to send, preserved verbatim.
	Headers json.RawMessage `json:"headers,omitempty"`
}

// Create adds a new webhook and returns it.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Webhook, error) {
	return instantly.PostResult[Webhook](ctx, s.client, basePath, req)
}

// List returns a single page of webhooks filtered by the supplied options.
//
// Pagination is cursor based: pass the returned NextStartingAfter back with
// WithStartingAfter to fetch the following page.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}

// Get returns a single webhook by its unique identifier.
func (s *Service) Get(ctx context.Context, id string) (*Webhook, error) {
	return instantly.GetResult[Webhook](ctx, s.client, instantly.JoinPath(basePath, id))
}

// Update patches a webhook and returns its updated state.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Webhook, error) {
	return instantly.PatchResult[Webhook](ctx, s.client, instantly.JoinPath(basePath, id), req)
}

// Delete deletes a webhook and returns the webhook that was deleted.
func (s *Service) Delete(ctx context.Context, id string) (*Webhook, error) {
	return instantly.DeleteResult[Webhook](ctx, s.client, instantly.JoinPath(basePath, id))
}

// EventTypes returns the event types a webhook can subscribe to.
//
// The list is not paginated. It relies on the router matching this literal path
// ahead of the /webhooks/{id} route.
func (s *Service) EventTypes(ctx context.Context) ([]EventTypeInfo, error) {
	out, err := instantly.GetResult[eventTypesResponse](ctx, s.client, basePath+"/event-types")
	if err != nil {
		return nil, err
	}

	return out.EventTypes, nil
}

// Test sends a test delivery to a webhook and returns the outcome.
func (s *Service) Test(ctx context.Context, id string) (*TestResult, error) {
	return instantly.PostResult[TestResult](ctx, s.client, instantly.JoinPath(basePath, id, "test"), nil)
}

// Resume resumes a webhook that was disabled by delivery failures and returns it.
func (s *Service) Resume(ctx context.Context, id string) (*Webhook, error) {
	return instantly.PostResult[Webhook](ctx, s.client, instantly.JoinPath(basePath, id, "resume"), nil)
}
