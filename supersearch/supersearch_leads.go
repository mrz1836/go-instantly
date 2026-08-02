package supersearch

import (
	"context"
	"encoding/json"

	"github.com/mrz1836/go-instantly"
)

// Lead is a single lead previewed from a SuperSearch query.
type Lead struct {
	// FirstName is the lead's first name.
	FirstName string `json:"firstName"`

	// LastName is the lead's last name.
	LastName string `json:"lastName"`

	// FullName is the lead's full name.
	FullName string `json:"fullName"`

	// JobTitle is the lead's job title.
	JobTitle string `json:"jobTitle"`

	// CompanyID identifies the lead's company.
	CompanyID string `json:"companyId"`

	// CompanyName is the lead's company name.
	CompanyName string `json:"companyName"`

	// CompanyLogo is the URL of the company's logo.
	CompanyLogo string `json:"companyLogo"`

	// LinkedIn is the lead's LinkedIn URL.
	LinkedIn string `json:"linkedIn"`

	// Location is the lead's location.
	Location string `json:"location"`

	// IsOwned reports whether the lead is already owned in the workspace.
	IsOwned bool `json:"isOwned"`
}

// Keyword is a single faceted signal keyword and its count.
type Keyword struct {
	// Keyword is the signal keyword.
	Keyword string `json:"keyword"`

	// Count is how many results carry the keyword.
	Count float64 `json:"count"`
}

// LeadCount is the number of leads a SuperSearch query matches.
type LeadCount struct {
	// NumberOfLeads is how many leads the query matches.
	NumberOfLeads float64 `json:"number_of_leads"`
}

// Preview is a sample of the leads a SuperSearch query matches.
type Preview struct {
	// Leads is the sample of matching leads.
	Leads []Lead `json:"leads,omitempty"`

	// NumberOfLeads is how many leads the query matches in total.
	NumberOfLeads float64 `json:"number_of_leads"`

	// NumberOfRedactedResults is how many results are redacted from the preview.
	NumberOfRedactedResults float64 `json:"number_of_redacted_results"`
}

// EnrichLeadsResponse is the outcome of enriching leads from a SuperSearch query.
type EnrichLeadsResponse struct {
	// ID is the unique identifier of the enrichment.
	ID string `json:"id"`

	// OrganizationID identifies the organization the enrichment belongs to.
	OrganizationID string `json:"organization_id"`

	// ResourceID identifies the list the leads are enriched into.
	ResourceID string `json:"resource_id"`

	// ResourceType is whether the enrichment targets a campaign or a list.
	ResourceType *ResourceType `json:"resource_type,omitempty"`

	// Limit is the maximum number of leads enriched.
	Limit *float64 `json:"limit,omitempty"`

	// ListName is the name of the list the leads are enriched into.
	ListName string `json:"list_name,omitempty"`

	// CustomFlow is the custom flow applied to the enrichment.
	CustomFlow []string `json:"custom_flow,omitempty"`

	// BackgroundJobID identifies the background job carrying out the enrichment.
	BackgroundJobID *string `json:"background_job_id,omitempty"`

	// LiveListWorkflowID identifies the live-list workflow, when one is created.
	LiveListWorkflowID *string `json:"live_list_workflow_id,omitempty"`

	// SearchFilters carries the raw SuperSearch query, preserved verbatim.
	SearchFilters json.RawMessage `json:"search_filters,omitempty"`
}

// SearchRequest is the body of a count- or preview-leads request.
type SearchRequest struct {
	// SearchFilters carries the raw SuperSearch query. Required.
	SearchFilters json.RawMessage `json:"search_filters"`

	// ShowOneLeadPerCompany limits results to one lead per company when true.
	ShowOneLeadPerCompany *bool `json:"show_one_lead_per_company,omitempty"`

	// SkipOwnedLeads excludes leads already owned in the workspace when true.
	SkipOwnedLeads *bool `json:"skip_owned_leads,omitempty"`
}

// EnrichLeadsRequest is the body of an enrich-leads-from-supersearch request.
type EnrichLeadsRequest struct {
	// SearchFilters carries the raw SuperSearch query. Required.
	SearchFilters json.RawMessage `json:"search_filters"`

	// Limit is the maximum number of leads to enrich. Required.
	Limit float64 `json:"limit"`

	// ResourceID identifies the list to enrich the leads into.
	ResourceID string `json:"resource_id,omitempty"`

	// ListName names the list to enrich the leads into.
	ListName string `json:"list_name,omitempty"`

	// SearchName names the saved search.
	SearchName string `json:"search_name,omitempty"`

	// CustomFlow is the custom flow to apply to the enrichment.
	CustomFlow []string `json:"custom_flow,omitempty"`

	// WorkEmailEnrichment enriches work email addresses when true.
	WorkEmailEnrichment *bool `json:"work_email_enrichment,omitempty"`

	// FullyEnrichedProfile enriches a fully enriched profile when true.
	FullyEnrichedProfile *bool `json:"fully_enriched_profile,omitempty"`

	// AutoUpdate sets whether the enrichment updates automatically.
	AutoUpdate *bool `json:"auto_update,omitempty"`

	// SkipRowsWithoutEmail skips rows without an email when true.
	SkipRowsWithoutEmail *bool `json:"skip_rows_without_email,omitempty"`

	// AIEnrichment carries the raw AI enrichment settings, sent verbatim.
	AIEnrichment json.RawMessage `json:"ai_enrichment,omitempty"`

	// SignalEnrichment carries the raw signal enrichment settings, sent verbatim.
	SignalEnrichment json.RawMessage `json:"signal_enrichment,omitempty"`
}

// FacetRequest is the body of a signal-keywords-facet request.
type FacetRequest struct {
	// Category is the signal category to facet. Required.
	Category string `json:"category"`

	// Field is the field to facet keywords for. Required.
	Field string `json:"field"`

	// Prefix restricts the keywords to those starting with a prefix.
	Prefix string `json:"prefix,omitempty"`

	// Limit is the maximum number of keywords to return.
	Limit *float64 `json:"limit,omitempty"`
}

// signalKeywordsResponse is the {"keywords":[...]} wrapper the facet endpoint
// returns, unwrapped to a plain slice for the caller.
type signalKeywordsResponse struct {
	Keywords []Keyword `json:"keywords"`
}

// CountLeads returns how many leads a SuperSearch query matches.
func (s *Service) CountLeads(ctx context.Context, req SearchRequest) (*LeadCount, error) {
	return instantly.PostResult[LeadCount](ctx, s.client, basePath+"/count-leads-from-supersearch", req)
}

// PreviewLeads returns a sample of the leads a SuperSearch query matches.
func (s *Service) PreviewLeads(ctx context.Context, req SearchRequest) (*Preview, error) {
	return instantly.PostResult[Preview](ctx, s.client, basePath+"/preview-leads-from-supersearch", req)
}

// EnrichLeads enriches leads from a SuperSearch query into a list.
func (s *Service) EnrichLeads(ctx context.Context, req EnrichLeadsRequest) (*EnrichLeadsResponse, error) {
	return instantly.PostResult[EnrichLeadsResponse](ctx, s.client, basePath+"/enrich-leads-from-supersearch", req)
}

// SignalKeywords facets the signal keywords for a category, returning each
// keyword and its count.
func (s *Service) SignalKeywords(ctx context.Context, req FacetRequest) ([]Keyword, error) {
	out, err := instantly.PostResult[signalKeywordsResponse](ctx, s.client, basePath+"/signal-keywords-facet", req)
	if err != nil {
		return nil, err
	}

	return out.Keywords, nil
}
