package inboxreport

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Inbox Placement Report API.
const basePath = "/api/v2/inbox-placement-reports"

// Service provides access to the Instantly.ai V2 Inbox Placement blacklist and
// SpamAssassin report API.
type Service struct {
	client *instantly.Client
}

// New builds an Inbox Placement Report API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Report is a single blacklist and SpamAssassin report returned by the
// Instantly.ai V2 API.
//
// Fields the API declares as nullable are pointers so an absent value stays
// distinguishable from a zero value. The nested blacklist and SpamAssassin
// reports are preserved verbatim as json.RawMessage — the API sends the
// SpamAssassin per-rule score as a string, which raw preservation keeps intact —
// and either can be trimmed from the payload with the skip options.
type Report struct {
	// ID is the unique identifier of the report.
	ID string `json:"id"`

	// OrganizationID identifies the organization the report belongs to.
	OrganizationID string `json:"organization_id"`

	// TestID identifies the inbox placement test the report belongs to.
	TestID string `json:"test_id"`

	// TimestampCreated is when the report was created.
	TimestampCreated string `json:"timestamp_created"`

	// TimestampCreatedDate is the calendar date the report was created on.
	TimestampCreatedDate string `json:"timestamp_created_date"`

	// Domain is the sending domain the report is for.
	Domain string `json:"domain"`

	// DomainIP is the sending domain's IP address.
	DomainIP string `json:"domain_ip"`

	// SpamAssassinScore is the overall SpamAssassin score for the message.
	SpamAssassinScore float64 `json:"spam_assassin_score"`

	// DomainBlacklistCount is how many blacklists the domain appears on.
	DomainBlacklistCount *float64 `json:"domain_blacklist_count,omitempty"`

	// DomainIPBlacklistCount is how many blacklists the domain's IP appears on.
	DomainIPBlacklistCount *float64 `json:"domain_ip_blacklist_count,omitempty"`

	// BlacklistReport carries the raw blacklist report, which the API models as a
	// nested payload, so it is preserved verbatim. It can be trimmed with
	// WithSkipBlacklistReport.
	BlacklistReport json.RawMessage `json:"blacklist_report,omitempty"`

	// SpamAssassinReport carries the raw SpamAssassin report, which the API models
	// as a nested payload whose per-rule score is a string, so it is preserved
	// verbatim. It can be trimmed with WithSkipSpamAssassinReport.
	SpamAssassinReport json.RawMessage `json:"spam_assassin_report,omitempty"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded report re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (r *Report) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, r.TimestampCreated)
}

// ListResponse is a single page of inbox placement reports.
//
// It aliases instantly.Page[Report], the cursor-paginated envelope every
// resource shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[Report]

// List returns a single page of reports for a test, filtered by the supplied
// options.
//
// A test_id filter is required, so it is a positional argument. Pagination is
// cursor based: pass the returned NextStartingAfter back with WithStartingAfter
// to fetch the following page.
func (s *Service) List(ctx context.Context, testID string, opts ...ListOption) (*ListResponse, error) {
	q := instantly.ApplyOptions(opts...).SetString("test_id", testID)

	return instantly.GetResult[ListResponse](ctx, s.client, q.Path(basePath))
}

// Get returns a single report by its unique identifier.
func (s *Service) Get(ctx context.Context, id string) (*Report, error) {
	return instantly.GetResult[Report](ctx, s.client, instantly.JoinPath(basePath, id))
}
