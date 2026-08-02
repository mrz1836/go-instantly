package email_test

import (
	"context"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/email"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// fuzzEmailRequest carries the fuzzed field values for the email request types.
type fuzzEmailRequest struct {
	EAccount           string `json:"eaccount"`
	ToAddressEmailList string `json:"to_address_email_list"`
	Subject            string `json:"subject"`
	HTML               string `json:"html"`
}

// FuzzEmailSerialization round trips arbitrary field values through every email
// request body, asserting the encoding never panics and never drifts.
func FuzzEmailSerialization(f *testing.F) {
	f.Add("sender@example.com", "lead@example.com", "Quick question", "<p>Hello</p>")
	f.Add("", "", "", "")
	f.Add("a@b.c", "x@y.z,q@r.s", "Re: Fwd: Re:", `{"not":"html"}`)
	f.Add("sender@example.com", "lead@example.com", "Ünïcödé ✉", "<p>Héllo</p>")
	f.Add("sender@example.com", "lead@example.com", "line\r\nbreak", "<p>\x00</p>")

	f.Fuzz(func(t *testing.T, eaccount, recipients, subj, html string) {
		fuzzed := fuzzEmailRequest{
			EAccount:           eaccount,
			ToAddressEmailList: recipients,
			Subject:            subj,
			HTML:               html,
		}

		// Encoding coerces invalid UTF-8 to the replacement character, so exact
		// equality is only asserted for input the encoder can represent.
		lossless := utf8.ValidString(eaccount) && utf8.ValidString(recipients) &&
			utf8.ValidString(subj) && utf8.ValidString(html)

		instantlytest.RequireStableRoundTrip(t, sendTestRequest(fuzzed), lossless)
		instantlytest.RequireStableRoundTrip(t, replyRequest(fuzzed), lossless)
		instantlytest.RequireStableRoundTrip(t, forwardRequest(fuzzed), lossless)
		instantlytest.RequireStableRoundTrip(
			t, email.UpdateRequest{IsUnread: instantly.Ptr(true), ReminderTS: &subj}, lossless,
		)
	})
}

// FuzzEmailResponseDecoding feeds arbitrary bytes back as an email response,
// asserting the client never panics and never hands back a value it could not
// fully decode.
func FuzzEmailResponseDecoding(f *testing.F) {
	f.Add(emailFixture)
	f.Add(emailFixtureNulls)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{"items":[{"id":"e1"}],"next_starting_after":"cursor-2"}`)
	f.Add(`{"count":42}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"id":123}`)
	f.Add(`{"is_unread":"not a number"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := email.New(instantlytest.FuzzClient(http.StatusOK, body))

		got, err := svc.Get(ctx, emailID)
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated email")
		} else {
			require.NotNil(t, got)
		}

		page, err := svc.List(ctx)
		if err != nil {
			require.Nil(t, page, "a decode failure must never hand back a partly populated page")
		} else {
			require.NotNil(t, page)
		}

		count, err := svc.CountUnread(ctx)
		if err != nil {
			require.Zero(t, count, "a failed count must not report a number the API never sent")
		}

		// The write endpoints decode the same bodies and must survive them too.
		require.NotPanics(t, func() {
			_ = svc.SendTest(ctx, email.SendTestRequest{EAccount: eAccount})
			_ = svc.MarkThreadAsRead(ctx, threadID)
		})
	})
}

// sendTestRequest builds a send-test-email body from the fuzzed values.
func sendTestRequest(fuzzed fuzzEmailRequest) email.SendTestRequest {
	return email.SendTestRequest{
		EAccount:           fuzzed.EAccount,
		ToAddressEmailList: fuzzed.ToAddressEmailList,
		Subject:            fuzzed.Subject,
		Body:               email.Body{HTML: fuzzed.HTML},
	}
}

// replyRequest builds a reply body from the fuzzed values.
func replyRequest(fuzzed fuzzEmailRequest) email.ReplyRequest {
	return email.ReplyRequest{
		ReplyToUUID:          fuzzed.EAccount,
		EAccount:             fuzzed.EAccount,
		Subject:              fuzzed.Subject,
		Body:                 email.Body{HTML: fuzzed.HTML},
		AdditionalRecipients: []string{fuzzed.ToAddressEmailList},
		CCAddressEmailList:   fuzzed.ToAddressEmailList,
	}
}

// forwardRequest builds a forward body from the fuzzed values.
func forwardRequest(fuzzed fuzzEmailRequest) email.ForwardRequest {
	return email.ForwardRequest{
		ReplyToUUID:        fuzzed.EAccount,
		ToAddressEmailList: fuzzed.ToAddressEmailList,
		EAccount:           fuzzed.EAccount,
		Subject:            fuzzed.Subject,
		Body:               &email.Body{HTML: fuzzed.HTML},
	}
}
