package inboxtest

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Inbox Placement Test API.
const basePath = "/api/v2/inbox-placement-tests"

// Service provides access to the Instantly.ai V2 Inbox Placement Test API.
type Service struct {
	client *instantly.Client
}

// New builds an Inbox Placement Test API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Type is whether an inbox placement test runs once or on a schedule.
type Type int64

// The kinds of inbox placement test.
const (
	// TypeOneTime is a one-time test.
	TypeOneTime Type = 1

	// TypeAutomated is an automated, recurring test.
	TypeAutomated Type = 2
)

// SendingMethod is where an inbox placement test's emails are sent from.
type SendingMethod int64

// The ways an inbox placement test can be sent.
const (
	// SendingFromInstantly sends the test from Instantly.
	SendingFromInstantly SendingMethod = 1

	// SendingFromOutside sends the test from outside Instantly.
	SendingFromOutside SendingMethod = 2
)

// DeliveryMode is whether a test's emails are sent one by one or all together.
type DeliveryMode int64

// The delivery modes an inbox placement test can use.
const (
	// DeliveryOneByOne sends the emails one by one.
	DeliveryOneByOne DeliveryMode = 1

	// DeliveryAllTogether sends the emails all together.
	DeliveryAllTogether DeliveryMode = 2
)

// Status is the current status of an inbox placement test.
type Status int64

// The statuses an inbox placement test can be in.
const (
	// StatusActive means the test is active.
	StatusActive Status = 1

	// StatusPaused means the test is paused.
	StatusPaused Status = 2

	// StatusCompleted means the test has completed.
	StatusCompleted Status = 3
)

// RecipientLabel identifies an email service provider and the audience a test
// sends to. The available combinations come from ESPOptions.
type RecipientLabel struct {
	// ESP is the email service provider to send to.
	ESP string `json:"esp"`

	// Region is the region to send to.
	Region string `json:"region"`

	// SubRegion is the sub-region to send to.
	SubRegion string `json:"sub_region"`

	// Type is the kind of audience to send to.
	Type string `json:"type"`
}

// Test is a single inbox placement test returned by the Instantly.ai V2 API.
//
// Fields the API declares as nullable are pointers so an absent value stays
// distinguishable from a zero value. The nested schedule, automations, and
// metadata payloads are preserved verbatim as json.RawMessage.
type Test struct {
	// ID is the unique identifier of the test.
	ID string `json:"id"`

	// OrganizationID identifies the organization the test belongs to.
	OrganizationID string `json:"organization_id"`

	// Name is the name of the test.
	Name string `json:"name"`

	// Type is whether the test is one-time or automated.
	Type Type `json:"type"`

	// SendingMethod is where the test's emails are sent from.
	SendingMethod SendingMethod `json:"sending_method"`

	// EmailSubject is the subject line of the test emails.
	EmailSubject string `json:"email_subject"`

	// EmailBody is the body of the test emails.
	EmailBody string `json:"email_body"`

	// Emails are the addresses the test is sent to.
	Emails []string `json:"emails"`

	// Recipients are the seed addresses the test measures placement against.
	Recipients []string `json:"recipients"`

	// TimestampCreated is when the test was created.
	TimestampCreated string `json:"timestamp_created"`

	// CampaignID identifies the campaign the test belongs to.
	CampaignID *string `json:"campaign_id,omitempty"`

	// DeliveryMode is whether emails are sent one by one or all together.
	DeliveryMode *DeliveryMode `json:"delivery_mode,omitempty"`

	// Description is the description of the test.
	Description *string `json:"description,omitempty"`

	// NotSendingStatus is why the test is currently not sending, and is an empty
	// string when there is no issue.
	NotSendingStatus *string `json:"not_sending_status,omitempty"`

	// Status is the current status of the test.
	Status *Status `json:"status,omitempty"`

	// Tags are the tag IDs the test uses for sending.
	Tags []string `json:"tags,omitempty"`

	// TestCode identifies tests sent from outside Instantly.
	TestCode *string `json:"test_code,omitempty"`

	// TextOnly disables open tracking when true.
	TextOnly *bool `json:"text_only,omitempty"`

	// TimestampNextRun is when the test will run next.
	TimestampNextRun *string `json:"timestamp_next_run,omitempty"`

	// RecipientsLabels are the provider/audience combinations the test sends to.
	RecipientsLabels []RecipientLabel `json:"recipients_labels,omitempty"`

	// Schedule carries the automated-test schedule, which the API models as a
	// deeply nested payload, so it is preserved verbatim.
	Schedule json.RawMessage `json:"schedule,omitempty"`

	// Automations carries the test's condition-based automations, which the API
	// models as a free-form payload, so it is preserved verbatim.
	Automations json.RawMessage `json:"automations,omitempty"`

	// Metadata carries the associated campaign and tag details, present only when
	// the test is fetched with WithMetadata. It is preserved verbatim.
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded test re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (t *Test) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, t.TimestampCreated)
}

// ListResponse is a single page of inbox placement tests.
//
// It aliases instantly.Page[Test], the cursor-paginated envelope every resource
// shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[Test]

// CreateRequest is the body of a create-inbox-placement-test request.
type CreateRequest struct {
	// Name is the name of the test. Required.
	Name string `json:"name"`

	// Type is whether the test is one-time or automated. Required.
	Type Type `json:"type"`

	// SendingMethod is where the test's emails are sent from. Required.
	SendingMethod SendingMethod `json:"sending_method"`

	// EmailSubject is the subject line of the test emails. Required.
	EmailSubject string `json:"email_subject"`

	// EmailBody is the body of the test emails. Required.
	EmailBody string `json:"email_body"`

	// Emails are the addresses the test is sent to. Required.
	Emails []string `json:"emails"`

	// Description is the description of the test.
	Description *string `json:"description,omitempty"`

	// CampaignID identifies the campaign the test belongs to.
	CampaignID *string `json:"campaign_id,omitempty"`

	// DeliveryMode is whether emails are sent one by one or all together.
	DeliveryMode *DeliveryMode `json:"delivery_mode,omitempty"`

	// Status is the status to create the test in.
	Status *Status `json:"status,omitempty"`

	// NotSendingStatus is why the test is not sending.
	NotSendingStatus *string `json:"not_sending_status,omitempty"`

	// RecipientsLabels are the provider/audience combinations to send to.
	RecipientsLabels []RecipientLabel `json:"recipients_labels,omitempty"`

	// RunImmediately runs the test as soon as it is created when true.
	RunImmediately *bool `json:"run_immediately,omitempty"`

	// Schedule carries the automated-test schedule, sent verbatim.
	Schedule json.RawMessage `json:"schedule,omitempty"`

	// Tags are the tag IDs the test uses for sending.
	Tags []string `json:"tags,omitempty"`

	// TestCode identifies tests sent from outside Instantly.
	TestCode *string `json:"test_code,omitempty"`

	// TextOnly disables open tracking when true.
	TextOnly *bool `json:"text_only,omitempty"`

	// TimestampNextRun is when the test will run next.
	TimestampNextRun *string `json:"timestamp_next_run,omitempty"`

	// Automations carries the test's condition-based automations, sent verbatim.
	Automations json.RawMessage `json:"automations,omitempty"`
}

// UpdateRequest is the body of a patch-inbox-placement-test request. No field is
// required; an omitted field leaves the current value unchanged.
type UpdateRequest struct {
	// Name is the name of the test.
	Name string `json:"name,omitempty"`

	// Status is the status to move the test to.
	Status *Status `json:"status,omitempty"`

	// Schedule carries the automated-test schedule, sent verbatim.
	Schedule json.RawMessage `json:"schedule,omitempty"`

	// Automations carries the test's condition-based automations, sent verbatim.
	Automations json.RawMessage `json:"automations,omitempty"`
}

// Create adds a new inbox placement test and returns it.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Test, error) {
	return instantly.PostResult[Test](ctx, s.client, basePath, req)
}

// List returns a single page of inbox placement tests filtered by the supplied
// options.
//
// Pagination is cursor based: pass the returned NextStartingAfter back with
// WithStartingAfter to fetch the following page.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}

// Get returns a single inbox placement test by its unique identifier.
//
// Pass WithMetadata to include the associated campaign and tag details in the
// test's Metadata field.
func (s *Service) Get(ctx context.Context, id string, opts ...GetOption) (*Test, error) {
	path := instantly.ApplyOptions(opts...).Path(instantly.JoinPath(basePath, id))

	return instantly.GetResult[Test](ctx, s.client, path)
}

// Update patches an inbox placement test and returns its updated state.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Test, error) {
	return instantly.PatchResult[Test](ctx, s.client, instantly.JoinPath(basePath, id), req)
}

// Delete deletes an inbox placement test and returns the test that was deleted.
func (s *Service) Delete(ctx context.Context, id string) (*Test, error) {
	return instantly.DeleteResult[Test](ctx, s.client, instantly.JoinPath(basePath, id))
}

// ESPOptions returns the email service provider options an inbox placement test
// can target, which are the valid values for a RecipientLabel.
func (s *Service) ESPOptions(ctx context.Context) ([]RecipientLabel, error) {
	var out []RecipientLabel
	if err := s.client.Get(ctx, basePath+"/email-service-provider-options", &out); err != nil {
		return nil, err
	}

	return out, nil
}
