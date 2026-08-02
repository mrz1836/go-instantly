package customtag

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Custom Tag API.
const basePath = "/api/v2/custom-tags"

// Service provides access to the Instantly.ai V2 Custom Tag API.
type Service struct {
	client *instantly.Client
}

// New builds a Custom Tag API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// ResourceType is the kind of resource a tag can be assigned to.
type ResourceType int64

// The kinds of resource a tag can be assigned to.
const (
	// ResourceTypeAccount is a sending account.
	ResourceTypeAccount ResourceType = 1

	// ResourceTypeCampaign is a campaign.
	ResourceTypeCampaign ResourceType = 2
)

// Tag is a single custom tag returned by the Instantly.ai V2 API.
//
// Fields the API declares as nullable are pointers so an absent value stays
// distinguishable from a zero value.
type Tag struct {
	// ID is the unique identifier of the tag.
	ID string `json:"id"`

	// Label is the display label of the tag.
	Label string `json:"label"`

	// OrganizationID identifies the organization the tag belongs to.
	OrganizationID string `json:"organization_id"`

	// TimestampCreated is when the tag was created.
	TimestampCreated string `json:"timestamp_created"`

	// TimestampUpdated is when the tag was last updated.
	TimestampUpdated string `json:"timestamp_updated"`

	// Description is the detailed description of the tag.
	Description *string `json:"description,omitempty"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded tag re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (t *Tag) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, t.TimestampCreated)
}

// ListResponse is a single page of custom tags.
//
// It aliases instantly.Page[Tag], the cursor-paginated envelope every resource
// shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[Tag]

// ToggleResult reports whether a toggle-resource request succeeded.
type ToggleResult struct {
	// Success reports whether the assignment change succeeded.
	Success bool `json:"success"`
}

// CreateRequest is the body of a create-custom-tag request.
type CreateRequest struct {
	// Label is the display label of the tag. Required.
	Label string `json:"label"`

	// Description is the detailed description of the tag.
	Description *string `json:"description,omitempty"`
}

// UpdateRequest is the body of a patch-custom-tag request. No field is required;
// an omitted field leaves the current value unchanged.
type UpdateRequest struct {
	// Label is the display label of the tag.
	Label string `json:"label,omitempty"`

	// Description is the detailed description of the tag.
	Description *string `json:"description,omitempty"`
}

// ToggleRequest is the body of an assign-or-unassign-tags request.
//
// Target resources either by listing them in ResourceIDs, or by setting
// SelectedAll with an optional Search, Filter, and ExcludedResourceIDs to affect
// everything matching except the exclusions.
type ToggleRequest struct {
	// TagIDs are the tags to assign or unassign. Required.
	TagIDs []string `json:"tag_ids"`

	// ResourceType is the kind of resource the tags apply to. Required.
	ResourceType ResourceType `json:"resource_type"`

	// Assign assigns the tags when true and unassigns them when false. Required.
	Assign bool `json:"assign"`

	// ResourceIDs are the specific resources to target.
	ResourceIDs []string `json:"resource_ids,omitempty"`

	// ExcludedResourceIDs are resources to skip when SelectedAll is set.
	ExcludedResourceIDs []string `json:"excluded_resource_ids,omitempty"`

	// SelectedAll targets every matching resource when true.
	SelectedAll *bool `json:"selected_all,omitempty"`

	// Search narrows the targeted resources when SelectedAll is set.
	Search string `json:"search,omitempty"`

	// Filter carries the optional selected-all filter, which the API models as a
	// free-form payload, so it is sent verbatim.
	Filter json.RawMessage `json:"filter,omitempty"`
}

// Create adds a new custom tag and returns it.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Tag, error) {
	return instantly.PostResult[Tag](ctx, s.client, basePath, req)
}

// List returns a single page of custom tags filtered by the supplied options.
//
// Pagination is cursor based: pass the returned NextStartingAfter back with
// WithStartingAfter to fetch the following page.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}

// Get returns a single custom tag by its unique identifier.
func (s *Service) Get(ctx context.Context, id string) (*Tag, error) {
	return instantly.GetResult[Tag](ctx, s.client, instantly.JoinPath(basePath, id))
}

// Update patches a custom tag and returns its updated state.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Tag, error) {
	return instantly.PatchResult[Tag](ctx, s.client, instantly.JoinPath(basePath, id), req)
}

// Delete deletes a custom tag and returns the tag that was deleted.
func (s *Service) Delete(ctx context.Context, id string) (*Tag, error) {
	return instantly.DeleteResult[Tag](ctx, s.client, instantly.JoinPath(basePath, id))
}

// Toggle assigns or unassigns tags to resources and reports the outcome.
func (s *Service) Toggle(ctx context.Context, req ToggleRequest) (*ToggleResult, error) {
	return instantly.PostResult[ToggleResult](ctx, s.client, basePath+"/toggle-resource", req)
}
