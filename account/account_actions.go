package account

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/mrz1836/go-instantly"
)

// PauseBulkRequest is the body of a bulk-pause request.
type PauseBulkRequest struct {
	// Emails are the accounts to pause. Required.
	Emails []string `json:"emails"`
}

// PauseBulkResponse reports which accounts a bulk pause paused and which failed.
type PauseBulkResponse struct {
	// PausedEmails are the accounts that were paused.
	PausedEmails []string `json:"paused_emails"`

	// FailedEmails are the accounts that could not be paused.
	FailedEmails []string `json:"failed_emails"`
}

// MoveRequest is the body of a move-accounts request.
type MoveRequest struct {
	// Emails are the accounts to move. Required.
	Emails []string `json:"emails"`

	// SourceWorkspaceID is the workspace the accounts are moved from. Required.
	SourceWorkspaceID string `json:"source_workspace_id"`

	// DestinationWorkspaceID is the workspace the accounts are moved to. Required.
	DestinationWorkspaceID string `json:"destination_workspace_id"`
}

// MoveResponse is the outcome of a move-accounts request.
type MoveResponse struct {
	// Status is the status of the move.
	Status string `json:"status"`
}

// WarmupToggleRequest is the body of an enable- or disable-warmup request.
//
// Target accounts either by listing them in Emails, or by setting
// IncludeAllEmails with an optional Search and ExcludedEmails to warm up
// everything matching except the exclusions.
type WarmupToggleRequest struct {
	// Emails are the specific accounts to target.
	Emails []string `json:"emails,omitempty"`

	// ExcludedEmails are accounts to skip when IncludeAllEmails is set.
	ExcludedEmails []string `json:"excluded_emails,omitempty"`

	// Filter carries an optional account filter, which the API does not document
	// as a fixed schema, so it is sent verbatim.
	Filter json.RawMessage `json:"filter,omitempty"`

	// IncludeAllEmails targets every account when true.
	IncludeAllEmails *bool `json:"include_all_emails,omitempty"`

	// Search narrows the targeted accounts when IncludeAllEmails is set.
	Search string `json:"search,omitempty"`
}

// BackgroundJob is the asynchronous job a warmup toggle enqueues.
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

	// Data carries the job payload, which the API does not document as a fixed
	// schema, so it is preserved verbatim.
	Data json.RawMessage `json:"data,omitempty"`
}

// WarmupAnalyticsRequest is the body of a warmup-analytics request.
type WarmupAnalyticsRequest struct {
	// Emails are the accounts to report warmup analytics for. Required.
	Emails []string `json:"emails"`
}

// WarmupAnalyticsResponse is the warmup analytics for the requested accounts.
//
// Both fields carry the raw API payload, which is keyed by account and date and
// is not documented as a fixed schema, so they are preserved verbatim.
type WarmupAnalyticsResponse struct {
	// AggregateData is the aggregate warmup analytics across the accounts.
	AggregateData json.RawMessage `json:"aggregate_data,omitempty"`

	// EmailDateData is the per-account, per-date warmup analytics.
	EmailDateData json.RawMessage `json:"email_date_data,omitempty"`
}

// DailyAnalytics is a single day of sending analytics for one account.
type DailyAnalytics struct {
	// Date is the day the analytics are for.
	Date string `json:"date"`

	// EmailAccount is the account the analytics are for.
	EmailAccount string `json:"email_account"`

	// Sent is the number of emails sent.
	Sent int64 `json:"sent"`

	// Opened is the number of opens.
	Opened int64 `json:"opened"`

	// UniqueOpened is the number of unique opens.
	UniqueOpened int64 `json:"unique_opened"`

	// Replies is the number of replies.
	Replies int64 `json:"replies"`

	// UniqueReplies is the number of unique replies.
	UniqueReplies int64 `json:"unique_replies"`

	// RepliesAutomatic is the number of automatic replies.
	RepliesAutomatic int64 `json:"replies_automatic"`

	// UniqueRepliesAutomatic is the number of unique automatic replies.
	UniqueRepliesAutomatic int64 `json:"unique_replies_automatic"`

	// Clicks is the number of clicks.
	Clicks int64 `json:"clicks"`

	// UniqueClicks is the number of unique clicks.
	UniqueClicks int64 `json:"unique_clicks"`

	// Contacted is the number of leads contacted.
	Contacted int64 `json:"contacted"`

	// NewLeadsContacted is the number of new leads contacted.
	NewLeadsContacted int64 `json:"new_leads_contacted"`

	// Bounced is the number of bounces.
	Bounced int64 `json:"bounced"`
}

// CtdStatus is the status of a custom tracking domain.
type CtdStatus struct {
	// Host is the tracking-domain host that was checked.
	Host string `json:"host"`

	// Success reports whether the tracking domain is configured correctly.
	Success bool `json:"success"`

	// CNAME reports whether the CNAME record is configured.
	CNAME bool `json:"cname"`

	// SSL reports whether SSL is configured.
	SSL bool `json:"ssl"`
}

// VitalsRequest is the body of a test-account-vitals request.
type VitalsRequest struct {
	// Accounts are the account email addresses to test.
	Accounts []string `json:"accounts,omitempty"`
}

// VitalsCheck is the DNS-record result for a single account's domain.
type VitalsCheck struct {
	// Domain is the domain the checks are for.
	Domain string `json:"domain"`

	// AllPass reports whether every check passed.
	AllPass bool `json:"allPass"`

	// DKIM reports whether the DKIM record passed.
	DKIM bool `json:"dkim"`

	// DMARC reports whether the DMARC record passed.
	DMARC bool `json:"dmarc"`

	// MX reports whether the MX record passed.
	MX bool `json:"mx"`

	// SPF reports whether the SPF record passed.
	SPF bool `json:"spf"`
}

// VitalsResponse is the outcome of a test-account-vitals request.
type VitalsResponse struct {
	// Status is the overall status of the vitals test.
	Status string `json:"status"`

	// SuccessList are the accounts whose vitals passed.
	SuccessList []VitalsCheck `json:"success_list"`

	// FailureList are the accounts whose vitals failed.
	FailureList []VitalsCheck `json:"failure_list"`
}

// Pause pauses a single account and returns its updated state.
func (s *Service) Pause(ctx context.Context, email string) (*Account, error) {
	return s.action(ctx, email, "pause")
}

// Resume resumes a single paused account and returns its updated state.
func (s *Service) Resume(ctx context.Context, email string) (*Account, error) {
	return s.action(ctx, email, "resume")
}

// MarkFixed marks a single account as fixed and returns its updated state.
func (s *Service) MarkFixed(ctx context.Context, email string) (*Account, error) {
	return s.action(ctx, email, "mark-fixed")
}

// PauseBulk pauses several accounts at once, reporting which paused and which
// failed.
func (s *Service) PauseBulk(ctx context.Context, req PauseBulkRequest) (*PauseBulkResponse, error) {
	out := &PauseBulkResponse{}
	if err := s.client.Post(ctx, basePath+"/pause", req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// Move moves accounts between workspaces.
func (s *Service) Move(ctx context.Context, req MoveRequest) (*MoveResponse, error) {
	out := &MoveResponse{}
	if err := s.client.Post(ctx, basePath+"/move", req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// EnableWarmup enables warmup for the targeted accounts and returns the
// background job that carries it out.
func (s *Service) EnableWarmup(ctx context.Context, req WarmupToggleRequest) (*BackgroundJob, error) {
	return s.warmupToggle(ctx, "enable", req)
}

// DisableWarmup disables warmup for the targeted accounts and returns the
// background job that carries it out.
func (s *Service) DisableWarmup(ctx context.Context, req WarmupToggleRequest) (*BackgroundJob, error) {
	return s.warmupToggle(ctx, "disable", req)
}

// WarmupAnalytics returns warmup analytics for the requested accounts.
func (s *Service) WarmupAnalytics(
	ctx context.Context, req WarmupAnalyticsRequest,
) (*WarmupAnalyticsResponse, error) {
	out := &WarmupAnalyticsResponse{}
	if err := s.client.Post(ctx, basePath+"/warmup-analytics", req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// DailyAnalytics returns daily sending analytics filtered by the supplied
// options.
func (s *Service) DailyAnalytics(ctx context.Context, opts ...AnalyticsOption) ([]DailyAnalytics, error) {
	q := instantly.NewQuery()
	for _, opt := range opts {
		if opt != nil {
			opt(q)
		}
	}

	var out []DailyAnalytics
	if err := s.client.Get(ctx, q.Path(basePath+"/analytics/daily"), &out); err != nil {
		return nil, err
	}

	return out, nil
}

// CtdStatus returns the status of a custom tracking domain by host.
func (s *Service) CtdStatus(ctx context.Context, host string) (*CtdStatus, error) {
	q := instantly.NewQuery().SetString("host", host)

	out := &CtdStatus{}
	if err := s.client.Get(ctx, q.Path(basePath+"/ctd/status"), out); err != nil {
		return nil, err
	}

	return out, nil
}

// TestVitals tests the DNS vitals of the given accounts.
func (s *Service) TestVitals(ctx context.Context, req VitalsRequest) (*VitalsResponse, error) {
	out := &VitalsResponse{}
	if err := s.client.Post(ctx, basePath+"/test/vitals", req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// action performs a single-account POST action that returns the account.
func (s *Service) action(ctx context.Context, email, verb string) (*Account, error) {
	out := &Account{}
	if err := s.client.Post(ctx, basePath+"/"+url.PathEscape(email)+"/"+verb, nil, out); err != nil {
		return nil, err
	}

	return out, nil
}

// warmupToggle performs an enable- or disable-warmup POST that returns a job.
func (s *Service) warmupToggle(
	ctx context.Context, verb string, req WarmupToggleRequest,
) (*BackgroundJob, error) {
	out := &BackgroundJob{}
	if err := s.client.Post(ctx, basePath+"/warmup/"+verb, req, out); err != nil {
		return nil, err
	}

	return out, nil
}
