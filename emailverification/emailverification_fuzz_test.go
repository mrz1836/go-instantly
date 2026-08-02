package emailverification_test

import (
	"context"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly/emailverification"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// FuzzEmailVerificationSerialization round trips arbitrary field values through
// the create body, asserting the encoding never panics and never drifts.
func FuzzEmailVerificationSerialization(f *testing.F) {
	f.Add("example@example.com", "https://example.com/webhook")
	f.Add("", "")
	f.Add("a@b.c", "https://héllo.example/ünïcödé")
	f.Add("line\r\nbreak@x.com", "\x00webhook")

	f.Fuzz(func(t *testing.T, email, webhook string) {
		lossless := utf8.ValidString(email) && utf8.ValidString(webhook)

		instantlytest.RequireStableRoundTrip(t, emailverification.CreateRequest{
			Email:      email,
			WebhookURL: webhook,
		}, lossless)
	})
}

// FuzzEmailVerificationResponseDecoding feeds arbitrary bytes back as a
// verification response, asserting the client never panics and never hands back a
// value it could not fully decode.
func FuzzEmailVerificationResponseDecoding(f *testing.F) {
	f.Add(verifiedFixture)
	f.Add(pendingFixture)
	f.Add(`{"email":"x@y.z","verification_status":"invalid","catch_all":false}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"credits":"not a number"}`)
	f.Add(`{"catch_all":123}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := emailverification.New(instantlytest.FuzzClient(http.StatusOK, body))

		got, err := svc.Check(ctx, "x@y.z")
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated verification")
		} else {
			require.NotNil(t, got)
		}

		created, err := svc.Create(ctx, emailverification.CreateRequest{Email: "x@y.z"})
		if err != nil {
			require.Nil(t, created, "a decode failure must never hand back a partly populated verification")
		} else {
			require.NotNil(t, created)
		}
	})
}
