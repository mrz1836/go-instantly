package leadlist

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Lead List API.
const basePath = "/api/v2/lead-lists"

// Service provides access to the Instantly.ai V2 Lead List API.
type Service struct {
	client *instantly.Client
}

// New builds a Lead List API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// LeadList is a single lead list returned by the Instantly.ai V2 API.
type LeadList struct {
	// ID is the unique identifier of the list.
	ID string `json:"id"`

	// Name is the name of the list.
	Name string `json:"name"`

	// OrganizationID is the organization the list belongs to.
	OrganizationID string `json:"organization_id"`

	// TimestampCreated is when the list was created.
	TimestampCreated string `json:"timestamp_created"`

	// HasEnrichmentTask reports whether the list has an enrichment task.
	HasEnrichmentTask *bool `json:"has_enrichment_task,omitempty"`

	// OwnedBy is the owner ID of the list.
	OwnedBy *string `json:"owned_by,omitempty"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded list re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (l *LeadList) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, l.TimestampCreated)
}

// ListResponse is a single page of lead lists.
//
// It aliases instantly.Page[LeadList], the cursor-paginated envelope every
// resource shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[LeadList]

// VerificationStats are the verification statistics of a lead list.
type VerificationStats struct {
	// TotalLeads is the total number of leads in the list.
	TotalLeads int64 `json:"total_leads"`

	// Stats carries the per-status verification breakdown, preserved verbatim.
	Stats json.RawMessage `json:"stats,omitempty"`
}

// CreateRequest is the body of a create-lead-list request.
type CreateRequest struct {
	// Name is the name of the list. Required.
	Name string `json:"name"`

	// HasEnrichmentTask attaches an enrichment task to the list when set.
	HasEnrichmentTask *bool `json:"has_enrichment_task,omitempty"`

	// OwnedBy is the owner ID to assign the list to.
	OwnedBy *string `json:"owned_by,omitempty"`
}

// UpdateRequest is the body of a patch-lead-list request. No field is required;
// an omitted field leaves the current value unchanged.
type UpdateRequest struct {
	// Name is the name of the list.
	Name string `json:"name,omitempty"`

	// HasEnrichmentTask attaches an enrichment task to the list when set.
	HasEnrichmentTask *bool `json:"has_enrichment_task,omitempty"`

	// OwnedBy is the owner ID to assign the list to.
	OwnedBy *string `json:"owned_by,omitempty"`
}

// Create adds a new lead list and returns it.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*LeadList, error) {
	return instantly.PostResult[LeadList](ctx, s.client, basePath, req)
}

// List returns a single page of lead lists filtered by the supplied options.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}

// Get returns a single lead list by its unique identifier.
func (s *Service) Get(ctx context.Context, id string) (*LeadList, error) {
	return instantly.GetResult[LeadList](ctx, s.client, instantly.JoinPath(basePath, id))
}

// Update patches a lead list and returns its updated state.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*LeadList, error) {
	return instantly.PatchResult[LeadList](ctx, s.client, instantly.JoinPath(basePath, id), req)
}

// Delete deletes a lead list and returns the list that was deleted.
func (s *Service) Delete(ctx context.Context, id string) (*LeadList, error) {
	return instantly.DeleteResult[LeadList](ctx, s.client, instantly.JoinPath(basePath, id))
}

// VerificationStats returns the verification statistics of a lead list.
func (s *Service) VerificationStats(ctx context.Context, id string) (*VerificationStats, error) {
	return instantly.GetResult[VerificationStats](ctx, s.client, instantly.JoinPath(basePath, id, "verification-stats"))
}
