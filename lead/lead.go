// Package lead provides typed access to the Instantly.ai V2 Lead API.
//
// It wraps the /api/v2/leads endpoints: creating, reading, patching, and
// deleting leads; listing them (via POST /leads/list, whose filters and cursor
// travel in the request body); bulk add, delete, assign, move, and merge; and
// updating interest status and subsequence membership.
//
//	svc := lead.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, lead.ListRequest{Campaign: "campaign-id", Limit: 50})
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package lead

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Lead API.
const basePath = "/api/v2/leads"

// Service provides access to the Instantly.ai V2 Lead API.
type Service struct {
	client *instantly.Client
}

// New builds a Lead API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Status is the sending status of a lead within a campaign.
type Status int64

// The statuses a lead can be in.
const (
	// StatusActive means the lead is active.
	StatusActive Status = 1

	// StatusPaused means the lead is paused.
	StatusPaused Status = 2

	// StatusCompleted means the lead has completed the campaign.
	StatusCompleted Status = 3

	// StatusBounced means the lead bounced.
	StatusBounced Status = -1

	// StatusUnsubscribed means the lead unsubscribed.
	StatusUnsubscribed Status = -2

	// StatusSkipped means the lead was skipped.
	StatusSkipped Status = -3
)

// InterestStatus is the interest status recorded for a lead.
type InterestStatus int64

// The interest statuses a lead can have.
const (
	// InterestOutOfOffice means the lead replied out of office.
	InterestOutOfOffice InterestStatus = 0

	// InterestInterested means the lead is interested.
	InterestInterested InterestStatus = 1

	// InterestMeetingBooked means a meeting was booked with the lead.
	InterestMeetingBooked InterestStatus = 2

	// InterestMeetingCompleted means a meeting with the lead was completed.
	InterestMeetingCompleted InterestStatus = 3

	// InterestWon means the lead was won.
	InterestWon InterestStatus = 4

	// InterestNotInterested means the lead is not interested.
	InterestNotInterested InterestStatus = -1

	// InterestWrongPerson means the lead was the wrong person.
	InterestWrongPerson InterestStatus = -2

	// InterestLost means the lead was lost.
	InterestLost InterestStatus = -3

	// InterestNoShow means the lead was a no-show.
	InterestNoShow InterestStatus = -4
)

// Lead is a single lead returned by the Instantly.ai V2 API.
//
// Fields the API declares as nullable are pointers so an absent value stays
// distinguishable from a zero value. Free-form payloads are preserved as raw
// JSON so nothing is lost.
type Lead struct {
	// ID is the unique identifier of the lead.
	ID string `json:"id"`

	// Organization is the organization the lead belongs to.
	Organization string `json:"organization"`

	// CompanyDomain is the company domain of the lead.
	CompanyDomain string `json:"company_domain"`

	// Status is the sending status of the lead.
	Status Status `json:"status"`

	// LtInterestStatus is the interest status of the lead.
	LtInterestStatus InterestStatus `json:"lt_interest_status"`

	// VerificationStatus is the verification status of the lead's email.
	VerificationStatus int64 `json:"verification_status"`

	// EnrichmentStatus is the enrichment status of the lead.
	EnrichmentStatus int64 `json:"enrichment_status"`

	// ESGCode is the email service group code of the lead.
	ESGCode int64 `json:"esg_code"`

	// ESPCode is the email service provider code of the lead.
	ESPCode int64 `json:"esp_code"`

	// EmailOpenCount is the number of times the lead opened an email.
	EmailOpenCount float64 `json:"email_open_count"`

	// EmailReplyCount is the number of times the lead replied.
	EmailReplyCount float64 `json:"email_reply_count"`

	// EmailClickCount is the number of times the lead clicked a link.
	EmailClickCount float64 `json:"email_click_count"`

	// UploadMethod is how the lead was uploaded.
	UploadMethod string `json:"upload_method,omitempty"`

	// TimestampCreated is when the lead was created.
	TimestampCreated string `json:"timestamp_created"`

	// TimestampUpdated is when the lead was last updated.
	TimestampUpdated string `json:"timestamp_updated"`

	// Email is the email address of the lead.
	Email *string `json:"email,omitempty"`

	// FirstName is the first name of the lead.
	FirstName *string `json:"first_name,omitempty"`

	// LastName is the last name of the lead.
	LastName *string `json:"last_name,omitempty"`

	// CompanyName is the company name of the lead.
	CompanyName *string `json:"company_name,omitempty"`

	// JobTitle is the job title of the lead.
	JobTitle *string `json:"job_title,omitempty"`

	// Phone is the phone number of the lead.
	Phone *string `json:"phone,omitempty"`

	// Website is the website of the lead.
	Website *string `json:"website,omitempty"`

	// Personalization is the personalization snippet for the lead.
	Personalization *string `json:"personalization,omitempty"`

	// AssignedTo is the user the lead is assigned to.
	AssignedTo *string `json:"assigned_to,omitempty"`

	// Campaign is the campaign the lead belongs to.
	Campaign *string `json:"campaign,omitempty"`

	// ListID is the lead list the lead belongs to.
	ListID *string `json:"list_id,omitempty"`

	// SubsequenceID is the subsequence the lead belongs to.
	SubsequenceID *string `json:"subsequence_id,omitempty"`

	// PLValueLead is the positive-lead value of the lead.
	PLValueLead *string `json:"pl_value_lead,omitempty"`

	// IsWebsiteVisitor reports whether the lead is a website visitor.
	IsWebsiteVisitor *bool `json:"is_website_visitor,omitempty"`

	// UploadedByUser is the user that uploaded the lead.
	UploadedByUser *string `json:"uploaded_by_user,omitempty"`

	// LastContactedFrom is the account the lead was last contacted from.
	LastContactedFrom *string `json:"last_contacted_from,omitempty"`

	// LastStepFrom is the account the last step was sent from.
	LastStepFrom *string `json:"last_step_from,omitempty"`

	// LastStepID identifies the last step executed for the lead.
	LastStepID *string `json:"last_step_id,omitempty"`

	// LastStepTimestampExecuted is when the last step was executed.
	LastStepTimestampExecuted *string `json:"last_step_timestamp_executed,omitempty"`

	// EmailClickedStep is the step whose email the lead clicked.
	EmailClickedStep *float64 `json:"email_clicked_step,omitempty"`

	// EmailClickedVariant is the variant whose email the lead clicked.
	EmailClickedVariant *float64 `json:"email_clicked_variant,omitempty"`

	// EmailOpenedStep is the step whose email the lead opened.
	EmailOpenedStep *float64 `json:"email_opened_step,omitempty"`

	// EmailOpenedVariant is the variant whose email the lead opened.
	EmailOpenedVariant *float64 `json:"email_opened_variant,omitempty"`

	// EmailRepliedStep is the step whose email the lead replied to.
	EmailRepliedStep *float64 `json:"email_replied_step,omitempty"`

	// EmailRepliedVariant is the variant whose email the lead replied to.
	EmailRepliedVariant *float64 `json:"email_replied_variant,omitempty"`

	// TimestampAddedSubsequence is when the lead was added to a subsequence.
	TimestampAddedSubsequence *string `json:"timestamp_added_subsequence,omitempty"`

	// TimestampLastContact is when the lead was last contacted.
	TimestampLastContact *string `json:"timestamp_last_contact,omitempty"`

	// TimestampLastOpen is when the lead last opened an email.
	TimestampLastOpen *string `json:"timestamp_last_open,omitempty"`

	// TimestampLastReply is when the lead last replied.
	TimestampLastReply *string `json:"timestamp_last_reply,omitempty"`

	// TimestampLastClick is when the lead last clicked a link.
	TimestampLastClick *string `json:"timestamp_last_click,omitempty"`

	// TimestampLastTouch is when the lead was last touched.
	TimestampLastTouch *string `json:"timestamp_last_touch,omitempty"`

	// TimestampLastInterestChange is when the lead's interest status last changed.
	TimestampLastInterestChange *string `json:"timestamp_last_interest_change,omitempty"`

	// Payload carries the raw lead payload, preserved verbatim.
	Payload json.RawMessage `json:"payload,omitempty"`

	// StatusSummary carries the raw status summary, preserved verbatim.
	StatusSummary json.RawMessage `json:"status_summary,omitempty"`

	// StatusSummarySubseq carries the raw subsequence status summary, preserved verbatim.
	StatusSummarySubseq json.RawMessage `json:"status_summary_subseq,omitempty"`
}

// ListRequest is the body of a list-leads request.
//
// Listing leads is a POST whose filters and pagination cursor travel in the
// request body rather than the query string. Every field is optional.
type ListRequest struct {
	// Campaign restricts results to a campaign.
	Campaign string `json:"campaign,omitempty"`

	// ListID restricts results to a lead list.
	ListID string `json:"list_id,omitempty"`

	// Search restricts results to leads matching a search term.
	Search string `json:"search,omitempty"`

	// Filter restricts results to a documented filter.
	Filter string `json:"filter,omitempty"`

	// IDs restricts results to specific lead identifiers.
	IDs []string `json:"ids,omitempty"`

	// ExcludedIDs excludes specific lead identifiers from the results.
	ExcludedIDs []string `json:"excluded_ids,omitempty"`

	// OrganizationUserIDs restricts results to leads assigned to these users.
	OrganizationUserIDs []string `json:"organization_user_ids,omitempty"`

	// SmartViewID restricts results to a smart view.
	SmartViewID string `json:"smart_view_id,omitempty"`

	// ESGCode restricts results to an email service group code.
	ESGCode string `json:"esg_code,omitempty"`

	// EnrichmentStatus restricts results to an enrichment status.
	EnrichmentStatus *float64 `json:"enrichment_status,omitempty"`

	// InCampaign restricts results to leads that are (or are not) in a campaign.
	InCampaign *bool `json:"in_campaign,omitempty"`

	// InList restricts results to leads that are (or are not) in a list.
	InList *bool `json:"in_list,omitempty"`

	// IsWebsiteVisitor restricts results to website visitors.
	IsWebsiteVisitor *bool `json:"is_website_visitor,omitempty"`

	// DistinctContacts returns only distinct contacts when set.
	DistinctContacts *bool `json:"distinct_contacts,omitempty"`

	// Contacts carries a contacts filter, sent verbatim.
	Contacts json.RawMessage `json:"contacts,omitempty"`

	// Queries carries a queries filter, sent verbatim.
	Queries json.RawMessage `json:"queries,omitempty"`

	// Limit is the maximum number of leads returned in a single page.
	Limit int `json:"limit,omitempty"`

	// StartingAfter is the pagination cursor to resume from.
	StartingAfter string `json:"starting_after,omitempty"`
}

// ListResponse is a single page of leads.
type ListResponse struct {
	// Items are the leads on this page.
	Items []Lead `json:"items"`

	// NextStartingAfter is the cursor for the following page, and is empty on
	// the last page.
	NextStartingAfter string `json:"next_starting_after,omitempty"`
}

// CreateRequest is the body of a create-lead request. No field is required.
type CreateRequest struct {
	// Email is the email address of the lead.
	Email *string `json:"email,omitempty"`

	// FirstName is the first name of the lead.
	FirstName *string `json:"first_name,omitempty"`

	// LastName is the last name of the lead.
	LastName *string `json:"last_name,omitempty"`

	// CompanyName is the company name of the lead.
	CompanyName *string `json:"company_name,omitempty"`

	// JobTitle is the job title of the lead.
	JobTitle *string `json:"job_title,omitempty"`

	// Phone is the phone number of the lead.
	Phone *string `json:"phone,omitempty"`

	// Website is the website of the lead.
	Website *string `json:"website,omitempty"`

	// Personalization is the personalization snippet for the lead.
	Personalization *string `json:"personalization,omitempty"`

	// AssignedTo is the user to assign the lead to.
	AssignedTo *string `json:"assigned_to,omitempty"`

	// Campaign is the campaign to add the lead to.
	Campaign *string `json:"campaign,omitempty"`

	// ListID is the lead list to add the lead to.
	ListID *string `json:"list_id,omitempty"`

	// BlocklistID is the blocklist to check the lead against.
	BlocklistID string `json:"blocklist_id,omitempty"`

	// PLValueLead is the positive-lead value of the lead.
	PLValueLead *string `json:"pl_value_lead,omitempty"`

	// LtInterestStatus is the interest status to set on the lead.
	LtInterestStatus *InterestStatus `json:"lt_interest_status,omitempty"`

	// CustomVariables carries custom variables for the lead, sent verbatim.
	CustomVariables json.RawMessage `json:"custom_variables,omitempty"`

	// SkipIfInCampaign skips the lead if it is already in a campaign.
	SkipIfInCampaign *bool `json:"skip_if_in_campaign,omitempty"`

	// SkipIfInList skips the lead if it is already in a list.
	SkipIfInList *bool `json:"skip_if_in_list,omitempty"`

	// SkipIfInWorkspace skips the lead if it is already in the workspace.
	SkipIfInWorkspace *bool `json:"skip_if_in_workspace,omitempty"`

	// VerifyLeadsOnImport verifies the lead's email on import when set.
	VerifyLeadsOnImport *bool `json:"verify_leads_on_import,omitempty"`

	// VerifyLeadsForLeadFinder verifies the lead for lead finder when set.
	VerifyLeadsForLeadFinder *bool `json:"verify_leads_for_lead_finder,omitempty"`
}

// UpdateRequest is the body of a patch-lead request. No field is required; an
// omitted field leaves the current value unchanged.
type UpdateRequest struct {
	// FirstName is the first name of the lead.
	FirstName *string `json:"first_name,omitempty"`

	// LastName is the last name of the lead.
	LastName *string `json:"last_name,omitempty"`

	// CompanyName is the company name of the lead.
	CompanyName *string `json:"company_name,omitempty"`

	// JobTitle is the job title of the lead.
	JobTitle *string `json:"job_title,omitempty"`

	// Phone is the phone number of the lead.
	Phone *string `json:"phone,omitempty"`

	// Website is the website of the lead.
	Website *string `json:"website,omitempty"`

	// Personalization is the personalization snippet for the lead.
	Personalization *string `json:"personalization,omitempty"`

	// AssignedTo is the user to assign the lead to.
	AssignedTo *string `json:"assigned_to,omitempty"`

	// PLValueLead is the positive-lead value of the lead.
	PLValueLead *string `json:"pl_value_lead,omitempty"`

	// LtInterestStatus is the interest status to set on the lead.
	LtInterestStatus *InterestStatus `json:"lt_interest_status,omitempty"`

	// CustomVariables carries custom variables for the lead, sent verbatim.
	CustomVariables json.RawMessage `json:"custom_variables,omitempty"`
}

// Create adds a new lead and returns it.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Lead, error) {
	out := &Lead{}
	if err := s.client.Post(ctx, basePath, req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// List returns a single page of leads matching the request.
//
// Listing is a POST: the filters and the pagination cursor travel in the request
// body. Pass the returned NextStartingAfter back as req.StartingAfter to fetch
// the following page, or use ListIter to walk them.
func (s *Service) List(ctx context.Context, req ListRequest) (*ListResponse, error) {
	out := &ListResponse{}
	if err := s.client.Post(ctx, basePath+"/list", req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// Get returns a single lead by its unique identifier.
func (s *Service) Get(ctx context.Context, id string) (*Lead, error) {
	out := &Lead{}
	if err := s.client.Get(ctx, basePath+"/"+url.PathEscape(id), out); err != nil {
		return nil, err
	}

	return out, nil
}

// Update patches a lead and returns its updated state.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Lead, error) {
	out := &Lead{}
	if err := s.client.Patch(ctx, basePath+"/"+url.PathEscape(id), req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// Delete deletes a lead and returns the lead that was deleted.
func (s *Service) Delete(ctx context.Context, id string) (*Lead, error) {
	out := &Lead{}
	if err := s.client.Delete(ctx, basePath+"/"+url.PathEscape(id), out); err != nil {
		return nil, err
	}

	return out, nil
}
