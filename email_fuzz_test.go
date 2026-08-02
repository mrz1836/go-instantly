package instantly

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

// fuzzEmailRequest carries the fuzzed field values for the email request types.
// It is declared with explicit JSON tags so it round trips under the same rules
// as the exported request bodies.
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

	f.Fuzz(func(t *testing.T, eaccount, recipients, subject, html string) {
		fuzzed := fuzzEmailRequest{
			EAccount:           eaccount,
			ToAddressEmailList: recipients,
			Subject:            subject,
			HTML:               html,
		}

		// Encoding coerces invalid UTF-8 to the replacement character, so exact
		// equality is only asserted for input the encoder can represent.
		lossless := utf8.ValidString(eaccount) && utf8.ValidString(recipients) &&
			utf8.ValidString(subject) && utf8.ValidString(html)

		requireStableRoundTrip(t, sendTestRequest(fuzzed), lossless)
		requireStableRoundTrip(t, replyRequest(fuzzed), lossless)
		requireStableRoundTrip(t, forwardRequest(fuzzed), lossless)
		requireStableRoundTrip(t, UpdateEmailRequest{IsUnread: ptrTo(true), ReminderTS: &subject}, lossless)
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
		client := fuzzClient(http.StatusOK, body)

		email, err := client.GetEmail(ctx, testEmailID)
		if err != nil {
			require.Nil(t, email, "a decode failure must never hand back a partly populated email")
		} else {
			require.NotNil(t, email)
		}

		page, err := client.ListEmails(ctx)
		if err != nil {
			require.Nil(t, page, "a decode failure must never hand back a partly populated page")
		} else {
			require.NotNil(t, page)
		}

		count, err := client.CountUnreadEmails(ctx)
		if err != nil {
			require.Zero(t, count, "a failed count must not report a number the API never sent")
		}

		// The write endpoints decode the same bodies and must survive them too.
		require.NotPanics(t, func() {
			_ = client.SendTestEmail(ctx, SendTestEmailRequest{EAccount: testEAccount})
			_ = client.MarkThreadAsRead(ctx, testThreadID)
		})
	})
}

// requireStableRoundTrip asserts a request body survives encoding and decoding
// unchanged, and that a second encoding produces the same bytes.
func requireStableRoundTrip[T any](t *testing.T, request T, lossless bool) {
	t.Helper()

	encoded, err := json.Marshal(request)
	require.NoError(t, err)

	var decoded T
	require.NoError(t, json.Unmarshal(encoded, &decoded))

	reencoded, err := json.Marshal(decoded)
	require.NoError(t, err)
	require.JSONEq(t, string(encoded), string(reencoded), "a decoded body must re-encode identically")

	if lossless {
		require.Equal(t, request, decoded)
	}
}

// sendTestRequest builds a send-test-email body from the fuzzed values.
func sendTestRequest(fuzzed fuzzEmailRequest) SendTestEmailRequest {
	return SendTestEmailRequest{
		EAccount:           fuzzed.EAccount,
		ToAddressEmailList: fuzzed.ToAddressEmailList,
		Subject:            fuzzed.Subject,
		Body:               EmailBody{HTML: fuzzed.HTML},
	}
}

// replyRequest builds a reply body from the fuzzed values.
func replyRequest(fuzzed fuzzEmailRequest) ReplyToEmailRequest {
	return ReplyToEmailRequest{
		ReplyToUUID:          fuzzed.EAccount,
		EAccount:             fuzzed.EAccount,
		Subject:              fuzzed.Subject,
		Body:                 EmailBody{HTML: fuzzed.HTML},
		AdditionalRecipients: []string{fuzzed.ToAddressEmailList},
		CCAddressEmailList:   fuzzed.ToAddressEmailList,
	}
}

// forwardRequest builds a forward body from the fuzzed values.
func forwardRequest(fuzzed fuzzEmailRequest) ForwardEmailRequest {
	return ForwardEmailRequest{
		ReplyToUUID:        fuzzed.EAccount,
		ToAddressEmailList: fuzzed.ToAddressEmailList,
		EAccount:           fuzzed.EAccount,
		Subject:            fuzzed.Subject,
		Body:               &EmailBody{HTML: fuzzed.HTML},
	}
}
