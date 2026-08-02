package account

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Account API.
const basePath = "/api/v2/accounts"

// Service provides access to the Instantly.ai V2 Account API.
type Service struct {
	client *instantly.Client
}

// New builds an Account API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Status is the current status of a sending account.
type Status int64

// The statuses a sending account can be in.
const (
	// StatusActive means the account is active.
	StatusActive Status = 1

	// StatusPaused means the account is paused.
	StatusPaused Status = 2

	// StatusMaintenance means the account is temporarily paused for maintenance
	// and will be resumed automatically.
	StatusMaintenance Status = 3

	// StatusConnectionError means the account has a connection error.
	StatusConnectionError Status = -1

	// StatusSoftBounceError means the account has a soft bounce error.
	StatusSoftBounceError Status = -2

	// StatusSendingError means the account has a sending error.
	StatusSendingError Status = -3
)

// ProviderCode identifies the mail provider backing a sending account.
type ProviderCode int64

// The provider codes a sending account can use.
const (
	// ProviderCustomIMAP is a custom IMAP/SMTP provider.
	ProviderCustomIMAP ProviderCode = 1

	// ProviderGoogle is Google.
	ProviderGoogle ProviderCode = 2

	// ProviderMicrosoft is Microsoft.
	ProviderMicrosoft ProviderCode = 3

	// ProviderAWS is AWS.
	ProviderAWS ProviderCode = 4

	// ProviderAirMail is AirMail.
	ProviderAirMail ProviderCode = 8

	// ProviderAirmailInstant is AirMail Instant.
	ProviderAirmailInstant ProviderCode = 11
)

// WarmupIncrement is the warmup ramp-up increment setting.
type WarmupIncrement string

// The warmup increment settings.
const (
	// WarmupIncrementDisabled disables ramp up.
	WarmupIncrementDisabled WarmupIncrement = "disabled"

	// WarmupIncrement0 is the slowest ramp up.
	WarmupIncrement0 WarmupIncrement = "0"

	// WarmupIncrement1 is a ramp-up setting.
	WarmupIncrement1 WarmupIncrement = "1"

	// WarmupIncrement2 is a ramp-up setting.
	WarmupIncrement2 WarmupIncrement = "2"

	// WarmupIncrement3 is a ramp-up setting.
	WarmupIncrement3 WarmupIncrement = "3"

	// WarmupIncrement4 is the fastest ramp-up setting.
	WarmupIncrement4 WarmupIncrement = "4"
)

// Warmup is the warmup configuration of a sending account.
type Warmup struct {
	// Limit is the daily warmup sending limit.
	Limit *float64 `json:"limit,omitempty"`

	// ReplyRate is the warmup reply rate.
	ReplyRate *float64 `json:"reply_rate,omitempty"`

	// Increment is the ramp-up increment setting.
	Increment WarmupIncrement `json:"increment,omitempty"`

	// WarmupCustomFTag is the custom filter tag used for warmup.
	WarmupCustomFTag string `json:"warmup_custom_ftag,omitempty"`

	// Advanced carries the advanced warmup settings, which the API does not
	// document as a fixed schema, so they are preserved verbatim.
	Advanced json.RawMessage `json:"advanced,omitempty"`
}

// Tag is a custom tag assigned to an account, returned only when a list request
// asks for tags with WithIncludeTags.
type Tag struct {
	// ID is the unique identifier of the tag.
	ID string `json:"id"`

	// Label is the display label of the tag.
	Label string `json:"label"`

	// Description is the detailed description of the tag, when present.
	Description *string `json:"description,omitempty"`
}

// Account is a single sending account returned by the Instantly.ai V2 API.
//
// Fields the API declares as nullable are pointers so an absent value stays
// distinguishable from a zero value.
type Account struct {
	// Email is the email address of the account, and its unique identifier.
	Email string `json:"email"`

	// FirstName is the first name associated with the account.
	FirstName string `json:"first_name"`

	// LastName is the last name associated with the account.
	LastName string `json:"last_name"`

	// Organization is the organization ID that owns the account.
	Organization string `json:"organization"`

	// TimestampCreated is when the account was created.
	TimestampCreated string `json:"timestamp_created"`

	// TimestampUpdated is when the account was last updated.
	TimestampUpdated string `json:"timestamp_updated"`

	// Status is the current status of the account.
	Status Status `json:"status"`

	// WarmupStatus is the current warmup status of the account.
	WarmupStatus int64 `json:"warmup_status"`

	// ProviderCode is the provider backing the account.
	ProviderCode ProviderCode `json:"provider_code"`

	// SetupPending reports whether account setup is still pending.
	SetupPending bool `json:"setup_pending"`

	// IsManagedAccount reports whether this is a managed account.
	IsManagedAccount bool `json:"is_managed_account"`

	// SendingGap is the gap between emails sent from this account, in minutes.
	SendingGap float64 `json:"sending_gap,omitempty"`

	// Signature is the email signature for the account.
	Signature *string `json:"signature,omitempty"`

	// ReplyTo is the custom reply-to address for the account.
	ReplyTo *string `json:"reply_to,omitempty"`

	// AddedBy is the user ID that added the account.
	AddedBy *string `json:"added_by,omitempty"`

	// ModifiedBy is the user ID that last modified the account.
	ModifiedBy *string `json:"modified_by,omitempty"`

	// DailyLimit is the daily email sending limit.
	DailyLimit *float64 `json:"daily_limit,omitempty"`

	// DailyLimitMax is the maximum daily sending limit for AirMail accounts.
	DailyLimitMax *float64 `json:"daily_limit_max,omitempty"`

	// InboxPlacementTestLimit is the limit for inbox placement tests.
	InboxPlacementTestLimit *float64 `json:"inbox_placement_test_limit,omitempty"`

	// StatWarmupScore is the warmup score for the account.
	StatWarmupScore *float64 `json:"stat_warmup_score,omitempty"`

	// WarmupLimitMax is the maximum daily warmup limit for AirMail accounts.
	WarmupLimitMax *float64 `json:"warmup_limit_max,omitempty"`

	// EnableSlowRamp reports whether slow ramp up is enabled.
	EnableSlowRamp *bool `json:"enable_slow_ramp,omitempty"`

	// AutofixFailed reports whether automatic reconnection attempts have failed.
	// A nil value means reconnection is still in progress.
	AutofixFailed *bool `json:"autofix_failed,omitempty"`

	// TrackingDomainName is the custom tracking domain.
	TrackingDomainName *string `json:"tracking_domain_name,omitempty"`

	// TrackingDomainStatus is the status of the custom tracking domain.
	TrackingDomainStatus *string `json:"tracking_domain_status,omitempty"`

	// TimestampWarmupStart is when warmup was started.
	TimestampWarmupStart *string `json:"timestamp_warmup_start,omitempty"`

	// StatusMessage carries the raw status message, which the API does not
	// document as a fixed schema, so it is preserved verbatim.
	StatusMessage json.RawMessage `json:"status_message,omitempty"`

	// Warmup is the warmup configuration of the account.
	Warmup *Warmup `json:"warmup,omitempty"`

	// Tags are the custom tags assigned to the account, populated only when a
	// list request is made with WithIncludeTags.
	Tags []Tag `json:"tags,omitempty"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded account re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (a *Account) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, a.TimestampCreated)
}

// ListResponse is a single page of accounts.
//
// It aliases instantly.Page[Account], the cursor-paginated envelope every
// resource shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[Account]

// CreateRequest is the body of a create-account request.
//
// The account is connected over IMAP/SMTP, so the credentials and host/port for
// both are required, along with the identity fields.
type CreateRequest struct {
	// Email is the email address of the account. Required.
	Email string `json:"email"`

	// FirstName is the first name associated with the account. Required.
	FirstName string `json:"first_name"`

	// LastName is the last name associated with the account. Required.
	LastName string `json:"last_name"`

	// ProviderCode is the provider backing the account. Required.
	ProviderCode ProviderCode `json:"provider_code"`

	// IMAPUsername is the IMAP username. Required.
	IMAPUsername string `json:"imap_username"`

	// IMAPPassword is the IMAP password. Required.
	IMAPPassword string `json:"imap_password"`

	// IMAPHost is the IMAP host. Required.
	IMAPHost string `json:"imap_host"`

	// IMAPPort is the IMAP port. Required.
	IMAPPort int64 `json:"imap_port"`

	// SMTPUsername is the SMTP username. Required.
	SMTPUsername string `json:"smtp_username"`

	// SMTPPassword is the SMTP password. Required.
	SMTPPassword string `json:"smtp_password"`

	// SMTPHost is the SMTP host. Required.
	SMTPHost string `json:"smtp_host"`

	// SMTPPort is the SMTP port. Required.
	SMTPPort int64 `json:"smtp_port"`

	// DailyLimit is the daily email sending limit.
	DailyLimit *float64 `json:"daily_limit,omitempty"`

	// EnableSlowRamp reports whether slow ramp up is enabled.
	EnableSlowRamp *bool `json:"enable_slow_ramp,omitempty"`

	// InboxPlacementTestLimit is the limit for inbox placement tests.
	InboxPlacementTestLimit *float64 `json:"inbox_placement_test_limit,omitempty"`

	// ReplyTo is the custom reply-to address for the account.
	ReplyTo string `json:"reply_to,omitempty"`

	// SendingGap is the gap between emails, in minutes.
	SendingGap *float64 `json:"sending_gap,omitempty"`

	// Signature is the email signature for the account.
	Signature *string `json:"signature,omitempty"`

	// SkipCNAMECheck skips the tracking-domain CNAME check when true.
	SkipCNAMECheck *bool `json:"skip_cname_check,omitempty"`

	// TrackingDomainName is the custom tracking domain.
	TrackingDomainName *string `json:"tracking_domain_name,omitempty"`

	// TrackingDomainStatus is the status of the custom tracking domain.
	TrackingDomainStatus *string `json:"tracking_domain_status,omitempty"`

	// Warmup is the warmup configuration for the account.
	Warmup *Warmup `json:"warmup,omitempty"`

	// WarmupCustomFTag is the custom filter tag used for warmup.
	WarmupCustomFTag string `json:"warmup_custom_ftag,omitempty"`
}

// UpdateRequest is the body of a patch-account request. No field is required; an
// omitted field leaves the current value unchanged.
type UpdateRequest struct {
	// FirstName is the first name associated with the account.
	FirstName string `json:"first_name,omitempty"`

	// LastName is the last name associated with the account.
	LastName string `json:"last_name,omitempty"`

	// DailyLimit is the daily email sending limit.
	DailyLimit *float64 `json:"daily_limit,omitempty"`

	// EnableSlowRamp reports whether slow ramp up is enabled.
	EnableSlowRamp *bool `json:"enable_slow_ramp,omitempty"`

	// InboxPlacementTestLimit is the limit for inbox placement tests.
	InboxPlacementTestLimit *float64 `json:"inbox_placement_test_limit,omitempty"`

	// RemoveTrackingDomain removes the tracking domain when true.
	RemoveTrackingDomain *bool `json:"remove_tracking_domain,omitempty"`

	// ReplyTo is the custom reply-to address for the account.
	ReplyTo *string `json:"reply_to,omitempty"`

	// SendingGap is the gap between emails, in minutes.
	SendingGap *float64 `json:"sending_gap,omitempty"`

	// Signature is the email signature for the account.
	Signature *string `json:"signature,omitempty"`

	// SkipCNAMECheck skips the tracking-domain CNAME check when true.
	SkipCNAMECheck *bool `json:"skip_cname_check,omitempty"`

	// TrackingDomainName is the custom tracking domain.
	TrackingDomainName *string `json:"tracking_domain_name,omitempty"`

	// TrackingDomainStatus is the status of the custom tracking domain.
	TrackingDomainStatus *string `json:"tracking_domain_status,omitempty"`

	// Warmup is the warmup configuration for the account.
	Warmup *Warmup `json:"warmup,omitempty"`
}

// Create adds a new sending account and returns it.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Account, error) {
	return instantly.PostResult[Account](ctx, s.client, basePath, req)
}

// List returns a single page of accounts filtered by the supplied options.
//
// Pagination is cursor based: pass the returned NextStartingAfter back with
// WithStartingAfter to fetch the following page.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}

// Get returns a single account by its email address.
func (s *Service) Get(ctx context.Context, email string) (*Account, error) {
	return instantly.GetResult[Account](ctx, s.client, instantly.JoinPath(basePath, email))
}

// Update patches an account and returns its updated state.
func (s *Service) Update(ctx context.Context, email string, req UpdateRequest) (*Account, error) {
	return instantly.PatchResult[Account](ctx, s.client, instantly.JoinPath(basePath, email), req)
}

// Delete deletes an account and returns the account that was deleted.
func (s *Service) Delete(ctx context.Context, email string) (*Account, error) {
	return instantly.DeleteResult[Account](ctx, s.client, instantly.JoinPath(basePath, email))
}
