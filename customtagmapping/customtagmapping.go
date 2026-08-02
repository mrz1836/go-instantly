package customtagmapping

import (
	"context"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Custom Tag Mapping API.
const basePath = "/api/v2/custom-tag-mappings"

// Service provides access to the Instantly.ai V2 Custom Tag Mapping API.
type Service struct {
	client *instantly.Client
}

// New builds a Custom Tag Mapping API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// ResourceType is the kind of resource a tag is mapped to.
type ResourceType int64

// The kinds of resource a tag can be mapped to.
const (
	// ResourceTypeAccount is a sending account.
	ResourceTypeAccount ResourceType = 1

	// ResourceTypeCampaign is a campaign.
	ResourceTypeCampaign ResourceType = 2
)

// Mapping is a single association between a custom tag and a resource.
type Mapping struct {
	// ID is the unique identifier of the mapping.
	ID string `json:"id"`

	// TagID identifies the custom tag.
	TagID string `json:"tag_id"`

	// ResourceID identifies the resource the tag is assigned to.
	ResourceID string `json:"resource_id"`

	// ResourceType is the kind of resource the tag is assigned to.
	ResourceType ResourceType `json:"resource_type"`

	// OrganizationID identifies the organization the mapping belongs to.
	OrganizationID string `json:"organization_id"`

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

// ListResponse is a single page of custom tag mappings.
//
// It aliases instantly.Page[Mapping], the cursor-paginated envelope every
// resource shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[Mapping]

// List returns a single page of custom tag mappings filtered by the supplied
// options.
//
// Pagination is cursor based: pass the returned NextStartingAfter back with
// WithStartingAfter to fetch the following page.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}
