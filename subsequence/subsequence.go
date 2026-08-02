// Package subsequence provides typed access to the Instantly.ai V2 Campaign
// Subsequence API.
//
// It wraps the /api/v2/subsequences endpoints: creating, listing, reading,
// patching, and deleting campaign subsequences; duplicating them; pausing and
// resuming; and reading a subsequence's sending status.
//
//	svc := subsequence.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, subsequence.WithParentCampaign("campaign-id"))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package subsequence

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Campaign Subsequence API.
const basePath = "/api/v2/subsequences"

// Service provides access to the Instantly.ai V2 Campaign Subsequence API.
type Service struct {
	client *instantly.Client
}

// New builds a Campaign Subsequence API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Status is the current status of a subsequence.
type Status int64

// The statuses a subsequence can be in.
const (
	// StatusDraft means the subsequence is a draft and not yet active.
	StatusDraft Status = 0

	// StatusActive means the subsequence is currently running.
	StatusActive Status = 1

	// StatusPaused means the subsequence has been manually paused.
	StatusPaused Status = 2

	// StatusCompleted means the subsequence has finished running.
	StatusCompleted Status = 3

	// StatusRunningSubsequences means the subsequence has active child sequences.
	StatusRunningSubsequences Status = 4

	// StatusAccountsUnhealthy means the subsequence is paused for unhealthy accounts.
	StatusAccountsUnhealthy Status = -1

	// StatusBounceProtect means the subsequence is paused for high bounce rates.
	StatusBounceProtect Status = -2

	// StatusAccountSuspended means the subsequence's account is suspended.
	StatusAccountSuspended Status = -99
)

// DailyLimitMode is how a subsequence applies its daily sending limit.
type DailyLimitMode string

// The daily-limit modes a subsequence can use.
const (
	// DailyLimitInherit inherits the parent campaign's daily limit.
	DailyLimitInherit DailyLimitMode = "inherit"

	// DailyLimitCustom uses a custom daily limit.
	DailyLimitCustom DailyLimitMode = "custom"

	// DailyLimitUnlimited applies no daily limit.
	DailyLimitUnlimited DailyLimitMode = "unlimited"
)

// Subsequence is a single campaign subsequence returned by the Instantly.ai V2
// API.
//
// The deeply nested, free-form payloads (conditions, schedule, and sequences)
// are preserved as raw JSON so nothing is lost.
type Subsequence struct {
	// ID is the unique identifier of the subsequence.
	ID string `json:"id"`

	// Name is the name of the subsequence.
	Name string `json:"name"`

	// ParentCampaign is the campaign the subsequence belongs to.
	ParentCampaign string `json:"parent_campaign"`

	// Workspace is the workspace the subsequence belongs to.
	Workspace string `json:"workspace"`

	// Status is the current status of the subsequence.
	Status Status `json:"status"`

	// TimestampCreated is when the subsequence was created.
	TimestampCreated string `json:"timestamp_created"`

	// TimestampLeadsUpdated is when the subsequence's leads were last updated.
	TimestampLeadsUpdated string `json:"timestamp_leads_updated"`

	// DailyLimit is the daily sending limit, when a custom one is set.
	DailyLimit *float64 `json:"daily_limit,omitempty"`

	// DailyLimitMode is how the daily limit is applied.
	DailyLimitMode DailyLimitMode `json:"daily_limit_mode,omitempty"`

	// IgnoreAccountDailyLimit reports whether the account daily limit is ignored.
	IgnoreAccountDailyLimit bool `json:"ignore_account_daily_limit,omitempty"`

	// Conditions carries the trigger conditions, preserved verbatim.
	Conditions json.RawMessage `json:"conditions,omitempty"`

	// SubsequenceSchedule carries the sending schedule, preserved verbatim.
	SubsequenceSchedule json.RawMessage `json:"subsequence_schedule,omitempty"`

	// Sequences carries the email copy of the subsequence, preserved verbatim.
	Sequences json.RawMessage `json:"sequences,omitempty"`
}

// ListResponse is a single page of subsequences.
type ListResponse struct {
	// Items are the subsequences on this page.
	Items []Subsequence `json:"items"`

	// NextStartingAfter is the cursor for the following page, and is empty on
	// the last page.
	NextStartingAfter string `json:"next_starting_after,omitempty"`
}

// CreateRequest is the body of a create-subsequence request.
type CreateRequest struct {
	// ParentCampaign is the campaign to create the subsequence in. Required.
	ParentCampaign string `json:"parent_campaign"`

	// Name is the name of the subsequence. Required.
	Name string `json:"name"`

	// Conditions carries the trigger conditions, sent verbatim. Required.
	Conditions json.RawMessage `json:"conditions"`

	// SubsequenceSchedule carries the sending schedule, sent verbatim. Required.
	SubsequenceSchedule json.RawMessage `json:"subsequence_schedule"`

	// Sequences carries the email copy, sent verbatim. Required.
	Sequences json.RawMessage `json:"sequences"`

	// DailyLimit is the daily sending limit.
	DailyLimit *float64 `json:"daily_limit,omitempty"`

	// DailyLimitMode is how the daily limit is applied.
	DailyLimitMode DailyLimitMode `json:"daily_limit_mode,omitempty"`

	// IgnoreAccountDailyLimit ignores the account daily limit when set.
	IgnoreAccountDailyLimit *bool `json:"ignore_account_daily_limit,omitempty"`
}

// UpdateRequest is the body of a patch-subsequence request. No field is
// required; an omitted field leaves the current value unchanged.
type UpdateRequest struct {
	// Name is the name of the subsequence.
	Name string `json:"name,omitempty"`

	// DailyLimit is the daily sending limit.
	DailyLimit *float64 `json:"daily_limit,omitempty"`

	// DailyLimitMode is how the daily limit is applied.
	DailyLimitMode DailyLimitMode `json:"daily_limit_mode,omitempty"`

	// IgnoreAccountDailyLimit ignores the account daily limit when set.
	IgnoreAccountDailyLimit *bool `json:"ignore_account_daily_limit,omitempty"`
}

// DuplicateRequest is the body of a duplicate-subsequence request.
type DuplicateRequest struct {
	// ParentCampaign is the campaign to duplicate the subsequence into. Required.
	ParentCampaign string `json:"parent_campaign"`

	// Name is the name of the duplicated subsequence. Required.
	Name string `json:"name"`
}

// SendingStatus is the sending status of a subsequence.
//
// Both fields carry the raw API payload, which the API does not document as a
// fixed schema, so they are preserved verbatim.
type SendingStatus struct {
	// Summary is a summary of the subsequence's sending status.
	Summary json.RawMessage `json:"summary,omitempty"`

	// Diagnostics are the detailed sending diagnostics.
	Diagnostics json.RawMessage `json:"diagnostics,omitempty"`
}

// Create adds a new subsequence and returns it.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Subsequence, error) {
	out := &Subsequence{}
	if err := s.client.Post(ctx, basePath, req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// List returns a single page of subsequences filtered by the supplied options.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	q := instantly.NewQuery()
	for _, opt := range opts {
		if opt != nil {
			opt(q)
		}
	}

	out := &ListResponse{}
	if err := s.client.Get(ctx, q.Path(basePath), out); err != nil {
		return nil, err
	}

	return out, nil
}

// Get returns a single subsequence by its unique identifier.
func (s *Service) Get(ctx context.Context, id string) (*Subsequence, error) {
	out := &Subsequence{}
	if err := s.client.Get(ctx, basePath+"/"+url.PathEscape(id), out); err != nil {
		return nil, err
	}

	return out, nil
}

// Update patches a subsequence and returns its updated state.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Subsequence, error) {
	out := &Subsequence{}
	if err := s.client.Patch(ctx, basePath+"/"+url.PathEscape(id), req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// Delete deletes a subsequence and returns the subsequence that was deleted.
func (s *Service) Delete(ctx context.Context, id string) (*Subsequence, error) {
	out := &Subsequence{}
	if err := s.client.Delete(ctx, basePath+"/"+url.PathEscape(id), out); err != nil {
		return nil, err
	}

	return out, nil
}

// Duplicate copies a subsequence into a campaign and returns the new subsequence.
func (s *Service) Duplicate(ctx context.Context, id string, req DuplicateRequest) (*Subsequence, error) {
	out := &Subsequence{}
	if err := s.client.Post(ctx, basePath+"/"+url.PathEscape(id)+"/duplicate", req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// Pause pauses a subsequence and returns its updated state.
func (s *Service) Pause(ctx context.Context, id string) (*Subsequence, error) {
	return s.action(ctx, id, "pause")
}

// Resume resumes a paused subsequence and returns its updated state.
func (s *Service) Resume(ctx context.Context, id string) (*Subsequence, error) {
	return s.action(ctx, id, "resume")
}

// SendingStatus returns the sending status of a subsequence.
func (s *Service) SendingStatus(ctx context.Context, id string) (*SendingStatus, error) {
	out := &SendingStatus{}
	if err := s.client.Get(ctx, basePath+"/"+url.PathEscape(id)+"/sending-status", out); err != nil {
		return nil, err
	}

	return out, nil
}

// action performs a POST lifecycle action that returns the subsequence.
func (s *Service) action(ctx context.Context, id, verb string) (*Subsequence, error) {
	out := &Subsequence{}
	if err := s.client.Post(ctx, basePath+"/"+url.PathEscape(id)+"/"+verb, nil, out); err != nil {
		return nil, err
	}

	return out, nil
}
