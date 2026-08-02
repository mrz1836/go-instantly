package campaign

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Campaign API.
const basePath = "/api/v2/campaigns"

// Service provides access to the Instantly.ai V2 Campaign API.
type Service struct {
	client *instantly.Client
}

// New builds a Campaign API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Status is the current status of a campaign.
type Status int64

// The statuses a campaign can be in.
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

// ScheduleItem is one named sending window in a campaign schedule.
type ScheduleItem struct {
	// Name is the label of the sending window.
	Name string `json:"name,omitempty"`

	// Timezone is the timezone the window is expressed in.
	Timezone string `json:"timezone,omitempty"`

	// Timing carries the from/to times of the window, preserved verbatim.
	Timing json.RawMessage `json:"timing,omitempty"`

	// Days carries which days the window is active on, preserved verbatim.
	Days json.RawMessage `json:"days,omitempty"`
}

// Schedule is the sending schedule of a campaign.
type Schedule struct {
	// StartDate is the date the campaign starts sending.
	StartDate *string `json:"start_date,omitempty"`

	// EndDate is the date the campaign stops sending.
	EndDate *string `json:"end_date,omitempty"`

	// Schedules are the named sending windows.
	Schedules []ScheduleItem `json:"schedules,omitempty"`
}

// Campaign is a single campaign returned by the Instantly.ai V2 API.
//
// Fields the API declares as nullable are pointers so an absent value stays
// distinguishable from a zero value. Deeply nested, free-form payloads (the
// sequences, provider routing rules, and variable maps) are preserved as raw
// JSON so nothing is lost.
type Campaign struct {
	// ID is the unique identifier of the campaign.
	ID string `json:"id"`

	// Name is the name of the campaign.
	Name string `json:"name"`

	// Status is the current status of the campaign.
	Status Status `json:"status"`

	// CampaignSchedule is the sending schedule of the campaign.
	CampaignSchedule Schedule `json:"campaign_schedule"`

	// TimestampCreated is when the campaign was created.
	TimestampCreated string `json:"timestamp_created"`

	// TimestampUpdated is when the campaign was last updated.
	TimestampUpdated string `json:"timestamp_updated"`

	// OpenTracking reports whether opens are tracked.
	OpenTracking bool `json:"open_tracking"`

	// Organization is the organization ID that owns the campaign.
	Organization *string `json:"organization,omitempty"`

	// OwnedBy is the owner ID of the campaign.
	OwnedBy *string `json:"owned_by,omitempty"`

	// AISDRID is the AI sales agent that created the campaign.
	AISDRID *string `json:"ai_sdr_id,omitempty"`

	// AllowRiskyContacts reports whether risky contacts are allowed.
	AllowRiskyContacts *bool `json:"allow_risky_contacts,omitempty"`

	// DailyLimit is the daily sending limit.
	DailyLimit *float64 `json:"daily_limit,omitempty"`

	// DailyMaxLeads is the daily maximum number of new leads to contact.
	DailyMaxLeads *int64 `json:"daily_max_leads,omitempty"`

	// DisableBounceProtect reports whether bounce protection is disabled.
	DisableBounceProtect *bool `json:"disable_bounce_protect,omitempty"`

	// EmailGap is the gap between emails, in minutes.
	EmailGap *float64 `json:"email_gap,omitempty"`

	// FirstEmailTextOnly reports whether the first email is sent as text only.
	FirstEmailTextOnly *bool `json:"first_email_text_only,omitempty"`

	// InsertUnsubscribeHeader reports whether an unsubscribe header is inserted.
	InsertUnsubscribeHeader *bool `json:"insert_unsubscribe_header,omitempty"`

	// IsEvergreen reports whether the campaign is evergreen.
	IsEvergreen *bool `json:"is_evergreen,omitempty"`

	// LinkTracking reports whether links are tracked.
	LinkTracking *bool `json:"link_tracking,omitempty"`

	// MatchLeadESP reports whether leads are matched by ESP.
	MatchLeadESP *bool `json:"match_lead_esp,omitempty"`

	// NotSendingStatus is the reason the campaign is not sending, when set.
	NotSendingStatus *float64 `json:"not_sending_status,omitempty"`

	// PLValue is the value of every positive lead.
	PLValue *float64 `json:"pl_value,omitempty"`

	// PrioritizeNewLeads reports whether new leads are prioritized.
	PrioritizeNewLeads *bool `json:"prioritize_new_leads,omitempty"`

	// RandomWaitMax is the maximum random wait between emails, in minutes.
	RandomWaitMax *float64 `json:"random_wait_max,omitempty"`

	// StopForCompany reports whether the campaign stops for the whole company.
	StopForCompany *bool `json:"stop_for_company,omitempty"`

	// StopOnAutoReply reports whether the campaign stops on an auto reply.
	StopOnAutoReply *bool `json:"stop_on_auto_reply,omitempty"`

	// StopOnReply reports whether the campaign stops on a reply.
	StopOnReply *bool `json:"stop_on_reply,omitempty"`

	// TextOnly reports whether the campaign is text only.
	TextOnly *bool `json:"text_only,omitempty"`

	// BCCList are the accounts to BCC on emails.
	BCCList []string `json:"bcc_list,omitempty"`

	// CCList are the accounts to CC on emails.
	CCList []string `json:"cc_list,omitempty"`

	// EmailList are the sending accounts the campaign uses.
	EmailList []string `json:"email_list,omitempty"`

	// EmailTagList are the tags whose accounts the campaign sends from.
	EmailTagList []string `json:"email_tag_list,omitempty"`

	// Sequences carries the email copy of the campaign, preserved verbatim.
	Sequences json.RawMessage `json:"sequences,omitempty"`

	// ProviderRoutingRules carries the provider routing rules, preserved verbatim.
	ProviderRoutingRules json.RawMessage `json:"provider_routing_rules,omitempty"`

	// AutoVariantSelect carries the auto-variant-select settings, preserved verbatim.
	AutoVariantSelect json.RawMessage `json:"auto_variant_select,omitempty"`

	// CoreVariables carries the campaign core variables, preserved verbatim.
	CoreVariables json.RawMessage `json:"core_variables,omitempty"`

	// CustomVariables carries the campaign custom variables, preserved verbatim.
	CustomVariables json.RawMessage `json:"custom_variables,omitempty"`

	// LimitEmailsPerCompanyOverride carries the per-company limit override,
	// preserved verbatim.
	LimitEmailsPerCompanyOverride json.RawMessage `json:"limit_emails_per_company_override,omitempty"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded campaign re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (c *Campaign) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, c.TimestampCreated)
}

// ListResponse is a single page of campaigns.
//
// It aliases instantly.Page[Campaign], the cursor-paginated envelope every
// resource shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[Campaign]

// CreateRequest is the body of a create-campaign request. Only Name and
// CampaignSchedule are required; every other field defaults when omitted.
type CreateRequest struct {
	// Name is the name of the campaign. Required.
	Name string `json:"name"`

	// CampaignSchedule is the sending schedule. Required.
	CampaignSchedule Schedule `json:"campaign_schedule"`

	// AISDRID is the AI sales agent to attribute the campaign to.
	AISDRID *string `json:"ai_sdr_id,omitempty"`

	// AllowRiskyContacts allows risky contacts when set.
	AllowRiskyContacts *bool `json:"allow_risky_contacts,omitempty"`

	// AutoVariantSelect carries the auto-variant-select settings, sent verbatim.
	AutoVariantSelect json.RawMessage `json:"auto_variant_select,omitempty"`

	// BCCList are the accounts to BCC on emails.
	BCCList []string `json:"bcc_list,omitempty"`

	// CCList are the accounts to CC on emails.
	CCList []string `json:"cc_list,omitempty"`

	// DailyLimit is the daily sending limit.
	DailyLimit *float64 `json:"daily_limit,omitempty"`

	// DailyMaxLeads is the daily maximum number of new leads to contact.
	DailyMaxLeads *int64 `json:"daily_max_leads,omitempty"`

	// DisableBounceProtect disables bounce protection when set.
	DisableBounceProtect *bool `json:"disable_bounce_protect,omitempty"`

	// EmailGap is the gap between emails, in minutes.
	EmailGap *float64 `json:"email_gap,omitempty"`

	// EmailList are the sending accounts the campaign uses.
	EmailList []string `json:"email_list,omitempty"`

	// EmailTagList are the tags whose accounts the campaign sends from.
	EmailTagList []string `json:"email_tag_list,omitempty"`

	// FirstEmailTextOnly sends the first email as text only when set.
	FirstEmailTextOnly *bool `json:"first_email_text_only,omitempty"`

	// InsertUnsubscribeHeader inserts an unsubscribe header when set.
	InsertUnsubscribeHeader *bool `json:"insert_unsubscribe_header,omitempty"`

	// IsEvergreen marks the campaign evergreen when set.
	IsEvergreen *bool `json:"is_evergreen,omitempty"`

	// LimitEmailsPerCompanyOverride overrides the per-company limit, sent verbatim.
	LimitEmailsPerCompanyOverride json.RawMessage `json:"limit_emails_per_company_override,omitempty"`

	// LinkTracking tracks links when set.
	LinkTracking *bool `json:"link_tracking,omitempty"`

	// MatchLeadESP matches leads by ESP when set.
	MatchLeadESP *bool `json:"match_lead_esp,omitempty"`

	// OpenTracking tracks opens when set.
	OpenTracking *bool `json:"open_tracking,omitempty"`

	// OwnedBy is the owner ID to assign the campaign to.
	OwnedBy *string `json:"owned_by,omitempty"`

	// PLValue is the value of every positive lead.
	PLValue *float64 `json:"pl_value,omitempty"`

	// PrioritizeNewLeads prioritizes new leads when set.
	PrioritizeNewLeads *bool `json:"prioritize_new_leads,omitempty"`

	// ProviderRoutingRules carries the provider routing rules, sent verbatim.
	ProviderRoutingRules json.RawMessage `json:"provider_routing_rules,omitempty"`

	// RandomWaitMax is the maximum random wait between emails, in minutes.
	RandomWaitMax *float64 `json:"random_wait_max,omitempty"`

	// Sequences carries the email copy of the campaign, sent verbatim.
	Sequences json.RawMessage `json:"sequences,omitempty"`

	// StopForCompany stops the campaign for the whole company when set.
	StopForCompany *bool `json:"stop_for_company,omitempty"`

	// StopOnAutoReply stops the campaign on an auto reply when set.
	StopOnAutoReply *bool `json:"stop_on_auto_reply,omitempty"`

	// StopOnReply stops the campaign on a reply when set.
	StopOnReply *bool `json:"stop_on_reply,omitempty"`

	// TextOnly makes the campaign text only when set.
	TextOnly *bool `json:"text_only,omitempty"`
}

// UpdateRequest is the body of a patch-campaign request. No field is required;
// an omitted field leaves the current value unchanged.
type UpdateRequest struct {
	// Name is the name of the campaign.
	Name string `json:"name,omitempty"`

	// CampaignSchedule is the sending schedule.
	CampaignSchedule *Schedule `json:"campaign_schedule,omitempty"`

	// AllowRiskyContacts allows risky contacts when set.
	AllowRiskyContacts *bool `json:"allow_risky_contacts,omitempty"`

	// AutoVariantSelect carries the auto-variant-select settings, sent verbatim.
	AutoVariantSelect json.RawMessage `json:"auto_variant_select,omitempty"`

	// BCCList are the accounts to BCC on emails.
	BCCList []string `json:"bcc_list,omitempty"`

	// CCList are the accounts to CC on emails.
	CCList []string `json:"cc_list,omitempty"`

	// DailyLimit is the daily sending limit.
	DailyLimit *float64 `json:"daily_limit,omitempty"`

	// DailyMaxLeads is the daily maximum number of new leads to contact.
	DailyMaxLeads *int64 `json:"daily_max_leads,omitempty"`

	// DisableBounceProtect disables bounce protection when set.
	DisableBounceProtect *bool `json:"disable_bounce_protect,omitempty"`

	// EmailGap is the gap between emails, in minutes.
	EmailGap *float64 `json:"email_gap,omitempty"`

	// EmailList are the sending accounts the campaign uses.
	EmailList []string `json:"email_list,omitempty"`

	// EmailTagList are the tags whose accounts the campaign sends from.
	EmailTagList []string `json:"email_tag_list,omitempty"`

	// FirstEmailTextOnly sends the first email as text only when set.
	FirstEmailTextOnly *bool `json:"first_email_text_only,omitempty"`

	// InsertUnsubscribeHeader inserts an unsubscribe header when set.
	InsertUnsubscribeHeader *bool `json:"insert_unsubscribe_header,omitempty"`

	// IsEvergreen marks the campaign evergreen when set.
	IsEvergreen *bool `json:"is_evergreen,omitempty"`

	// LinkTracking tracks links when set.
	LinkTracking *bool `json:"link_tracking,omitempty"`

	// MatchLeadESP matches leads by ESP when set.
	MatchLeadESP *bool `json:"match_lead_esp,omitempty"`

	// OpenTracking tracks opens when set.
	OpenTracking *bool `json:"open_tracking,omitempty"`

	// PLValue is the value of every positive lead.
	PLValue *float64 `json:"pl_value,omitempty"`

	// PrioritizeNewLeads prioritizes new leads when set.
	PrioritizeNewLeads *bool `json:"prioritize_new_leads,omitempty"`

	// ProviderRoutingRules carries the provider routing rules, sent verbatim.
	ProviderRoutingRules json.RawMessage `json:"provider_routing_rules,omitempty"`

	// RandomWaitMax is the maximum random wait between emails, in minutes.
	RandomWaitMax *float64 `json:"random_wait_max,omitempty"`

	// Sequences carries the email copy of the campaign, sent verbatim.
	Sequences json.RawMessage `json:"sequences,omitempty"`

	// StopForCompany stops the campaign for the whole company when set.
	StopForCompany *bool `json:"stop_for_company,omitempty"`

	// StopOnAutoReply stops the campaign on an auto reply when set.
	StopOnAutoReply *bool `json:"stop_on_auto_reply,omitempty"`

	// StopOnReply stops the campaign on a reply when set.
	StopOnReply *bool `json:"stop_on_reply,omitempty"`

	// TextOnly makes the campaign text only when set.
	TextOnly *bool `json:"text_only,omitempty"`
}

// Create adds a new campaign and returns it.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Campaign, error) {
	return instantly.PostResult[Campaign](ctx, s.client, basePath, req)
}

// List returns a single page of campaigns filtered by the supplied options.
//
// Pagination is cursor based: pass the returned NextStartingAfter back with
// WithStartingAfter to fetch the following page.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}

// Get returns a single campaign by its unique identifier.
func (s *Service) Get(ctx context.Context, id string) (*Campaign, error) {
	return instantly.GetResult[Campaign](ctx, s.client, instantly.JoinPath(basePath, id))
}

// Update patches a campaign and returns its updated state.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Campaign, error) {
	return instantly.PatchResult[Campaign](ctx, s.client, instantly.JoinPath(basePath, id), req)
}

// Delete deletes a campaign and returns the campaign that was deleted.
func (s *Service) Delete(ctx context.Context, id string) (*Campaign, error) {
	return instantly.DeleteResult[Campaign](ctx, s.client, instantly.JoinPath(basePath, id))
}
