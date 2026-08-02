package instantly

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Router patterns and identifiers the email endpoints are exercised with. The
// patterns carry the full request path, including the /api/v2 prefix.
const (
	// testEmailIDPattern is the router pattern for the single-email endpoints.
	testEmailIDPattern = "/api/v2/emails/:id"

	// testSendTestPath is the send-test-email endpoint.
	testSendTestPath = "/api/v2/emails/test"

	// testReplyPath is the reply endpoint.
	testReplyPath = "/api/v2/emails/reply"

	// testForwardPath is the forward endpoint.
	testForwardPath = "/api/v2/emails/forward"

	// testThreadPattern is the router pattern for the mark-thread-as-read endpoint.
	testThreadPattern = "/api/v2/emails/threads/:thread_id/mark-as-read"

	// testEmailID identifies the email the single-email endpoints operate on.
	testEmailID = "email-uuid-1"

	// testThreadID identifies the thread the mark-as-read endpoint operates on.
	testThreadID = "thread-uuid-1"

	// testEAccount is the sending account requests are made from.
	testEAccount = "sender@example.com"

	// testLeadEmail is the address of the lead emails relate to.
	testLeadEmail = "lead@example.com"

	// testSubject is the subject line used across the endpoint fixtures.
	testSubject = "Quick question"
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

// TestSendTestEmail verifies the request body reaches the API unchanged and a
// success body is reported as success.
func (s *InstantlyTestSuite) TestSendTestEmail() {
	s.mux.Post(testSendTestPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodPost, req.Method)
		s.Equal(testSendTestPath, req.URL.Path)

		var received SendTestEmailRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(testEAccount, received.EAccount)
		s.Equal(testLeadEmail, received.ToAddressEmailList)
		s.Equal(testSubject, received.Subject)
		s.Equal("<p>Hello</p>", received.Body.HTML)

		_, _ = w.Write([]byte(successBody))
	})

	err := s.client.SendTestEmail(context.Background(), SendTestEmailRequest{
		EAccount:           testEAccount,
		ToAddressEmailList: testLeadEmail,
		Subject:            testSubject,
		Body:               EmailBody{HTML: "<p>Hello</p>"},
	})

	s.Require().NoError(err)
}

// TestSendTestEmailAccountError verifies a sending-account failure delivered
// inside an HTTP 200 body is still reported as an error. This endpoint is the
// one the API documents as behaving that way.
func (s *InstantlyTestSuite) TestSendTestEmailAccountError() {
	// The body carries the wire value verbatim, so the exported constant is
	// pinned to what the API actually sends rather than to itself.
	s.mux.Post(testSendTestPath, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"error":"ACC_AUTH_ERROR"}`))
	})

	err := s.client.SendTestEmail(context.Background(), SendTestEmailRequest{
		EAccount:           testEAccount,
		ToAddressEmailList: testLeadEmail,
		Subject:            testSubject,
		Body:               EmailBody{HTML: "<p>Hello</p>"},
	})

	s.Require().Error(err, "an error code at HTTP 200 must still be an error")

	var apiErr *APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(ErrCodeAccountAuthError, apiErr.Code)
	s.Equal(int64(http.StatusOK), apiErr.StatusCode)
}

// TestSendTestEmailFailure verifies a rate-limited send surfaces the envelope.
func (s *InstantlyTestSuite) TestSendTestEmailFailure() {
	s.mux.Post(testSendTestPath, func(w http.ResponseWriter, _ *http.Request) {
		writeAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "10 requests per minute")
	})

	err := s.client.SendTestEmail(context.Background(), SendTestEmailRequest{EAccount: testEAccount})

	assertAPIError(s, err, http.StatusTooManyRequests)
}

// TestListEmails verifies a page decodes and the cursor is preserved.
func (s *InstantlyTestSuite) TestListEmails() {
	s.mux.Get(testPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodGet, req.Method)
		s.Equal(testPath, req.URL.Path)
		s.Equal("50", req.URL.Query().Get("limit"))
		s.Equal("true", req.URL.Query().Get("is_unread"))

		_, _ = fmt.Fprintf(
			w, `{"items":[%s,%s],"next_starting_after":"cursor-2"}`, emailFixture, emailFixtureNulls,
		)
	})

	page, err := s.client.ListEmails(
		context.Background(), WithEmailLimit(50), WithEmailIsUnread(true),
	)

	s.Require().NoError(err)
	s.Require().NotNil(page)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)

	populated := page.Items[0]
	s.Equal(testEmailID, populated.ID)
	s.Equal(testSubject, populated.Subject)
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

// TestListEmailsWithoutOptions verifies an unfiltered list sends no query string.
func (s *InstantlyTestSuite) TestListEmailsWithoutOptions() {
	s.mux.Get(testPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery, "an unfiltered list must not send an empty query string")
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.client.ListEmails(context.Background())

	s.Require().NoError(err)
	s.Require().NotNil(page)
	s.Empty(page.Items)
	s.Empty(page.NextStartingAfter)
}

// TestListEmailsFailure verifies a rate-limited list returns no page.
func (s *InstantlyTestSuite) TestListEmailsFailure() {
	s.mux.Get(testPath, func(w http.ResponseWriter, _ *http.Request) {
		writeAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "20 requests per minute")
	})

	page, err := s.client.ListEmails(context.Background())

	assertAPIError(s, err, http.StatusTooManyRequests)
	s.Nil(page, "a failed list must not hand back a page")
}

// TestGetEmail verifies a single email decodes, including its nullable fields.
func (s *InstantlyTestSuite) TestGetEmail() {
	s.mux.Get(testEmailIDPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodGet, req.Method)
		s.Equal(testEmailID, GetPathParam(req, "id"))

		_, _ = w.Write([]byte(emailFixture))
	})

	email, err := s.client.GetEmail(context.Background(), testEmailID)

	s.Require().NoError(err)
	s.Require().NotNil(email)
	s.Equal(testEmailID, email.ID)
	s.Equal(testSubject, email.Subject)
	s.Equal("lead@example.com,second@example.com", email.ToAddressEmailList)
	s.Equal("<p>Hello</p>", email.Body.HTML)
	s.Equal("Hello", email.Body.Text)
	s.Equal("org-uuid-1", email.OrganizationID)
	s.Equal(testEAccount, email.EAccount)

	s.Require().NotNil(email.FromAddressEmail)
	s.Equal(testLeadEmail, *email.FromAddressEmail)
	s.Require().NotNil(email.CampaignID)
	s.Equal("campaign-uuid-1", *email.CampaignID)
	s.Require().NotNil(email.ThreadID)
	s.Equal(testThreadID, *email.ThreadID)
	s.Require().NotNil(email.AIInterestValue)
	s.InDelta(3, *email.AIInterestValue, 0)
	s.JSONEq(`[{"name":"Lead","address":"lead@example.com"}]`, string(email.FromAddressJSON))
}

// TestGetEmailFailure verifies a missing email returns no value.
func (s *InstantlyTestSuite) TestGetEmailFailure() {
	s.mux.Get(testEmailIDPattern, func(w http.ResponseWriter, _ *http.Request) {
		writeAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "Email not found")
	})

	email, err := s.client.GetEmail(context.Background(), "missing-uuid")

	assertAPIError(s, err, http.StatusNotFound)
	s.Nil(email, "a failed lookup must not hand back an email")
}

// TestUpdateEmail verifies the patch body is sent and the updated email decodes.
func (s *InstantlyTestSuite) TestUpdateEmail() {
	s.mux.Patch(testEmailIDPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodPatch, req.Method)
		s.Equal(testEmailID, GetPathParam(req, "id"))

		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(false, received["is_unread"])
		s.Equal("2026-08-05T09:00:00.000Z", received["reminder_ts"])

		_, _ = w.Write([]byte(emailFixture))
	})

	email, err := s.client.UpdateEmail(context.Background(), testEmailID, UpdateEmailRequest{
		IsUnread:   ptrTo(false),
		ReminderTS: ptrTo("2026-08-05T09:00:00.000Z"),
	})

	s.Require().NoError(err)
	s.Require().NotNil(email)
	s.Equal(testEmailID, email.ID)
}

// TestUpdateEmailOmitsUnsetFields verifies an empty patch sends no field at all,
// so an untouched field is never overwritten.
func (s *InstantlyTestSuite) TestUpdateEmailOmitsUnsetFields() {
	s.mux.Patch(testEmailIDPattern, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Empty(received, "an unset patch field must not be sent")

		_, _ = w.Write([]byte(emailFixture))
	})

	email, err := s.client.UpdateEmail(context.Background(), testEmailID, UpdateEmailRequest{})

	s.Require().NoError(err)
	s.Require().NotNil(email)
}

// TestUpdateEmailFailure verifies a missing email returns no value.
func (s *InstantlyTestSuite) TestUpdateEmailFailure() {
	s.mux.Patch(testEmailIDPattern, func(w http.ResponseWriter, _ *http.Request) {
		writeAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "Email not found")
	})

	email, err := s.client.UpdateEmail(context.Background(), "missing-uuid", UpdateEmailRequest{})

	assertAPIError(s, err, http.StatusNotFound)
	s.Nil(email)
}

// TestDeleteEmail verifies the deleted email is returned to the caller.
func (s *InstantlyTestSuite) TestDeleteEmail() {
	s.mux.Delete(testEmailIDPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodDelete, req.Method)
		s.Equal(testEmailID, GetPathParam(req, "id"))

		_, _ = w.Write([]byte(emailFixture))
	})

	email, err := s.client.DeleteEmail(context.Background(), testEmailID)

	s.Require().NoError(err)
	s.Require().NotNil(email)
	s.Equal(testEmailID, email.ID)
}

// TestDeleteEmailFailure verifies a missing email returns no value.
func (s *InstantlyTestSuite) TestDeleteEmailFailure() {
	s.mux.Delete(testEmailIDPattern, func(w http.ResponseWriter, _ *http.Request) {
		writeAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "Email not found")
	})

	email, err := s.client.DeleteEmail(context.Background(), "missing-uuid")

	assertAPIError(s, err, http.StatusNotFound)
	s.Nil(email)
}

// TestReplyToEmail verifies every documented reply field reaches the API.
func (s *InstantlyTestSuite) TestReplyToEmail() {
	s.mux.Post(testReplyPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodPost, req.Method)
		s.Equal(testReplyPath, req.URL.Path)

		var received ReplyToEmailRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(testEmailID, received.ReplyToUUID)
		s.Equal(testEAccount, received.EAccount)
		s.Equal("Re: "+testSubject, received.Subject)
		s.Equal("<p>Answering</p>", received.Body.HTML)
		s.Equal([]string{"third@example.com"}, received.AdditionalRecipients)
		s.Equal("cc@example.com", received.CCAddressEmailList)
		s.Equal("2026-08-05T09:00:00.000Z", received.ReminderTS)

		_, _ = w.Write([]byte(emailFixture))
	})

	email, err := s.client.ReplyToEmail(context.Background(), ReplyToEmailRequest{
		ReplyToUUID:          testEmailID,
		EAccount:             testEAccount,
		Subject:              "Re: " + testSubject,
		Body:                 EmailBody{HTML: "<p>Answering</p>"},
		AdditionalRecipients: []string{"third@example.com"},
		CCAddressEmailList:   "cc@example.com",
		ReminderTS:           "2026-08-05T09:00:00.000Z",
	})

	s.Require().NoError(err)
	s.Require().NotNil(email)
	s.Equal(testEmailID, email.ID)
}

// TestReplyToEmailOmitsUnsetFields verifies the optional reply fields are left
// out of the body entirely when they are not set.
func (s *InstantlyTestSuite) TestReplyToEmailOmitsUnsetFields() {
	s.mux.Post(testReplyPath, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.NotContains(received, "cc_address_email_list")
		s.NotContains(received, "additional_recipients")
		s.NotContains(received, "assigned_to")

		_, _ = w.Write([]byte(emailFixture))
	})

	email, err := s.client.ReplyToEmail(context.Background(), ReplyToEmailRequest{
		ReplyToUUID: testEmailID,
		EAccount:    testEAccount,
		Subject:     "Re: " + testSubject,
		Body:        EmailBody{HTML: "<p>Answering</p>"},
	})

	s.Require().NoError(err)
	s.Require().NotNil(email)
}

// TestReplyToEmailFailure verifies a plan-limited reply returns no value.
func (s *InstantlyTestSuite) TestReplyToEmailFailure() {
	s.mux.Post(testReplyPath, func(w http.ResponseWriter, _ *http.Request) {
		writeAPIErrorEnvelope(w, http.StatusPaymentRequired, "Payment Required", "Plan limit reached")
	})

	email, err := s.client.ReplyToEmail(context.Background(), ReplyToEmailRequest{ReplyToUUID: testEmailID})

	assertAPIError(s, err, http.StatusPaymentRequired)
	s.Nil(email)
}

// TestForwardEmail verifies every documented forward field reaches the API.
func (s *InstantlyTestSuite) TestForwardEmail() {
	s.mux.Post(testForwardPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodPost, req.Method)
		s.Equal(testForwardPath, req.URL.Path)

		var received ForwardEmailRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(testEmailID, received.ReplyToUUID)
		s.Equal("third@example.com", received.ToAddressEmailList)
		s.Equal(testEAccount, received.EAccount)
		s.Equal("Fwd: "+testSubject, received.Subject)
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

	email, err := s.client.ForwardEmail(context.Background(), ForwardEmailRequest{
		ReplyToUUID:          testEmailID,
		ToAddressEmailList:   "third@example.com",
		EAccount:             testEAccount,
		Subject:              "Fwd: " + testSubject,
		Body:                 &EmailBody{HTML: "<p>See below</p>"},
		ForwardedAttachments: json.RawMessage(`[{"name":"quote.pdf"}]`),
		IncludeOriginalBody:  ptrTo(true),
	})

	s.Require().NoError(err)
	s.Require().NotNil(email)
	s.Equal(testEmailID, email.ID)
}

// TestForwardEmailFailure verifies an unauthorized forward returns no value.
func (s *InstantlyTestSuite) TestForwardEmailFailure() {
	s.mux.Post(testForwardPath, func(w http.ResponseWriter, _ *http.Request) {
		writeAPIErrorEnvelope(w, http.StatusUnauthorized, "Unauthorized", "Invalid API key")
	})

	email, err := s.client.ForwardEmail(context.Background(), ForwardEmailRequest{ReplyToUUID: testEmailID})

	assertAPIError(s, err, http.StatusUnauthorized)
	s.Nil(email)
}

// TestCountUnreadEmails verifies the unread count is unwrapped for the caller.
func (s *InstantlyTestSuite) TestCountUnreadEmails() {
	s.mux.Get(testCountPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodGet, req.Method)
		s.Equal(testCountPath, req.URL.Path)

		_, _ = w.Write([]byte(`{"count":42}`))
	})

	count, err := s.client.CountUnreadEmails(context.Background())

	s.Require().NoError(err)
	s.Equal(int64(42), count)
}

// TestCountUnreadEmailsFailure verifies a failed count reports zero and an error.
func (s *InstantlyTestSuite) TestCountUnreadEmailsFailure() {
	s.mux.Get(testCountPath, func(w http.ResponseWriter, _ *http.Request) {
		writeAPIErrorEnvelope(w, http.StatusUnauthorized, "Unauthorized", "Invalid API key")
	})

	count, err := s.client.CountUnreadEmails(context.Background())

	assertAPIError(s, err, http.StatusUnauthorized)
	s.Zero(count, "a failed count must not report a number the API never sent")
}

// TestMarkThreadAsRead verifies the thread endpoint is called without a body.
func (s *InstantlyTestSuite) TestMarkThreadAsRead() {
	s.mux.Post(testThreadPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodPost, req.Method)
		s.Equal(testThreadID, GetPathParam(req, "thread_id"))

		body, err := readAll(req)
		s.NoError(err)
		s.Empty(body, "marking a thread as read sends no request body")

		_, _ = w.Write([]byte(`{"success":true}`))
	})

	err := s.client.MarkThreadAsRead(context.Background(), testThreadID)

	s.Require().NoError(err)
}

// TestMarkThreadAsReadFailure verifies a missing thread surfaces the envelope.
func (s *InstantlyTestSuite) TestMarkThreadAsReadFailure() {
	s.mux.Post(testThreadPattern, func(w http.ResponseWriter, _ *http.Request) {
		writeAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "Thread not found")
	})

	err := s.client.MarkThreadAsRead(context.Background(), "missing-thread")

	assertAPIError(s, err, http.StatusNotFound)
}

// TestEmailPathParametersAreEscaped verifies a caller-supplied identifier cannot
// rewrite the request path. The transport is intercepted so the raw path can be
// asserted before a server decodes it.
func (s *InstantlyTestSuite) TestEmailPathParametersAreEscaped() {
	tests := []struct {
		name     string
		call     func(client *Client) error
		expected string
	}{
		{
			name: "email identifier",
			call: func(client *Client) error {
				_, err := client.GetEmail(context.Background(), "../admin?x=1")
				return err
			},
			expected: "/api/v2/emails/..%2Fadmin%3Fx=1",
		},
		{
			name: "thread identifier",
			call: func(client *Client) error {
				return client.MarkThreadAsRead(context.Background(), "../admin?x=1")
			},
			expected: "/api/v2/emails/threads/..%2Fadmin%3Fx=1/mark-as-read",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			var requestURI string

			client := NewClient(testAPIKey)
			client.HTTPClient = &http.Client{Transport: roundTripFunc(
				func(req *http.Request) (*http.Response, error) {
					requestURI = req.URL.EscapedPath()
					return jsonResponse(http.StatusOK, emailFixture), nil
				},
			)}

			s.Require().NoError(test.call(client))
			s.Equal(test.expected, requestURI)
		})
	}
}

// TestEmailListOptions verifies each of the documented list query parameters is
// rendered by exactly one option, under the key and value the API expects.
func (s *InstantlyTestSuite) TestEmailListOptions() {
	tests := []struct {
		name   string
		option EmailListOption
		key    string
		value  string
	}{
		{"limit", WithEmailLimit(50), "limit", "50"},
		{"starting after", WithEmailStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"search", WithEmailSearch("quick question"), "search", "quick question"},
		{"campaign id", WithEmailCampaignID("campaign-uuid-1"), "campaign_id", "campaign-uuid-1"},
		{"list id", WithEmailListID("list-uuid-1"), "list_id", "list-uuid-1"},
		{"interest status", WithEmailIStatus(1), "i_status", "1"},
		{"sending account", WithEmailAccount(testEAccount), "eaccount", testEAccount},
		{"is unread", WithEmailIsUnread(true), "is_unread", "true"},
		{"has reminder", WithEmailHasReminder(false), "has_reminder", "false"},
		{"mode", WithEmailMode(EmailModeFocused), "mode", "emode_focused"},
		{"preview only", WithEmailPreviewOnly(true), "preview_only", "true"},
		{"sort order", WithEmailSortOrder(SortOrderAsc), "sort_order", "asc"},
		{"scheduled only", WithEmailScheduledOnly(true), "scheduled_only", "true"},
		{"assigned to", WithEmailAssignedTo("user-uuid-1"), "assigned_to", "user-uuid-1"},
		{"lead", WithEmailLead(testLeadEmail), "lead", testLeadEmail},
		{"company domain", WithEmailCompanyDomain("example.com"), "company_domain", "example.com"},
		{"marked as done", WithEmailMarkedAsDone(true), "marked_as_done", "true"},
		{"email type", WithEmailType(EmailTypeReceived), "email_type", "received"},
		{
			"min timestamp created",
			WithEmailMinTimestampCreated("2026-01-01T00:00:00.000Z"),
			"min_timestamp_created", "2026-01-01T00:00:00.000Z",
		},
		{
			"max timestamp created",
			WithEmailMaxTimestampCreated("2026-12-31T23:59:59.000Z"),
			"max_timestamp_created", "2026-12-31T23:59:59.000Z",
		},
		{"latest of thread", WithEmailLatestOfThread(true), "latest_of_thread", "true"},
	}

	s.Require().Len(tests, 21, "every documented list query parameter needs an option")

	for _, test := range tests {
		s.Run(test.name, func() {
			values := newEmailListQuery(test.option)

			s.Require().Len(values, 1, "an option must render exactly one query parameter")
			s.Equal(test.value, values.Get(test.key))
		})
	}
}

// TestEmailListOptionsEnums verifies the enum options serialize to the exact
// values the API documents.
func (s *InstantlyTestSuite) TestEmailListOptionsEnums() {
	tests := []struct {
		name   string
		option EmailListOption
		key    string
		value  string
	}{
		{"focused inbox", WithEmailMode(EmailModeFocused), "mode", "emode_focused"},
		{"other inboxes", WithEmailMode(EmailModeOthers), "mode", "emode_others"},
		{"every inbox", WithEmailMode(EmailModeAll), "mode", "emode_all"},
		{"ascending", WithEmailSortOrder(SortOrderAsc), "sort_order", "asc"},
		{"descending", WithEmailSortOrder(SortOrderDesc), "sort_order", "desc"},
		{"received", WithEmailType(EmailTypeReceived), "email_type", "received"},
		{"sent", WithEmailType(EmailTypeSent), "email_type", "sent"},
		{"manual", WithEmailType(EmailTypeManual), "email_type", "manual"},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.Equal(test.value, newEmailListQuery(test.option).Get(test.key))
		})
	}
}

// TestEmailListOptionsUnsetSendNothing verifies an option that is never supplied
// sends no query parameter at all, rather than sending an empty one.
func (s *InstantlyTestSuite) TestEmailListOptionsUnsetSendNothing() {
	s.Nil(newEmailListQuery(), "no options must render no query parameters")
	s.Nil(newEmailListQuery(nil), "a nil option must be ignored rather than panic")

	values := newEmailListQuery(WithEmailLimit(10))

	s.Require().Len(values, 1)
	s.Empty(values.Get("search"), "an unsupplied option must not appear in the query")
	s.Empty(values.Get("is_unread"))
	s.Empty(values.Get("mode"))
}

// TestEmailListOptionsCombined verifies several options render together, and
// that the last value supplied for a parameter is the one sent.
func (s *InstantlyTestSuite) TestEmailListOptionsCombined() {
	values := newEmailListQuery(
		WithEmailLimit(25),
		WithEmailIsUnread(true),
		WithEmailMode(EmailModeAll),
		WithEmailSortOrder(SortOrderDesc),
		WithEmailLimit(100),
	)

	s.Require().Len(values, 4)
	s.Equal("100", values.Get("limit"), "the last value supplied wins")
	s.Equal("true", values.Get("is_unread"))
	s.Equal("emode_all", values.Get("mode"))
	s.Equal("desc", values.Get("sort_order"))
}

// TestEmailListOptionsReachTheAPI verifies the rendered options survive the trip
// to the server as an encoded query string.
func (s *InstantlyTestSuite) TestEmailListOptionsReachTheAPI() {
	s.mux.Get(testPath, func(w http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()
		s.Equal("25", query.Get("limit"))
		s.Equal("quick question", query.Get("search"))
		s.Equal(testEAccount, query.Get("eaccount"))
		s.Equal("emode_focused", query.Get("mode"))
		s.Empty(query.Get("company_domain"))

		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	_, err := s.client.ListEmails(
		context.Background(),
		WithEmailLimit(25),
		WithEmailSearch("quick question"),
		WithEmailAccount(testEAccount),
		WithEmailMode(EmailModeFocused),
	)

	s.Require().NoError(err)
}

// assertAPIError asserts an error carries the documented 4xx envelope.
func assertAPIError(s *InstantlyTestSuite, err error, statusCode int) {
	s.T().Helper()

	s.Require().Error(err)

	var apiErr *APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(int64(statusCode), apiErr.StatusCode)
	s.NotEmpty(apiErr.Code)
}

// writeAPIErrorEnvelope writes the documented 4xx error envelope.
func writeAPIErrorEnvelope(w http.ResponseWriter, statusCode int, code, message string) {
	w.WriteHeader(statusCode)
	_, _ = fmt.Fprintf(w, `{"statusCode":%d,"error":%q,"message":%q}`, statusCode, code, message)
}

// readAll drains a request body so a handler can assert on its contents.
func readAll(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}

	return io.ReadAll(req.Body)
}

// ptrTo returns a pointer to v, for the request fields the API models as
// optional.
func ptrTo[T any](v T) *T {
	return &v
}
