package email_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/email"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// Router patterns and identifiers the email endpoints are exercised with. The
// patterns carry the full request path, including the /api/v2 prefix.
const (
	// listPath is the list/collection endpoint.
	listPath = "/api/v2/emails"

	// countPath is the unread-count endpoint.
	countPath = "/api/v2/emails/unread/count"

	// idPattern is the router pattern for the single-email endpoints.
	idPattern = "/api/v2/emails/:id"

	// sendTestPath is the send-test-email endpoint.
	sendTestPath = "/api/v2/emails/test"

	// replyPath is the reply endpoint.
	replyPath = "/api/v2/emails/reply"

	// forwardPath is the forward endpoint.
	forwardPath = "/api/v2/emails/forward"

	// threadPattern is the router pattern for the mark-thread-as-read endpoint.
	threadPattern = "/api/v2/emails/threads/:thread_id/mark-as-read"

	// emailID identifies the email the single-email endpoints operate on.
	emailID = "email-uuid-1"

	// threadID identifies the thread the mark-as-read endpoint operates on.
	threadID = "thread-uuid-1"

	// eAccount is the sending account requests are made from.
	eAccount = "sender@example.com"

	// leadEmail is the address of the lead emails relate to.
	leadEmail = "lead@example.com"

	// subject is the subject line used across the endpoint fixtures.
	subject = "Quick question"
)

// emailFixture is a spec-shaped email with every documented field populated,
// including the nullable ones.
const emailFixture = `{
	"id": "email-uuid-1",
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"timestamp_email": "2026-08-01T09:59:00.000Z",
	"message_id": "<abc@mail.example.com>",
	"subject": "Quick question",
	"to_address_email_list": "lead@example.com,second@example.com",
	"body": {"html": "<p>Hello</p>", "text": "Hello"},
	"organization_id": "org-uuid-1",
	"eaccount": "sender@example.com",
	"from_address_email": "lead@example.com",
	"cc_address_email_list": "cc@example.com",
	"bcc_address_email_list": "bcc@example.com",
	"reply_to": "replies@example.com",
	"campaign_id": "campaign-uuid-1",
	"subsequence_id": "subsequence-uuid-1",
	"list_id": "list-uuid-1",
	"lead": "lead@example.com",
	"lead_id": "lead-uuid-1",
	"ue_type": 2,
	"step": "1",
	"is_unread": 1,
	"is_auto_reply": 0,
	"reminder_ts": "2026-08-05T09:00:00.000Z",
	"ai_interest_value": 3,
	"ai_assisted": 0,
	"is_focused": 1,
	"i_status": 1,
	"thread_id": "thread-uuid-1",
	"content_preview": "Hello there",
	"attachment_json": [],
	"from_address_json": [{"name": "Lead", "address": "lead@example.com"}],
	"to_address_json": [{"name": "Sender", "address": "sender@example.com"}],
	"cc_address_json": [],
	"ai_agent_id": "agent-uuid-1"
}`

// emailFixtureNulls is the same email with every nullable field explicitly
// null, so an absent value stays distinguishable from a zero value.
const emailFixtureNulls = `{
	"id": "email-uuid-2",
	"timestamp_created": "2026-08-01T11:00:00.000Z",
	"timestamp_email": "2026-08-01T10:59:00.000Z",
	"message_id": "<def@mail.example.com>",
	"subject": "No metadata at all",
	"to_address_email_list": "lead@example.com",
	"body": {"html": "<p>Bare</p>"},
	"organization_id": "org-uuid-1",
	"eaccount": "sender@example.com",
	"from_address_email": null,
	"cc_address_email_list": null,
	"bcc_address_email_list": null,
	"reply_to": null,
	"campaign_id": null,
	"subsequence_id": null,
	"list_id": null,
	"lead": null,
	"lead_id": null,
	"ue_type": null,
	"step": null,
	"is_unread": null,
	"is_auto_reply": null,
	"reminder_ts": null,
	"ai_interest_value": null,
	"ai_assisted": null,
	"is_focused": null,
	"i_status": null,
	"thread_id": null,
	"content_preview": null,
	"ai_agent_id": null
}`

// EmailTestSuite exercises the Email API service against the mock router.
type EmailTestSuite struct {
	instantlytest.Suite
}

// TestEmailSuite runs the Email API suite.
func TestEmailSuite(t *testing.T) {
	suite.Run(t, new(EmailTestSuite))
}

// TestSendTest verifies the request body reaches the API unchanged and a success
// body is reported as success.
func (s *EmailTestSuite) TestSendTest() {
	s.Router.Post(sendTestPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodPost, req.Method)
		s.Equal(sendTestPath, req.URL.Path)

		var received email.SendTestRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(eAccount, received.EAccount)
		s.Equal(leadEmail, received.ToAddressEmailList)
		s.Equal(subject, received.Subject)
		s.Equal("<p>Hello</p>", received.Body.HTML)

		_, _ = w.Write([]byte(instantlytest.SuccessBody))
	})

	err := s.svc().SendTest(context.Background(), email.SendTestRequest{
		EAccount:           eAccount,
		ToAddressEmailList: leadEmail,
		Subject:            subject,
		Body:               email.Body{HTML: "<p>Hello</p>"},
	})

	s.Require().NoError(err)
}

// TestSendTestAccountError verifies a sending-account failure delivered inside
// an HTTP 200 body is still reported as an error. This endpoint is the one the
// API documents as behaving that way.
func (s *EmailTestSuite) TestSendTestAccountError() {
	s.Router.Post(sendTestPath, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"error":"ACC_AUTH_ERROR"}`))
	})

	err := s.svc().SendTest(context.Background(), email.SendTestRequest{
		EAccount:           eAccount,
		ToAddressEmailList: leadEmail,
		Subject:            subject,
		Body:               email.Body{HTML: "<p>Hello</p>"},
	})

	s.Require().Error(err, "an error code at HTTP 200 must still be an error")

	var apiErr *instantly.APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(instantly.ErrCodeAccountAuthError, apiErr.Code)
	s.Equal(int64(http.StatusOK), apiErr.StatusCode)
}

// TestSendTestFailure verifies a rate-limited send surfaces the envelope.
func (s *EmailTestSuite) TestSendTestFailure() {
	s.Router.Post(sendTestPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "10 per minute")
	})

	err := s.svc().SendTest(context.Background(), email.SendTestRequest{EAccount: eAccount})

	instantlytest.AssertAPIError(s.T(), err, http.StatusTooManyRequests)
}

// TestList verifies a page decodes and the cursor is preserved.
func (s *EmailTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodGet, req.Method)
		s.Equal(listPath, req.URL.Path)
		s.Equal("50", req.URL.Query().Get("limit"))
		s.Equal("true", req.URL.Query().Get("is_unread"))

		_, _ = fmt.Fprintf(
			w, `{"items":[%s,%s],"next_starting_after":"cursor-2"}`, emailFixture, emailFixtureNulls,
		)
	})

	page, err := s.svc().List(context.Background(), email.WithLimit(50), email.WithIsUnread(true))

	s.Require().NoError(err)
	s.Require().NotNil(page)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)

	populated := page.Items[0]
	s.Equal(emailID, populated.ID)
	s.Equal(subject, populated.Subject)
	s.Require().NotNil(populated.IsUnread)
	s.InDelta(1, *populated.IsUnread, 0)

	// The nullable fields must stay nil rather than collapsing to a zero value,
	// so "the API said nothing" reads differently from "the API said zero".
	bare := page.Items[1]
	s.Nil(bare.IsUnread)
	s.Nil(bare.CampaignID)
	s.Nil(bare.ReminderTS)
	s.Nil(bare.ThreadID)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string.
func (s *EmailTestSuite) TestListWithoutOptions() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery, "an unfiltered list must not send an empty query string")
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background())

	s.Require().NoError(err)
	s.Require().NotNil(page)
	s.Empty(page.Items)
	s.Empty(page.NextStartingAfter)
}

// TestListIgnoresNilOption verifies a nil option is skipped rather than causing
// a panic.
func (s *EmailTestSuite) TestListIgnoresNilOption() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery)
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background(), nil)

	s.Require().NoError(err)
	s.Require().NotNil(page)
}

// TestListFailure verifies a rate-limited list returns no page.
func (s *EmailTestSuite) TestListFailure() {
	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "20 per minute")
	})

	page, err := s.svc().List(context.Background())

	instantlytest.AssertAPIError(s.T(), err, http.StatusTooManyRequests)
	s.Nil(page, "a failed list must not hand back a page")
}

// TestGet verifies a single email decodes, including its nullable fields.
func (s *EmailTestSuite) TestGet() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodGet, req.Method)
		s.Equal(emailID, instantlytest.PathParam(req, "id"))

		_, _ = w.Write([]byte(emailFixture))
	})

	got, err := s.svc().Get(context.Background(), emailID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(emailID, got.ID)
	s.Equal(subject, got.Subject)
	s.Equal("lead@example.com,second@example.com", got.ToAddressEmailList)
	s.Equal("<p>Hello</p>", got.Body.HTML)
	s.Equal("Hello", got.Body.Text)
	s.Equal("org-uuid-1", got.OrganizationID)
	s.Equal(eAccount, got.EAccount)

	s.Require().NotNil(got.FromAddressEmail)
	s.Equal(leadEmail, *got.FromAddressEmail)
	s.Require().NotNil(got.CampaignID)
	s.Equal("campaign-uuid-1", *got.CampaignID)
	s.Require().NotNil(got.ThreadID)
	s.Equal(threadID, *got.ThreadID)
	s.Require().NotNil(got.AIInterestValue)
	s.InDelta(3, *got.AIInterestValue, 0)
	s.JSONEq(`[{"name":"Lead","address":"lead@example.com"}]`, string(got.FromAddressJSON))
}

// TestGetFailure verifies a missing email returns no value.
func (s *EmailTestSuite) TestGetFailure() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "Email not found")
	})

	got, err := s.svc().Get(context.Background(), "missing-uuid")

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got, "a failed lookup must not hand back an email")
}

// TestUpdate verifies the patch body is sent and the updated email decodes.
func (s *EmailTestSuite) TestUpdate() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodPatch, req.Method)
		s.Equal(emailID, instantlytest.PathParam(req, "id"))

		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(false, received["is_unread"])
		s.Equal("2026-08-05T09:00:00.000Z", received["reminder_ts"])

		_, _ = w.Write([]byte(emailFixture))
	})

	got, err := s.svc().Update(context.Background(), emailID, email.UpdateRequest{
		IsUnread:   instantly.Ptr(false),
		ReminderTS: instantly.Ptr("2026-08-05T09:00:00.000Z"),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(emailID, got.ID)
}

// TestUpdateOmitsUnsetFields verifies an empty patch sends no field at all, so
// an untouched field is never overwritten.
func (s *EmailTestSuite) TestUpdateOmitsUnsetFields() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Empty(received, "an unset patch field must not be sent")

		_, _ = w.Write([]byte(emailFixture))
	})

	got, err := s.svc().Update(context.Background(), emailID, email.UpdateRequest{})

	s.Require().NoError(err)
	s.Require().NotNil(got)
}

// TestUpdateFailure verifies a missing email returns no value.
func (s *EmailTestSuite) TestUpdateFailure() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "Email not found")
	})

	got, err := s.svc().Update(context.Background(), "missing-uuid", email.UpdateRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestDelete verifies the deleted email is returned to the caller.
func (s *EmailTestSuite) TestDelete() {
	s.Router.Delete(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodDelete, req.Method)
		s.Equal(emailID, instantlytest.PathParam(req, "id"))

		_, _ = w.Write([]byte(emailFixture))
	})

	got, err := s.svc().Delete(context.Background(), emailID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(emailID, got.ID)
}

// TestDeleteFailure verifies a missing email returns no value.
func (s *EmailTestSuite) TestDeleteFailure() {
	s.Router.Delete(idPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "Email not found")
	})

	got, err := s.svc().Delete(context.Background(), "missing-uuid")

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestReply verifies every documented reply field reaches the API.
func (s *EmailTestSuite) TestReply() {
	s.Router.Post(replyPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodPost, req.Method)
		s.Equal(replyPath, req.URL.Path)

		var received email.ReplyRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(emailID, received.ReplyToUUID)
		s.Equal(eAccount, received.EAccount)
		s.Equal("Re: "+subject, received.Subject)
		s.Equal("<p>Answering</p>", received.Body.HTML)
		s.Equal([]string{"third@example.com"}, received.AdditionalRecipients)
		s.Equal("cc@example.com", received.CCAddressEmailList)
		s.Equal("2026-08-05T09:00:00.000Z", received.ReminderTS)

		_, _ = w.Write([]byte(emailFixture))
	})

	got, err := s.svc().Reply(context.Background(), email.ReplyRequest{
		ReplyToUUID:          emailID,
		EAccount:             eAccount,
		Subject:              "Re: " + subject,
		Body:                 email.Body{HTML: "<p>Answering</p>"},
		AdditionalRecipients: []string{"third@example.com"},
		CCAddressEmailList:   "cc@example.com",
		ReminderTS:           "2026-08-05T09:00:00.000Z",
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(emailID, got.ID)
}

// TestReplyOmitsUnsetFields verifies the optional reply fields are left out of
// the body entirely when they are not set.
func (s *EmailTestSuite) TestReplyOmitsUnsetFields() {
	s.Router.Post(replyPath, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.NotContains(received, "cc_address_email_list")
		s.NotContains(received, "additional_recipients")
		s.NotContains(received, "assigned_to")

		_, _ = w.Write([]byte(emailFixture))
	})

	got, err := s.svc().Reply(context.Background(), email.ReplyRequest{
		ReplyToUUID: emailID,
		EAccount:    eAccount,
		Subject:     "Re: " + subject,
		Body:        email.Body{HTML: "<p>Answering</p>"},
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
}

// TestReplyFailure verifies a plan-limited reply returns no value.
func (s *EmailTestSuite) TestReplyFailure() {
	s.Router.Post(replyPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusPaymentRequired, "Payment Required", "Plan limit reached")
	})

	got, err := s.svc().Reply(context.Background(), email.ReplyRequest{ReplyToUUID: emailID})

	instantlytest.AssertAPIError(s.T(), err, http.StatusPaymentRequired)
	s.Nil(got)
}

// TestForward verifies every documented forward field reaches the API.
func (s *EmailTestSuite) TestForward() {
	s.Router.Post(forwardPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodPost, req.Method)
		s.Equal(forwardPath, req.URL.Path)

		var received email.ForwardRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(emailID, received.ReplyToUUID)
		s.Equal("third@example.com", received.ToAddressEmailList)
		s.Equal(eAccount, received.EAccount)
		s.Equal("Fwd: "+subject, received.Subject)
		// A handler runs on the server goroutine, so a failure here is recorded
		// rather than aborting the test from underneath the client.
		if s.NotNil(received.Body) {
			s.Equal("<p>See below</p>", received.Body.HTML)
		}
		if s.NotNil(received.IncludeOriginalBody) {
			s.True(*received.IncludeOriginalBody)
		}
		s.JSONEq(`[{"name":"quote.pdf"}]`, string(received.ForwardedAttachments))

		_, _ = w.Write([]byte(emailFixture))
	})

	got, err := s.svc().Forward(context.Background(), email.ForwardRequest{
		ReplyToUUID:          emailID,
		ToAddressEmailList:   "third@example.com",
		EAccount:             eAccount,
		Subject:              "Fwd: " + subject,
		Body:                 &email.Body{HTML: "<p>See below</p>"},
		ForwardedAttachments: json.RawMessage(`[{"name":"quote.pdf"}]`),
		IncludeOriginalBody:  instantly.Ptr(true),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(emailID, got.ID)
}

// TestForwardFailure verifies an unauthorized forward returns no value.
func (s *EmailTestSuite) TestForwardFailure() {
	s.Router.Post(forwardPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusUnauthorized, "Unauthorized", "Invalid API key")
	})

	got, err := s.svc().Forward(context.Background(), email.ForwardRequest{ReplyToUUID: emailID})

	instantlytest.AssertAPIError(s.T(), err, http.StatusUnauthorized)
	s.Nil(got)
}

// TestCountUnread verifies the unread count is unwrapped for the caller.
func (s *EmailTestSuite) TestCountUnread() {
	s.Router.Get(countPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodGet, req.Method)
		s.Equal(countPath, req.URL.Path)

		_, _ = w.Write([]byte(`{"count":42}`))
	})

	count, err := s.svc().CountUnread(context.Background())

	s.Require().NoError(err)
	s.Equal(int64(42), count)
}

// TestCountUnreadFailure verifies a failed count reports zero and an error.
func (s *EmailTestSuite) TestCountUnreadFailure() {
	s.Router.Get(countPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusUnauthorized, "Unauthorized", "Invalid API key")
	})

	count, err := s.svc().CountUnread(context.Background())

	instantlytest.AssertAPIError(s.T(), err, http.StatusUnauthorized)
	s.Zero(count, "a failed count must not report a number the API never sent")
}

// TestMarkThreadAsRead verifies the thread endpoint is called without a body.
func (s *EmailTestSuite) TestMarkThreadAsRead() {
	s.Router.Post(threadPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodPost, req.Method)
		s.Equal(threadID, instantlytest.PathParam(req, "thread_id"))

		body, err := instantlytest.ReadAll(req)
		s.NoError(err)
		s.Empty(body, "marking a thread as read sends no request body")

		_, _ = w.Write([]byte(`{"success":true}`))
	})

	err := s.svc().MarkThreadAsRead(context.Background(), threadID)

	s.Require().NoError(err)
}

// TestMarkThreadAsReadFailure verifies a missing thread surfaces the envelope.
func (s *EmailTestSuite) TestMarkThreadAsReadFailure() {
	s.Router.Post(threadPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "Thread not found")
	})

	err := s.svc().MarkThreadAsRead(context.Background(), "missing-thread")

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
}

// TestPathParametersAreEscaped verifies a caller-supplied identifier cannot
// rewrite the request path. The transport is intercepted so the raw path can be
// asserted before a server decodes it.
func (s *EmailTestSuite) TestPathParametersAreEscaped() {
	tests := []struct {
		name     string
		call     func(svc *email.Service) error
		expected string
	}{
		{
			name: "email identifier",
			call: func(svc *email.Service) error {
				_, err := svc.Get(context.Background(), "../admin?x=1")
				return err
			},
			expected: "/api/v2/emails/..%2Fadmin%3Fx=1",
		},
		{
			name: "thread identifier",
			call: func(svc *email.Service) error {
				return svc.MarkThreadAsRead(context.Background(), "../admin?x=1")
			},
			expected: "/api/v2/emails/threads/..%2Fadmin%3Fx=1/mark-as-read",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			var requestURI string

			client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
				&http.Client{Transport: instantlytest.RoundTripFunc(
					func(req *http.Request) (*http.Response, error) {
						requestURI = req.URL.EscapedPath()
						return instantlytest.JSONResponse(http.StatusOK, emailFixture), nil
					},
				)},
			))

			s.Require().NoError(test.call(email.New(client)))
			s.Equal(test.expected, requestURI)
		})
	}
}

// TestListOptions verifies each documented list query parameter is rendered by
// exactly one option, under the key and value the API expects.
func (s *EmailTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option email.ListOption
		key    string
		value  string
	}{
		{"limit", email.WithLimit(50), "limit", "50"},
		{"starting after", email.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"search", email.WithSearch("quick question"), "search", "quick question"},
		{"campaign id", email.WithCampaignID("campaign-uuid-1"), "campaign_id", "campaign-uuid-1"},
		{"list id", email.WithListID("list-uuid-1"), "list_id", "list-uuid-1"},
		{"interest status", email.WithIStatus(1), "i_status", "1"},
		{"sending account", email.WithAccount(eAccount), "eaccount", eAccount},
		{"is unread", email.WithIsUnread(true), "is_unread", "true"},
		{"has reminder", email.WithHasReminder(false), "has_reminder", "false"},
		{"mode", email.WithMode(email.ModeFocused), "mode", "emode_focused"},
		{"preview only", email.WithPreviewOnly(true), "preview_only", "true"},
		{"sort order", email.WithSortOrder(instantly.SortOrderAsc), "sort_order", "asc"},
		{"scheduled only", email.WithScheduledOnly(true), "scheduled_only", "true"},
		{"assigned to", email.WithAssignedTo("user-uuid-1"), "assigned_to", "user-uuid-1"},
		{"lead", email.WithLead(leadEmail), "lead", leadEmail},
		{"company domain", email.WithCompanyDomain("example.com"), "company_domain", "example.com"},
		{"marked as done", email.WithMarkedAsDone(true), "marked_as_done", "true"},
		{"email type", email.WithType(email.TypeReceived), "email_type", "received"},
		{
			"min timestamp created",
			email.WithMinTimestampCreated("2026-01-01T00:00:00.000Z"),
			"min_timestamp_created", "2026-01-01T00:00:00.000Z",
		},
		{
			"max timestamp created",
			email.WithMaxTimestampCreated("2026-12-31T23:59:59.000Z"),
			"max_timestamp_created", "2026-12-31T23:59:59.000Z",
		},
		{"latest of thread", email.WithLatestOfThread(true), "latest_of_thread", "true"},
	}

	s.Require().Len(tests, 21, "every documented list query parameter needs an option")

	for _, test := range tests {
		s.Run(test.name, func() {
			q := instantly.NewQuery()
			test.option(q)

			s.Require().Equal(1, q.Len(), "an option must render exactly one query parameter")
			s.Equal(test.value, q.Get(test.key))
		})
	}
}

// TestListOptionsEnums verifies the enum options serialize to the exact values
// the API documents.
func (s *EmailTestSuite) TestListOptionsEnums() {
	tests := []struct {
		name   string
		option email.ListOption
		key    string
		value  string
	}{
		{"focused inbox", email.WithMode(email.ModeFocused), "mode", "emode_focused"},
		{"other inboxes", email.WithMode(email.ModeOthers), "mode", "emode_others"},
		{"every inbox", email.WithMode(email.ModeAll), "mode", "emode_all"},
		{"ascending", email.WithSortOrder(instantly.SortOrderAsc), "sort_order", "asc"},
		{"descending", email.WithSortOrder(instantly.SortOrderDesc), "sort_order", "desc"},
		{"received", email.WithType(email.TypeReceived), "email_type", "received"},
		{"sent", email.WithType(email.TypeSent), "email_type", "sent"},
		{"manual", email.WithType(email.TypeManual), "email_type", "manual"},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			q := instantly.NewQuery()
			test.option(q)
			s.Equal(test.value, q.Get(test.key))
		})
	}
}

// TestListOptionsCombined verifies several options render together, and that the
// last value supplied for a parameter is the one sent.
func (s *EmailTestSuite) TestListOptionsCombined() {
	q := instantly.NewQuery()
	for _, opt := range []email.ListOption{
		email.WithLimit(25),
		email.WithIsUnread(true),
		email.WithMode(email.ModeAll),
		email.WithSortOrder(instantly.SortOrderDesc),
		email.WithLimit(100),
	} {
		opt(q)
	}

	s.Require().Equal(4, q.Len())
	s.Equal("100", q.Get("limit"), "the last value supplied wins")
	s.Equal("true", q.Get("is_unread"))
	s.Equal("emode_all", q.Get("mode"))
	s.Equal("desc", q.Get("sort_order"))
}

// TestListOptionsReachTheAPI verifies the rendered options survive the trip to
// the server as an encoded query string.
func (s *EmailTestSuite) TestListOptionsReachTheAPI() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()
		s.Equal("25", query.Get("limit"))
		s.Equal("quick question", query.Get("search"))
		s.Equal(eAccount, query.Get("eaccount"))
		s.Equal("emode_focused", query.Get("mode"))
		s.Empty(query.Get("company_domain"))

		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	_, err := s.svc().List(
		context.Background(),
		email.WithLimit(25),
		email.WithSearch("quick question"),
		email.WithAccount(eAccount),
		email.WithMode(email.ModeFocused),
	)

	s.Require().NoError(err)
}

// svc builds an Email service pointed at the suite's mock client.
func (s *EmailTestSuite) svc() *email.Service {
	return email.New(s.Client)
}
