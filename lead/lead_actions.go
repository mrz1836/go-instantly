package lead

import (
	"context"
	"encoding/json"
	"net/http"
)

// BulkDeleteRequest is the body of a bulk-delete request.
type BulkDeleteRequest struct {
	// IDs are the leads to delete.
	IDs []string `json:"ids,omitempty"`

	// CampaignID restricts the delete to a campaign.
	CampaignID string `json:"campaign_id,omitempty"`

	// ListID restricts the delete to a lead list.
	ListID string `json:"list_id,omitempty"`

	// Status restricts the delete to leads with a status.
	Status *Status `json:"status,omitempty"`

	// Limit caps how many leads are deleted.
	Limit int `json:"limit,omitempty"`
}

// BulkAddRequest is the body of a bulk-add request.
type BulkAddRequest struct {
	// Leads are the leads to add, sent verbatim. It is a JSON array of lead objects.
	Leads json.RawMessage `json:"leads"`

	// CampaignID is the campaign to add the leads to.
	CampaignID string `json:"campaign_id,omitempty"`

	// ListID is the lead list to add the leads to.
	ListID string `json:"list_id,omitempty"`

	// AssignedTo is the user to assign the leads to.
	AssignedTo string `json:"assigned_to,omitempty"`

	// BlocklistID is the blocklist to check the leads against.
	BlocklistID *string `json:"blocklist_id,omitempty"`

	// SkipIfInCampaign skips leads already in a campaign.
	SkipIfInCampaign *bool `json:"skip_if_in_campaign,omitempty"`

	// SkipIfInList skips leads already in a list.
	SkipIfInList *bool `json:"skip_if_in_list,omitempty"`

	// SkipIfInWorkspace skips leads already in the workspace.
	SkipIfInWorkspace *bool `json:"skip_if_in_workspace,omitempty"`

	// VerifyLeadsOnImport verifies the leads' emails on import.
	VerifyLeadsOnImport *bool `json:"verify_leads_on_import,omitempty"`
}

// AddResponse is the outcome of a bulk-add request.
type AddResponse struct {
	// Status is the status of the import.
	Status string `json:"status"`

	// LeadsUploaded is the number of leads uploaded.
	LeadsUploaded int64 `json:"leads_uploaded"`

	// TotalSent is the number of leads sent for import.
	TotalSent int64 `json:"total_sent"`

	// SkippedCount is the number of leads skipped.
	SkippedCount int64 `json:"skipped_count"`

	// DuplicatedLeads is the number of duplicate leads.
	DuplicatedLeads int64 `json:"duplicated_leads"`

	// DuplicateEmailCount is the number of duplicate emails.
	DuplicateEmailCount int64 `json:"duplicate_email_count"`

	// InvalidEmailCount is the number of invalid emails.
	InvalidEmailCount int64 `json:"invalid_email_count"`

	// IncompleteCount is the number of incomplete leads.
	IncompleteCount int64 `json:"incomplete_count"`

	// InBlocklist is the number of leads found in the blocklist.
	InBlocklist int64 `json:"in_blocklist"`

	// BlocklistUsed is the blocklist that was checked, when one was.
	BlocklistUsed *string `json:"blocklist_used,omitempty"`

	// RemainingInPlan is how many leads remain in the workspace plan.
	RemainingInPlan *int64 `json:"remaining_in_plan,omitempty"`

	// CreatedLeads carries the created leads, preserved verbatim.
	CreatedLeads json.RawMessage `json:"created_leads,omitempty"`
}

// BulkAssignRequest is the body of a bulk-assign request.
type BulkAssignRequest struct {
	// OrganizationUserIDs are the users to assign leads to. Required.
	OrganizationUserIDs []string `json:"organization_user_ids"`

	// AssignedTo is the current assignee to reassign from.
	AssignedTo string `json:"assigned_to,omitempty"`

	// Campaign restricts the assignment to a campaign.
	Campaign string `json:"campaign,omitempty"`

	// ListID restricts the assignment to a lead list.
	ListID string `json:"list_id,omitempty"`

	// Filter restricts the assignment to a documented filter.
	Filter string `json:"filter,omitempty"`

	// Search restricts the assignment to leads matching a search term.
	Search string `json:"search,omitempty"`

	// SmartViewID restricts the assignment to a smart view.
	SmartViewID string `json:"smart_view_id,omitempty"`

	// IDs restricts the assignment to specific leads.
	IDs []string `json:"ids,omitempty"`

	// InCampaign restricts the assignment to leads that are in a campaign.
	InCampaign *bool `json:"in_campaign,omitempty"`

	// InList restricts the assignment to leads that are in a list.
	InList *bool `json:"in_list,omitempty"`

	// HasClause reports whether a query clause is present.
	HasClause *bool `json:"has_clause,omitempty"`

	// Queries carries a queries filter, sent verbatim.
	Queries json.RawMessage `json:"queries,omitempty"`

	// Limit caps how many leads are assigned.
	Limit int `json:"limit,omitempty"`
}

// MergeRequest is the body of a merge-leads request.
type MergeRequest struct {
	// LeadID is the lead to merge from. Required.
	LeadID string `json:"lead_id"`

	// DestinationLeadID is the lead to merge into. Required.
	DestinationLeadID string `json:"destination_lead_id"`
}

// UpdateInterestStatusRequest is the body of an update-interest-status request.
type UpdateInterestStatusRequest struct {
	// LeadEmail is the email of the lead to update. Required.
	LeadEmail string `json:"lead_email"`

	// InterestValue is the interest status to set. Required; a nil value clears
	// the interest status.
	InterestValue *InterestStatus `json:"interest_value"`

	// AIInterestValue is the AI-assigned interest value to set.
	AIInterestValue *float64 `json:"ai_interest_value,omitempty"`

	// CampaignID restricts the update to a campaign.
	CampaignID string `json:"campaign_id,omitempty"`

	// ListID restricts the update to a lead list.
	ListID string `json:"list_id,omitempty"`

	// DisableAutoInterest disables automatic interest detection when set.
	DisableAutoInterest *bool `json:"disable_auto_interest,omitempty"`
}

// SubsequenceRemoveRequest is the body of a remove-from-subsequence request.
type SubsequenceRemoveRequest struct {
	// ID is the lead to remove from its subsequence. Required.
	ID string `json:"id"`
}

// SubsequenceMoveRequest is the body of a move-to-subsequence request.
type SubsequenceMoveRequest struct {
	// ID is the lead to move. Required.
	ID string `json:"id"`

	// SubsequenceID is the subsequence to move the lead to. Required.
	SubsequenceID string `json:"subsequence_id"`
}

// MoveRequest is the body of a move-leads request.
type MoveRequest struct {
	// ToCampaignID is the campaign to move the leads to.
	ToCampaignID string `json:"to_campaign_id,omitempty"`

	// ToListID is the lead list to move the leads to.
	ToListID string `json:"to_list_id,omitempty"`

	// IDs restricts the move to specific leads.
	IDs []string `json:"ids,omitempty"`

	// ExcludedIDs excludes specific leads from the move.
	ExcludedIDs []string `json:"excluded_ids,omitempty"`

	// Campaign restricts the move to leads in a campaign.
	Campaign string `json:"campaign,omitempty"`

	// ListID restricts the move to leads in a list.
	ListID string `json:"list_id,omitempty"`

	// AssignedTo restricts the move to leads assigned to a user.
	AssignedTo string `json:"assigned_to,omitempty"`

	// Filter restricts the move to a documented filter.
	Filter string `json:"filter,omitempty"`

	// Search restricts the move to leads matching a search term.
	Search string `json:"search,omitempty"`

	// ESGCode restricts the move to an email service group code.
	ESGCode string `json:"esg_code,omitempty"`

	// ESPCode restricts the move to an email service provider code.
	ESPCode *float64 `json:"esp_code,omitempty"`

	// CopyLeads copies the leads instead of moving them when set.
	CopyLeads *bool `json:"copy_leads,omitempty"`

	// CheckDuplicates checks for duplicates when set.
	CheckDuplicates *bool `json:"check_duplicates,omitempty"`

	// CheckDuplicatesInCampaigns checks for duplicates across campaigns when set.
	CheckDuplicatesInCampaigns *bool `json:"check_duplicates_in_campaigns,omitempty"`

	// InCampaign restricts the move to leads that are in a campaign.
	InCampaign *bool `json:"in_campaign,omitempty"`

	// InList restricts the move to leads that are in a list.
	InList *bool `json:"in_list,omitempty"`

	// ResetInterestStatus resets the interest status on move when set.
	ResetInterestStatus *bool `json:"reset_interest_status,omitempty"`

	// SkipLeadsInVerification skips leads still being verified when set.
	SkipLeadsInVerification *bool `json:"skip_leads_in_verification,omitempty"`

	// IgnoreResourceFilterClauses ignores resource filter clauses when set.
	IgnoreResourceFilterClauses *bool `json:"ignore_resource_filter_clauses,omitempty"`

	// Contacts carries a contacts filter, sent verbatim.
	Contacts json.RawMessage `json:"contacts,omitempty"`

	// Queries carries a queries filter, sent verbatim.
	Queries json.RawMessage `json:"queries,omitempty"`

	// Limit caps how many leads are moved.
	Limit *float64 `json:"limit,omitempty"`
}

// BackgroundJob is the asynchronous job a long-running lead action enqueues.
type BackgroundJob struct {
	// ID is the unique identifier of the job.
	ID string `json:"id"`

	// Type is the job type.
	Type string `json:"type"`

	// Status is the current status of the job.
	Status string `json:"status"`

	// Progress is how far the job has progressed.
	Progress float64 `json:"progress"`

	// EntityType is the type of entity the job operates on.
	EntityType string `json:"entity_type"`

	// EntityID is the identifier of the entity the job operates on.
	EntityID *string `json:"entity_id,omitempty"`

	// WorkspaceID is the workspace the job belongs to.
	WorkspaceID string `json:"workspace_id"`

	// UserID is the user that started the job.
	UserID *string `json:"user_id,omitempty"`

	// CreatedAt is when the job was created.
	CreatedAt string `json:"created_at"`

	// UpdatedAt is when the job was last updated.
	UpdatedAt string `json:"updated_at"`

	// Data carries the job payload, preserved verbatim.
	Data json.RawMessage `json:"data,omitempty"`
}

// BulkDelete deletes several leads at once and returns how many were deleted.
//
// The API models this as a DELETE that carries a request body, so it goes
// through the low-level client directly.
func (s *Service) BulkDelete(ctx context.Context, req BulkDeleteRequest) (int64, error) {
	out := &struct {
		Count int64 `json:"count"`
	}{}
	if err := s.client.Do(ctx, http.MethodDelete, basePath, req, out); err != nil {
		return 0, err
	}

	return out.Count, nil
}

// BulkAdd adds several leads at once and returns the import summary.
func (s *Service) BulkAdd(ctx context.Context, req BulkAddRequest) (*AddResponse, error) {
	out := &AddResponse{}
	if err := s.client.Post(ctx, basePath+"/add", req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// BulkAssign assigns several leads to organization users. The endpoint returns
// no content, so a nil return is the only signal of success.
func (s *Service) BulkAssign(ctx context.Context, req BulkAssignRequest) error {
	return s.client.Post(ctx, basePath+"/bulk-assign", req, nil)
}

// Merge merges one lead into another and returns the merged lead.
func (s *Service) Merge(ctx context.Context, req MergeRequest) (*Lead, error) {
	out := &Lead{}
	if err := s.client.Post(ctx, basePath+"/merge", req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// UpdateInterestStatus updates the interest status of a lead. The endpoint
// returns no content, so a nil return is the only signal of success.
func (s *Service) UpdateInterestStatus(ctx context.Context, req UpdateInterestStatusRequest) error {
	return s.client.Post(ctx, basePath+"/update-interest-status", req, nil)
}

// RemoveFromSubsequence removes a lead from its subsequence and returns the lead.
func (s *Service) RemoveFromSubsequence(ctx context.Context, req SubsequenceRemoveRequest) (*Lead, error) {
	out := &Lead{}
	if err := s.client.Post(ctx, basePath+"/subsequence/remove", req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// MoveToSubsequence moves a lead to a subsequence and returns the lead.
func (s *Service) MoveToSubsequence(ctx context.Context, req SubsequenceMoveRequest) (*Lead, error) {
	out := &Lead{}
	if err := s.client.Post(ctx, basePath+"/subsequence/move", req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// Move moves leads to a campaign or list and returns the background job that
// carries out the move.
func (s *Service) Move(ctx context.Context, req MoveRequest) (*BackgroundJob, error) {
	out := &BackgroundJob{}
	if err := s.client.Post(ctx, basePath+"/move", req, out); err != nil {
		return nil, err
	}

	return out, nil
}
