package instantly

import (
	"context"
	"encoding/json"
	"net/url"
)

// EmailBody is the content of an email.
//
// The API requires HTML when sending an email; Text carries the plain-text
// alternative when one is present.
type EmailBody struct {
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
	Body EmailBody `json:"body"`

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

// EmailListResponse is a single page of emails.
type EmailListResponse struct {
	// Items are the emails on this page.
	Items []Email `json:"items"`

	// NextStartingAfter is the cursor for the following page, and is empty on
	// the last page. It is an opaque value that may be an identifier or a
	// timestamp depending on the request, so it is never parsed.
	NextStartingAfter string `json:"next_starting_after,omitempty"`
}

// UnreadCountResponse is the number of unread emails in the workspace.
type UnreadCountResponse struct {
	// Count is the number of unread emails.
	Count int64 `json:"count"`
}

// MarkThreadReadResponse is the outcome of marking a thread as read.
type MarkThreadReadResponse struct {
	// Success reports whether the thread was marked as read.
	Success bool `json:"success"`
}

// SendTestEmailRequest is the body of a send-test-email request. Every field is
// required by the API.
type SendTestEmailRequest struct {
	// EAccount is the sending account the test email is sent from.
	EAccount string `json:"eaccount"`

	// ToAddressEmailList is the comma-separated list of recipient addresses.
	ToAddressEmailList string `json:"to_address_email_list"`

	// Subject is the subject line of the test email.
	Subject string `json:"subject"`

	// Body is the content of the test email.
	Body EmailBody `json:"body"`
}

// ReplyToEmailRequest is the body of a reply request.
type ReplyToEmailRequest struct {
	// ReplyToUUID identifies the email being replied to. Required.
	ReplyToUUID string `json:"reply_to_uuid"`

	// EAccount is the sending account the reply is sent from. Required.
	EAccount string `json:"eaccount"`

	// Subject is the subject line of the reply. Required.
	Subject string `json:"subject"`

	// Body is the content of the reply. Required.
	Body EmailBody `json:"body"`

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

// ForwardEmailRequest is the body of a forward request.
type ForwardEmailRequest struct {
	// ReplyToUUID identifies the email being forwarded. Required.
	ReplyToUUID string `json:"reply_to_uuid"`

	// ToAddressEmailList is the comma-separated list of recipient addresses. Required.
	ToAddressEmailList string `json:"to_address_email_list"`

	// EAccount is the sending account the forward is sent from. Required.
	EAccount string `json:"eaccount"`

	// Subject is the subject line of the forwarded email. Required.
	Subject string `json:"subject"`

	// Body is the content prepended to the forwarded email.
	Body *EmailBody `json:"body,omitempty"`

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

// UpdateEmailRequest is the body of a patch-email request. No field is
// required; an omitted field leaves the current value unchanged.
type UpdateEmailRequest struct {
	// IsUnread marks the email as unread when true and as read when false.
	IsUnread *bool `json:"is_unread,omitempty"`

	// ReminderTS is the reminder timestamp to set on the email.
	ReminderTS *string `json:"reminder_ts,omitempty"`
}

// SendTestEmail sends a test email from one of the workspace sending accounts.
//
// The API reports a sending-account failure inside an otherwise successful HTTP
// 200 body, so a nil return is the only signal of success. The failure code is
// available on APIError:
//
//	if err := client.SendTestEmail(ctx, req); err != nil {
//		var apiErr *instantly.APIError
//		if errors.As(err, &apiErr) && apiErr.Code == instantly.ErrCodeAccountAuthError {
//			// the sending account failed to authenticate
//		}
//	}
//
// The endpoint is rate limited to 10 requests per minute per workspace.
//
// See https://developer.instantly.ai/api-reference/email/send-a-test-email
func (client *Client) SendTestEmail(ctx context.Context, req SendTestEmailRequest) error {
	return client.post(ctx, "/api/v2/emails/test", req, nil)
}

// ListEmails returns a single page of emails filtered by the supplied options.
//
// Pagination is cursor based: pass the returned NextStartingAfter back as the
// starting cursor to fetch the following page, which is empty once the last
// page has been reached. The endpoint is rate limited to 20 requests per
// minute.
func (client *Client) ListEmails(ctx context.Context, opts ...EmailListOption) (*EmailListResponse, error) {
	response := &EmailListResponse{}
	path := buildURLWithQuery("/api/v2/emails", newEmailListQuery(opts...))

	if err := client.get(ctx, path, response); err != nil {
		return nil, err
	}

	return response, nil
}

// GetEmail returns a single email by its unique identifier.
func (client *Client) GetEmail(ctx context.Context, id string) (*Email, error) {
	email := &Email{}
	if err := client.get(ctx, "/api/v2/emails/"+url.PathEscape(id), email); err != nil {
		return nil, err
	}

	return email, nil
}

// UpdateEmail patches an email and returns its updated state.
func (client *Client) UpdateEmail(ctx context.Context, id string, req UpdateEmailRequest) (*Email, error) {
	email := &Email{}
	if err := client.patch(ctx, "/api/v2/emails/"+url.PathEscape(id), req, email); err != nil {
		return nil, err
	}

	return email, nil
}

// DeleteEmail deletes an email and returns the email that was deleted.
func (client *Client) DeleteEmail(ctx context.Context, id string) (*Email, error) {
	email := &Email{}
	if err := client.delete(ctx, "/api/v2/emails/"+url.PathEscape(id), email); err != nil {
		return nil, err
	}

	return email, nil
}

// ReplyToEmail replies to an existing email and returns the reply that was sent.
func (client *Client) ReplyToEmail(ctx context.Context, req ReplyToEmailRequest) (*Email, error) {
	email := &Email{}
	if err := client.post(ctx, "/api/v2/emails/reply", req, email); err != nil {
		return nil, err
	}

	return email, nil
}

// ForwardEmail forwards an existing email and returns the forward that was sent.
func (client *Client) ForwardEmail(ctx context.Context, req ForwardEmailRequest) (*Email, error) {
	email := &Email{}
	if err := client.post(ctx, "/api/v2/emails/forward", req, email); err != nil {
		return nil, err
	}

	return email, nil
}

// CountUnreadEmails returns the number of unread emails in the workspace.
func (client *Client) CountUnreadEmails(ctx context.Context) (int64, error) {
	response := &UnreadCountResponse{}
	if err := client.get(ctx, "/api/v2/emails/unread/count", response); err != nil {
		return 0, err
	}

	return response.Count, nil
}

// MarkThreadAsRead marks every email in a thread as read.
//
// The API reports only a success flag, which carries no more information than a
// nil error, so the response is decoded purely to validate its shape.
func (client *Client) MarkThreadAsRead(ctx context.Context, threadID string) error {
	response := &MarkThreadReadResponse{}
	path := "/api/v2/emails/threads/" + url.PathEscape(threadID) + "/mark-as-read"

	return client.post(ctx, path, nil, response)
}
