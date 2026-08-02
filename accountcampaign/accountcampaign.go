package accountcampaign

import (
	"context"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the account-campaign-mappings API.
const basePath = "/api/v2/account-campaign-mappings"

// Service provides access to the account-campaign-mappings endpoint.
type Service struct {
	client *instantly.Client
}

// New builds an account-campaign-mappings service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Status is the current status of a campaign a mapping points to. It mirrors the
// campaign statuses in the campaign package.
type Status int64

// The statuses a mapped campaign can be in.
const (
	// StatusDraft means the campaign is a draft that has not launched.
	StatusDraft Status = 0

	// StatusActive means the campaign is active.
	StatusActive Status = 1

	// StatusPaused means the campaign is paused.
	StatusPaused Status = 2

	// StatusCompleted means the campaign has completed.
	StatusCompleted Status = 3

	// StatusRunningSubsequences means the campaign is running subsequences.
	StatusRunningSubsequences Status = 4

	// StatusAccountsUnhealthy means the campaign's sending accounts are unhealthy.
	StatusAccountsUnhealthy Status = -1

	// StatusBounceProtect means the campaign is held by bounce protection.
	StatusBounceProtect Status = -2

	// StatusAccountSuspended means the campaign's account is suspended.
	StatusAccountSuspended Status = -99
)

// Mapping is a single campaign a sending account is associated with.
type Mapping struct {
	// CampaignID is the unique identifier of the campaign.
	CampaignID string `json:"campaign_id"`

	// CampaignName is the name of the campaign.
	CampaignName string `json:"campaign_name"`

	// Status is the status of the campaign.
	Status Status `json:"status"`

	// TimestampCreated is when the mapping was created.
	TimestampCreated string `json:"timestamp_created"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded mapping re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (m *Mapping) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, m.TimestampCreated)
}

// ListResponse is a single page of account-campaign mappings.
//
// It aliases instantly.Page[Mapping], the cursor-paginated envelope every
// resource shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[Mapping]

// List returns a single page of campaigns associated with the given account
// email.
func (s *Service) List(ctx context.Context, email string, opts ...ListOption) (*ListResponse, error) {
	path := instantly.ApplyOptions(opts...).Path(instantly.JoinPath(basePath, email))

	return instantly.GetResult[ListResponse](ctx, s.client, path)
}
