package account_test

import (
	"context"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/account"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// FuzzAccountSerialization round trips arbitrary field values through the create
// and update bodies, asserting the encoding never panics and never drifts.
func FuzzAccountSerialization(f *testing.F) {
	f.Add("sender@example.com", "Jon", "Doe", "imap.example.com", 993)
	f.Add("", "", "", "", 0)
	f.Add("a@b.c", "Ünïcödé", "Doe\r\n", "\x00host", -1)

	f.Fuzz(func(t *testing.T, email, first, last, host string, port int) {
		lossless := utf8.ValidString(email) && utf8.ValidString(first) &&
			utf8.ValidString(last) && utf8.ValidString(host)

		instantlytest.RequireStableRoundTrip(t, account.CreateRequest{
			Email:        email,
			FirstName:    first,
			LastName:     last,
			ProviderCode: account.ProviderCode(port % 12),
			IMAPHost:     host,
			IMAPPort:     int64(port),
			Warmup:       &account.Warmup{Increment: account.WarmupIncrement(first)},
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, account.UpdateRequest{
			FirstName:  first,
			DailyLimit: instantly.Ptr(float64(port)),
			Signature:  &last,
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, account.MoveRequest{
			Emails:                 []string{email},
			SourceWorkspaceID:      first,
			DestinationWorkspaceID: last,
		}, lossless)
	})
}

// FuzzAccountResponseDecoding feeds arbitrary bytes back as account responses,
// asserting the client never panics and never hands back a partly decoded value.
func FuzzAccountResponseDecoding(f *testing.F) {
	f.Add(accountFixture)
	f.Add(accountFixtureNulls)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`{"status":"not a number"}`)
	f.Add(`{"provider_code":2.5}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := account.New(instantlytest.FuzzClient(http.StatusOK, body))

		got, err := svc.Get(ctx, testEmail)
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated account")
		} else {
			require.NotNil(t, got)
		}

		page, err := svc.List(ctx)
		if err != nil {
			require.Nil(t, page)
		} else {
			require.NotNil(t, page)
		}

		require.NotPanics(t, func() {
			_, _ = svc.CtdStatus(ctx, "x")
			_, _ = svc.DailyAnalytics(ctx)
			_, _ = svc.PauseBulk(ctx, account.PauseBulkRequest{})
		})
	})
}
