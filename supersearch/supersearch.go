package supersearch

import (
	"context"
	"encoding/json"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the SuperSearch Enrichment API.
const basePath = "/api/v2/supersearch-enrichment"

// Service provides access to the Instantly.ai V2 SuperSearch Enrichment API.
type Service struct {
	client *instantly.Client
}

// New builds a SuperSearch Enrichment API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// ResourceType is whether an enrichment targets a campaign or a lead list.
type ResourceType int64

// The kinds of resource an enrichment can target.
const (
	// ResourceTypeCampaign targets a campaign.
	ResourceTypeCampaign ResourceType = 1

	// ResourceTypeList targets a lead list.
	ResourceTypeList ResourceType = 2
)

// EnrichmentType is the kind of enrichment to perform.
type EnrichmentType string

// The kinds of enrichment the API supports.
const (
	// EnrichmentWorkEmail enriches work email addresses.
	EnrichmentWorkEmail EnrichmentType = "work_email_enrichment"

	// EnrichmentFullProfile enriches a fully enriched profile.
	EnrichmentFullProfile EnrichmentType = "fully_enriched_profile"

	// EnrichmentEmailVerification verifies email addresses.
	EnrichmentEmailVerification EnrichmentType = "email_verification"

	// EnrichmentJobListing enriches job listings.
	EnrichmentJobListing EnrichmentType = "joblisting"

	// EnrichmentTechnologies enriches the technologies a company uses.
	EnrichmentTechnologies EnrichmentType = "technologies"

	// EnrichmentNews enriches company news.
	EnrichmentNews EnrichmentType = "news"

	// EnrichmentFunding enriches company funding.
	EnrichmentFunding EnrichmentType = "funding"

	// EnrichmentEngagementScore enriches an engagement score.
	EnrichmentEngagementScore EnrichmentType = "engagement_score"

	// EnrichmentAI runs an AI enrichment.
	EnrichmentAI EnrichmentType = "ai_enrichment"

	// EnrichmentCustomFlow runs a custom enrichment flow.
	EnrichmentCustomFlow EnrichmentType = "custom_flow"
)

// Enrichment is a SuperSearch enrichment record.
//
// It is the shape shared by the create, run, and update-settings responses, so
// fields a given endpoint does not return stay nil or zero: a nil ResourceType
// means the endpoint reported nothing, which is not the same as reporting a
// value. The enrichment payload is preserved verbatim as json.RawMessage.
type Enrichment struct {
	// ID is the unique identifier of the enrichment.
	ID string `json:"id"`

	// OrganizationID identifies the organization the enrichment belongs to.
	OrganizationID string `json:"organization_id,omitempty"`

	// ResourceID identifies the campaign or list the enrichment targets.
	ResourceID string `json:"resource_id"`

	// ResourceType is whether the enrichment targets a campaign or a list.
	ResourceType *ResourceType `json:"resource_type,omitempty"`

	// Type is the kind of enrichment.
	Type EnrichmentType `json:"type,omitempty"`

	// Limit is the maximum number of leads to enrich.
	Limit *float64 `json:"limit,omitempty"`

	// AutoUpdate reports whether the enrichment updates automatically.
	AutoUpdate *bool `json:"auto_update,omitempty"`

	// InProgress reports whether the enrichment is currently running.
	InProgress *bool `json:"in_progress,omitempty"`

	// SkipRowsWithoutEmail reports whether rows without an email are skipped.
	SkipRowsWithoutEmail *bool `json:"skip_rows_without_email,omitempty"`

	// EnrichmentPayload carries the raw enrichment payload, which the API models
	// as a free-form object, so it is preserved verbatim.
	EnrichmentPayload json.RawMessage `json:"enrichment_payload,omitempty"`
}

// ResourceEnrichment is the enrichment status for a resource.
//
// It reports whether an enrichment exists for the resource and how it is
// progressing, alongside the raw search filters that define it.
type ResourceEnrichment struct {
	// ResourceID identifies the campaign or list the enrichment targets.
	ResourceID string `json:"resource_id"`

	// EnrichmentPayload carries the raw enrichment payload, preserved verbatim.
	EnrichmentPayload json.RawMessage `json:"enrichment_payload,omitempty"`

	// Exists reports whether an enrichment exists for the resource.
	Exists *bool `json:"exists,omitempty"`

	// HasNoLeads reports whether the resource has no leads to enrich.
	HasNoLeads *bool `json:"has_no_leads,omitempty"`

	// InProgress reports whether an enrichment is currently running.
	InProgress *bool `json:"in_progress,omitempty"`

	// IsEvergreen reports whether the enrichment runs continuously.
	IsEvergreen *bool `json:"is_evergreen,omitempty"`

	// AutoUpdate reports whether the enrichment updates automatically.
	AutoUpdate *bool `json:"auto_update,omitempty"`

	// SearchFilters carries the raw SuperSearch query, preserved verbatim.
	SearchFilters json.RawMessage `json:"search_filters,omitempty"`
}

// CreateRequest is the body of a create-enrichment request.
type CreateRequest struct {
	// ResourceID identifies the campaign or list to enrich. Required.
	ResourceID string `json:"resource_id"`

	// Type is the kind of enrichment to perform.
	Type EnrichmentType `json:"type,omitempty"`

	// Limit is the maximum number of leads to enrich.
	Limit *float64 `json:"limit,omitempty"`

	// CustomFlow is the custom flow to apply to the enrichment.
	CustomFlow []string `json:"custom_flow,omitempty"`

	// Filters carries the raw enrichment filters, sent verbatim.
	Filters json.RawMessage `json:"filters,omitempty"`

	// IntegrationActions carries the raw integration actions, sent verbatim.
	IntegrationActions json.RawMessage `json:"integration_actions,omitempty"`
}

// SettingsRequest is the body of an update-enrichment-settings request. No field
// is required; an omitted field leaves the current value unchanged.
type SettingsRequest struct {
	// AutoUpdate sets whether the enrichment updates automatically.
	AutoUpdate *bool `json:"auto_update,omitempty"`

	// IsEvergreen sets whether the enrichment runs continuously.
	IsEvergreen *bool `json:"is_evergreen,omitempty"`

	// SkipRowsWithoutEmail sets whether rows without an email are skipped.
	SkipRowsWithoutEmail *bool `json:"skip_rows_without_email,omitempty"`
}

// RunRequest is the body of a run-enrichment request.
type RunRequest struct {
	// ResourceID identifies the campaign or list to enrich. Required.
	ResourceID string `json:"resource_id"`

	// ColumnName is the column to run the enrichment for.
	ColumnName string `json:"column_name,omitempty"`

	// Count is how many rows to enrich.
	Count *int64 `json:"count,omitempty"`

	// Limit is the maximum number of rows to enrich.
	Limit *int64 `json:"limit,omitempty"`

	// StartingRow is the row to start enriching from.
	StartingRow *int64 `json:"starting_row,omitempty"`

	// Overwrite overwrites existing values when true.
	Overwrite *bool `json:"overwrite,omitempty"`

	// LeadIDs restricts the run to specific leads.
	LeadIDs []string `json:"lead_ids,omitempty"`

	// Filters carries the raw enrichment filters, sent verbatim.
	Filters json.RawMessage `json:"filters,omitempty"`
}

// Create creates a standard enrichment and returns it.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Enrichment, error) {
	return instantly.PostResult[Enrichment](ctx, s.client, basePath+"/", req)
}

// Get returns the enrichment status for a resource.
func (s *Service) Get(ctx context.Context, resourceID string) (*ResourceEnrichment, error) {
	return instantly.GetResult[ResourceEnrichment](ctx, s.client, instantly.JoinPath(basePath, resourceID))
}

// UpdateSettings patches an enrichment's settings and returns its updated state.
func (s *Service) UpdateSettings(
	ctx context.Context, resourceID string, req SettingsRequest,
) (*Enrichment, error) {
	path := instantly.JoinPath(basePath, resourceID, "settings")

	return instantly.PatchResult[Enrichment](ctx, s.client, path, req)
}

// Run runs an enrichment for a resource and returns it.
func (s *Service) Run(ctx context.Context, req RunRequest) (*Enrichment, error) {
	return instantly.PostResult[Enrichment](ctx, s.client, basePath+"/run", req)
}

// History returns the enrichment history for a resource.
//
// Each entry is preserved verbatim as json.RawMessage because the API does not
// document the history entries as a fixed schema.
func (s *Service) History(ctx context.Context, resourceID string) ([]json.RawMessage, error) {
	var out []json.RawMessage
	if err := s.client.Get(ctx, instantly.JoinPath(basePath, "history", resourceID), &out); err != nil {
		return nil, err
	}

	return out, nil
}
