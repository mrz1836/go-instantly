package auditlog

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Audit Log API.
const basePath = "/api/v2/audit-logs"

// Service provides access to the Instantly.ai V2 Audit Log API.
type Service struct {
	client *instantly.Client
}

// New builds an Audit Log API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// ActivityType is the kind of activity an audit log records.
//
// The API numbers the types non-contiguously (1-12 then 18-38), so the named
// constants below mirror the documented values rather than an unbroken range.
type ActivityType int64

// The documented activity types an audit log can record.
const (
	// ActivityTypeUserLogin records a "User login" event.
	ActivityTypeUserLogin ActivityType = 1

	// ActivityTypeLeadDeletion records a "Lead deletion" event.
	ActivityTypeLeadDeletion ActivityType = 2

	// ActivityTypeCampaignDeletion records a "Campaign deletion" event.
	ActivityTypeCampaignDeletion ActivityType = 3

	// ActivityTypeCampaignLaunch records a "Campaign launch" event.
	ActivityTypeCampaignLaunch ActivityType = 4

	// ActivityTypeCampaignPause records a "Campaign pause" event.
	ActivityTypeCampaignPause ActivityType = 5

	// ActivityTypeAccountAddition records a "Account addition" event.
	ActivityTypeAccountAddition ActivityType = 6

	// ActivityTypeAccountDeletion records a "Account deletion" event.
	ActivityTypeAccountDeletion ActivityType = 7

	// ActivityTypeLeadMoved records a "Lead moved" event.
	ActivityTypeLeadMoved ActivityType = 8

	// ActivityTypeLeadAdded records a "Lead added" event.
	ActivityTypeLeadAdded ActivityType = 9

	// ActivityTypeLeadMerged records a "Lead merged" event.
	ActivityTypeLeadMerged ActivityType = 10

	// ActivityTypeCampaignUpdate records a "Campaign update" event.
	ActivityTypeCampaignUpdate ActivityType = 11

	// ActivityTypeSubsequenceUpdate records a "Subsequence update" event.
	ActivityTypeSubsequenceUpdate ActivityType = 12

	// ActivityTypeWebhookCreated records a "Webhook created" event.
	ActivityTypeWebhookCreated ActivityType = 18

	// ActivityTypeWebhookUpdated records a "Webhook updated" event.
	ActivityTypeWebhookUpdated ActivityType = 19

	// ActivityTypeWebhookMarkedAsError records a "Webhook marked as error" event.
	ActivityTypeWebhookMarkedAsError ActivityType = 20

	// ActivityTypeWebhookResumed records a "Webhook resumed" event.
	ActivityTypeWebhookResumed ActivityType = 21

	// ActivityTypeTOTPEnrollmentStarted records a "TOTP enrollment started" event.
	ActivityTypeTOTPEnrollmentStarted ActivityType = 22

	// ActivityTypeTOTPEnabled records a "TOTP enabled" event.
	ActivityTypeTOTPEnabled ActivityType = 23

	// ActivityTypeTOTPReplacementStarted records a "TOTP replacement started" event.
	ActivityTypeTOTPReplacementStarted ActivityType = 24

	// ActivityTypeTOTPReplaced records a "TOTP replaced" event.
	ActivityTypeTOTPReplaced ActivityType = 25

	// ActivityTypeTOTPDisabled records a "TOTP disabled" event.
	ActivityTypeTOTPDisabled ActivityType = 26

	// ActivityTypeMFARecoveryCodesGenerated records a "MFA recovery codes generated" event.
	ActivityTypeMFARecoveryCodesGenerated ActivityType = 27

	// ActivityTypeMFARecoveryCodeUsed records a "MFA recovery code used" event.
	ActivityTypeMFARecoveryCodeUsed ActivityType = 28

	// ActivityTypeMFALoginChallengeFailed records a "MFA login challenge failed" event.
	ActivityTypeMFALoginChallengeFailed ActivityType = 29

	// ActivityTypeMFALoginChallengeFailedTooManyTimes records a "MFA login challenge failed too many times" event.
	ActivityTypeMFALoginChallengeFailedTooManyTimes ActivityType = 30

	// ActivityTypeMFALoginSucceeded records a "MFA login succeeded" event.
	ActivityTypeMFALoginSucceeded ActivityType = 31

	// ActivityTypeSubscriberImportStarted records a "Subscriber import started" event.
	ActivityTypeSubscriberImportStarted ActivityType = 32

	// ActivityTypeSubscriberImportCancelled records a "Subscriber import cancelled" event.
	ActivityTypeSubscriberImportCancelled ActivityType = 33

	// ActivityTypeLeadExported records a "Lead exported" event.
	ActivityTypeLeadExported ActivityType = 34

	// ActivityTypeSuperSearchEnrichmentCreated records a "SuperSearch enrichment created" event.
	ActivityTypeSuperSearchEnrichmentCreated ActivityType = 35

	// ActivityTypeAccountUpdate records a "Account update" event.
	ActivityTypeAccountUpdate ActivityType = 36

	// ActivityTypeAPIKeyCreated records a "API key created" event.
	ActivityTypeAPIKeyCreated ActivityType = 37

	// ActivityTypeAPIKeyDeleted records a "API key deleted" event.
	ActivityTypeAPIKeyDeleted ActivityType = 38
)

// Log is a single audit log record returned by the Instantly.ai V2 API.
//
// Fields the API declares as nullable are pointers, so an absent value stays
// distinguishable from a zero value: a nil AffectedCount means the API reported
// nothing, which is not the same as reporting zero.
type Log struct {
	// ID is the unique identifier of the audit log record.
	ID string `json:"id"`

	// Timestamp is when the activity occurred.
	Timestamp string `json:"timestamp"`

	// OrganizationID identifies the organization the activity belongs to.
	OrganizationID string `json:"organization_id"`

	// ActivityType is the kind of activity that was performed.
	ActivityType ActivityType `json:"activity_type"`

	// IPAddress is the address from which the activity was performed.
	IPAddress string `json:"ip_address"`

	// FromAPI reports whether the activity was performed via the API.
	FromAPI bool `json:"from_api"`

	// AffectedCount is the number of items affected by the activity.
	AffectedCount *float64 `json:"affected_count,omitempty"`

	// APIKeyID identifies the API key that performed the activity, when one did.
	APIKeyID *string `json:"api_key_id,omitempty"`

	// CampaignID identifies the campaign the activity relates to, if any.
	CampaignID *string `json:"campaign_id,omitempty"`

	// ListID identifies the lead list the activity relates to, if any.
	ListID *string `json:"list_id,omitempty"`

	// SubsequenceID identifies the subsequence the activity relates to, if any.
	SubsequenceID *string `json:"subsequence_id,omitempty"`

	// WebhookID identifies the webhook the activity relates to, if any.
	WebhookID *string `json:"webhook_id,omitempty"`

	// UserAgent is the user agent of the client that performed the activity.
	UserAgent *string `json:"user_agent,omitempty"`

	// UserID identifies the user who performed the activity, when known.
	UserID *string `json:"user_id,omitempty"`

	// UserName is the name of the user who performed the activity, when known.
	UserName *string `json:"user_name,omitempty"`

	// AuditMetadata carries the sanitized metadata about the activity, which the
	// API models as a free-form object, so it is preserved verbatim.
	AuditMetadata json.RawMessage `json:"audit_metadata,omitempty"`
}

// ParsedTimestamp parses Timestamp as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded record re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (l *Log) ParsedTimestamp() (time.Time, error) {
	return time.Parse(time.RFC3339, l.Timestamp)
}

// ListResponse is a single page of audit log records.
//
// It aliases instantly.Page[Log], the cursor-paginated envelope every resource
// shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[Log]

// List returns a single page of audit log records filtered by the supplied
// options.
//
// Pagination is cursor based: pass the returned NextStartingAfter back with
// WithStartingAfter to fetch the following page.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}
