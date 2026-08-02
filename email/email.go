package email

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Email API.
const basePath = "/api/v2/emails"

// Service provides access to the Instantly.ai V2 Email API.
type Service struct {
	client *instantly.Client
}

// New builds an Email API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Body is the content of an email.
//
// The API requires HTML when sending an email; Text carries the plain-text
// alternative when one is present.
type Body struct {
	// HTML is the HTML content of the email.
	HTML string `json:"html"`

	// Text is the plain-text alternative to the HTML content.
	Text string `json:"text,omitempty"`
}

// Email is a single email returned by the Instantly.ai V2 API.
//
// Fields the API declares as nullable are pointers so an absent value stays
// distinguishable from a zero value: a nil IsUnread means the API reported
// nothing, which is not the same as reporting zero.
//
// The API declares several numeric flags as JSON numbers rather than integers
// or booleans, so they are modeled as *float64 to decode any valid value.
type Email struct {
	// ID is the unique identifier of the email.
	ID string `json:"id"`

	// TimestampCreated is when the email record was created.
	TimestampCreated string `json:"timestamp_created"`

	// TimestampEmail is when the email itself was sent or received.
	TimestampEmail string `json:"timestamp_email"`

	// MessageID is the message identifier assigned by the mail transport.
	MessageID string `json:"message_id"`

	// Subject is the subject line of the email.
	Subject string `json:"subject"`

	// ToAddressEmailList is the comma-separated list of recipient addresses.
	ToAddressEmailList string `json:"to_address_email_list"`

	// Body is the content of the email.
	Body Body `json:"body"`

	// OrganizationID identifies the organization the email belongs to.
	OrganizationID string `json:"organization_id"`

	// EAccount is the sending account the email is associated with.
	EAccount string `json:"eaccount"`

	// FromAddressEmail is the sender address.
	FromAddressEmail *string `json:"from_address_email,omitempty"`

	// CCAddressEmailList is the comma-separated list of carbon-copy addresses.
	CCAddressEmailList *string `json:"cc_address_email_list,omitempty"`

	// BCCAddressEmailList is the comma-separated list of blind carbon-copy addresses.
	BCCAddressEmailList *string `json:"bcc_address_email_list,omitempty"`

	// ReplyTo is the address replies are directed to.
	ReplyTo *string `json:"reply_to,omitempty"`

	// CampaignID identifies the campaign the email belongs to.
	CampaignID *string `json:"campaign_id,omitempty"`

	// SubsequenceID identifies the campaign subsequence the email belongs to.
	SubsequenceID *string `json:"subsequence_id,omitempty"`

	// ListID identifies the lead list the email belongs to.
	ListID *string `json:"list_id,omitempty"`

	// Lead is the email address of the lead the email relates to.
	Lead *string `json:"lead,omitempty"`

	// LeadID identifies the lead the email relates to.
	LeadID *string `json:"lead_id,omitempty"`

	// UEType is the unibox event type of the email.
	UEType *float64 `json:"ue_type,omitempty"`

	// Step is the campaign step the email was sent from.
	Step *string `json:"step,omitempty"`

	// IsUnread reports whether the email is still unread.
	IsUnread *float64 `json:"is_unread,omitempty"`

	// IsAutoReply reports whether the email was detected as an auto reply.
	IsAutoReply *float64 `json:"is_auto_reply,omitempty"`

	// ReminderTS is the timestamp of the reminder set on the email.
	ReminderTS *string `json:"reminder_ts,omitempty"`

	// AIInterestValue is the interest score assigned to the email.
	AIInterestValue *float64 `json:"ai_interest_value,omitempty"`

	// AIAssisted reports whether the email was drafted with AI assistance.
	AIAssisted *float64 `json:"ai_assisted,omitempty"`

	// IsFocused reports whether the email belongs to the focused inbox.
	IsFocused *float64 `json:"is_focused,omitempty"`

	// IStatus is the interest status of the email.
	IStatus *float64 `json:"i_status,omitempty"`

	// ThreadID identifies the thread the email belongs to.
	ThreadID *string `json:"thread_id,omitempty"`

	// ContentPreview is a short preview of the email content.
	ContentPreview *string `json:"content_preview,omitempty"`

	// AttachmentJSON carries the raw attachment payload, which the API does not
	// document as a fixed schema. It is preserved verbatim so no data is lost.
	AttachmentJSON json.RawMessage `json:"attachment_json,omitempty"`

	// FromAddressJSON carries the raw structured sender payload.
	FromAddressJSON json.RawMessage `json:"from_address_json,omitempty"`

	// ToAddressJSON carries the raw structured recipient payload.
	ToAddressJSON json.RawMessage `json:"to_address_json,omitempty"`

	// CCAddressJSON carries the raw structured carbon-copy payload.
	CCAddressJSON json.RawMessage `json:"cc_address_json,omitempty"`

	// AIAgentID identifies the AI agent associated with the email.
	AIAgentID *string `json:"ai_agent_id,omitempty"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded email re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (e *Email) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, e.TimestampCreated)
}

// ListResponse is a single page of emails.
//
// It aliases instantly.Page[Email], the cursor-paginated envelope every resource
// shares, so the generic pagination helpers accept List directly. The cursor is
// an opaque value that may be an identifier or a timestamp depending on the
// request, so it is never parsed.
type ListResponse = instantly.Page[Email]

// MarkThreadReadResponse is the outcome of marking a thread as read.
type MarkThreadReadResponse struct {
	// Success reports whether the thread was marked as read.
	Success bool `json:"success"`
}

// SendTestRequest is the body of a send-test-email request. Every field is
// required by the API.
type SendTestRequest struct {
	// EAccount is the sending account the test email is sent from.
	EAccount string `json:"eaccount"`

	// ToAddressEmailList is the comma-separated list of recipient addresses.
	ToAddressEmailList string `json:"to_address_email_list"`

	// Subject is the subject line of the test email.
	Subject string `json:"subject"`

	// Body is the content of the test email.
	Body Body `json:"body"`
}

// ReplyRequest is the body of a reply request.
type ReplyRequest struct {
	// ReplyToUUID identifies the email being replied to. Required.
	ReplyToUUID string `json:"reply_to_uuid"`

	// EAccount is the sending account the reply is sent from. Required.
	EAccount string `json:"eaccount"`

	// Subject is the subject line of the reply. Required.
	Subject string `json:"subject"`

	// Body is the content of the reply. Required.
	Body Body `json:"body"`

	// AdditionalRecipients are extra addresses to include on the reply.
	AdditionalRecipients []string `json:"additional_recipients,omitempty"`

	// CCAddressEmailList is the comma-separated list of carbon-copy addresses.
	CCAddressEmailList string `json:"cc_address_email_list,omitempty"`

	// BCCAddressEmailList is the comma-separated list of blind carbon-copy addresses.
	BCCAddressEmailList string `json:"bcc_address_email_list,omitempty"`

	// ReminderTS is the timestamp of a reminder to set on the reply.
	ReminderTS string `json:"reminder_ts,omitempty"`

	// AssignedTo identifies the user the reply is assigned to.
	AssignedTo string `json:"assigned_to,omitempty"`
}

// ForwardRequest is the body of a forward request.
type ForwardRequest struct {
	// ReplyToUUID identifies the email being forwarded. Required.
	ReplyToUUID string `json:"reply_to_uuid"`

	// ToAddressEmailList is the comma-separated list of recipient addresses. Required.
	ToAddressEmailList string `json:"to_address_email_list"`

	// EAccount is the sending account the forward is sent from. Required.
	EAccount string `json:"eaccount"`

	// Subject is the subject line of the forwarded email. Required.
	Subject string `json:"subject"`

	// Body is the content prepended to the forwarded email.
	Body *Body `json:"body,omitempty"`

	// CCAddressEmailList is the comma-separated list of carbon-copy addresses.
	CCAddressEmailList string `json:"cc_address_email_list,omitempty"`

	// BCCAddressEmailList is the comma-separated list of blind carbon-copy addresses.
	BCCAddressEmailList string `json:"bcc_address_email_list,omitempty"`

	// ReplyTo is the address replies to the forward are directed to.
	ReplyTo string `json:"reply_to,omitempty"`

	// ForwardedAttachments carries the attachments to forward, which the API
	// does not document as a fixed schema, so it is sent verbatim.
	ForwardedAttachments json.RawMessage `json:"forwarded_attachments,omitempty"`

	// IncludeOriginalBody reports whether the original email content is
	// appended to the forward.
	IncludeOriginalBody *bool `json:"include_original_body,omitempty"`

	// AssignedTo identifies the user the forward is assigned to.
	AssignedTo string `json:"assigned_to,omitempty"`
}

// UpdateRequest is the body of a patch-email request. No field is required; an
// omitted field leaves the current value unchanged.
type UpdateRequest struct {
	// IsUnread marks the email as unread when true and as read when false.
	IsUnread *bool `json:"is_unread,omitempty"`

	// ReminderTS is the reminder timestamp to set on the email.
	ReminderTS *string `json:"reminder_ts,omitempty"`
}

// SendTest sends a test email from one of the workspace sending accounts.
//
// The API reports a sending-account failure inside an otherwise successful HTTP
// 200 body, so a nil return is the only signal of success. The failure code is
// available on instantly.APIError:
//
//	if err := svc.SendTest(ctx, req); err != nil {
//		var apiErr *instantly.APIError
//		if errors.As(err, &apiErr) && apiErr.Code == instantly.ErrCodeAccountAuthError {
//			// the sending account failed to authenticate
//		}
//	}
//
// The endpoint is rate limited to 10 requests per minute per workspace.
//
// See https://developer.instantly.ai/api-reference/email/send-a-test-email
func (s *Service) SendTest(ctx context.Context, req SendTestRequest) error {
	return s.client.Post(ctx, basePath+"/test", req, nil)
}

// List returns a single page of emails filtered by the supplied options.
//
// Pagination is cursor based: pass the returned NextStartingAfter back as the
// starting cursor (WithStartingAfter) to fetch the following page, which is
// empty once the last page has been reached. The endpoint is rate limited to 20
// requests per minute.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}

// Get returns a single email by its unique identifier.
func (s *Service) Get(ctx context.Context, id string) (*Email, error) {
	return instantly.GetResult[Email](ctx, s.client, instantly.JoinPath(basePath, id))
}

// Update patches an email and returns its updated state.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Email, error) {
	return instantly.PatchResult[Email](ctx, s.client, instantly.JoinPath(basePath, id), req)
}

// Delete deletes an email and returns the email that was deleted.
func (s *Service) Delete(ctx context.Context, id string) (*Email, error) {
	return instantly.DeleteResult[Email](ctx, s.client, instantly.JoinPath(basePath, id))
}

// Reply replies to an existing email and returns the reply that was sent.
func (s *Service) Reply(ctx context.Context, req ReplyRequest) (*Email, error) {
	return instantly.PostResult[Email](ctx, s.client, basePath+"/reply", req)
}

// Forward forwards an existing email and returns the forward that was sent.
func (s *Service) Forward(ctx context.Context, req ForwardRequest) (*Email, error) {
	return instantly.PostResult[Email](ctx, s.client, basePath+"/forward", req)
}

// CountUnread returns the number of unread emails in the workspace.
func (s *Service) CountUnread(ctx context.Context) (int64, error) {
	out, err := instantly.GetResult[instantly.CountResponse](ctx, s.client, basePath+"/unread/count")
	if err != nil {
		return 0, err
	}

	return out.Count, nil
}

// MarkThreadAsRead marks every email in a thread as read.
//
// The API reports only a success flag, which carries no more information than a
// nil error, so the response is decoded purely to validate its shape.
func (s *Service) MarkThreadAsRead(ctx context.Context, threadID string) error {
	out := &MarkThreadReadResponse{}
	path := instantly.JoinPath(basePath, "threads", threadID, "mark-as-read")

	return s.client.Post(ctx, path, nil, out)
}
