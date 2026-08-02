package inboxanalytics

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Inbox Placement Analytics API.
const basePath = "/api/v2/inbox-placement-analytics"

// Service provides access to the Instantly.ai V2 Inbox Placement Analytics API.
type Service struct {
	client *instantly.Client
}

// New builds an Inbox Placement Analytics API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// RecordType is whether a placement event records a sent or a received email.
type RecordType int64

// The record types a placement event can carry.
const (
	// RecordTypeSent records an email that was sent.
	RecordTypeSent RecordType = 1

	// RecordTypeReceived records an email that was received.
	RecordTypeReceived RecordType = 2
)

// ESP identifies an email service provider on a placement event. The values are
// non-contiguous, mirroring the API's numbering.
type ESP int64

// The email service providers a placement event can involve.
const (
	// ESPGoogle is Google.
	ESPGoogle ESP = 1

	// ESPMicrosoft is Microsoft.
	ESPMicrosoft ESP = 2

	// ESPAirMail is AirMail.
	ESPAirMail ESP = 8

	// ESPWebDE is Web.de.
	ESPWebDE ESP = 12

	// ESPLibero is Libero.it.
	ESPLibero ESP = 13
)

// Geo identifies the geographic region of a recipient on a placement event.
type Geo int64

// The geographic regions a recipient can be in.
const (
	// GeoUS is the United States.
	GeoUS Geo = 1

	// GeoItaly is Italy.
	GeoItaly Geo = 2

	// GeoGermany is Germany.
	GeoGermany Geo = 3

	// GeoFrance is France.
	GeoFrance Geo = 4
)

// RecipientType is whether a recipient is a professional or personal address.
type RecipientType int64

// The recipient types a placement event can involve.
const (
	// RecipientProfessional is a professional address.
	RecipientProfessional RecipientType = 1

	// RecipientPersonal is a personal address.
	RecipientPersonal RecipientType = 2
)

// Analytics is a single inbox placement event returned by the Instantly.ai V2
// API.
//
// The recipient-side fields are populated only when the event records a received
// email (record_type of 2), so they are pointers: an absent value stays
// distinguishable from a zero value.
type Analytics struct {
	// ID is the unique identifier of the placement event.
	ID string `json:"id"`

	// OrganizationID identifies the organization the event belongs to.
	OrganizationID string `json:"organization_id"`

	// TestID identifies the inbox placement test the event belongs to.
	TestID string `json:"test_id"`

	// TimestampCreated is when the event was created.
	TimestampCreated string `json:"timestamp_created"`

	// TimestampCreatedDate is the calendar date the event was created on.
	TimestampCreatedDate string `json:"timestamp_created_date"`

	// RecordType is whether the event records a sent or received email.
	RecordType *RecordType `json:"record_type,omitempty"`

	// RecipientEmail is the recipient address of a received email.
	RecipientEmail *string `json:"recipient_email,omitempty"`

	// RecipientESP is the recipient's email service provider.
	RecipientESP *ESP `json:"recipient_esp,omitempty"`

	// RecipientGeo is the recipient's geographic region.
	RecipientGeo *Geo `json:"recipient_geo,omitempty"`

	// RecipientType is whether the recipient is professional or personal.
	RecipientType *RecipientType `json:"recipient_type,omitempty"`

	// SenderEmail is the sender address of a received email.
	SenderEmail *string `json:"sender_email,omitempty"`

	// SenderESP is the sender's email service provider.
	SenderESP *ESP `json:"sender_esp,omitempty"`

	// IsSpam reports whether the email landed in spam.
	IsSpam *bool `json:"is_spam,omitempty"`

	// HasCategory reports whether the email landed in a category tab.
	HasCategory *bool `json:"has_category,omitempty"`

	// DKIMPass reports whether the email passed DKIM.
	DKIMPass *bool `json:"dkim_pass,omitempty"`

	// DMARCPass reports whether the email passed DMARC.
	DMARCPass *bool `json:"dmarc_pass,omitempty"`

	// SPFPass reports whether the email passed SPF.
	SPFPass *bool `json:"spf_pass,omitempty"`

	// SMTPIPBlacklistReport carries the raw SMTP IP blacklist report, which the
	// API models as a free-form payload, so it is preserved verbatim.
	SMTPIPBlacklistReport json.RawMessage `json:"smtp_ip_blacklist_report,omitempty"`

	// AuthenticationFailureResults carries the raw authentication failure detail,
	// which the API models as a free-form payload, so it is preserved verbatim.
	AuthenticationFailureResults json.RawMessage `json:"authentication_failure_results,omitempty"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded event re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (a *Analytics) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, a.TimestampCreated)
}

// ListResponse is a single page of inbox placement events.
//
// It aliases instantly.Page[Analytics], the cursor-paginated envelope every
// resource shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[Analytics]

// List returns a single page of placement events for a test, filtered by the
// supplied options.
//
// A test_id filter is required, so it is a positional argument. Pagination is
// cursor based: pass the returned NextStartingAfter back with WithStartingAfter
// to fetch the following page.
func (s *Service) List(ctx context.Context, testID string, opts ...ListOption) (*ListResponse, error) {
	q := instantly.ApplyOptions(opts...).SetString("test_id", testID)

	return instantly.GetResult[ListResponse](ctx, s.client, q.Path(basePath))
}

// Get returns a single placement event by its unique identifier.
func (s *Service) Get(ctx context.Context, id string) (*Analytics, error) {
	return instantly.GetResult[Analytics](ctx, s.client, instantly.JoinPath(basePath, id))
}
