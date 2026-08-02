package lead_test

import (
	"context"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/lead"
)

// fuzzClient answers every request with the given status and body, so no fuzz
// input ever reaches a network.
func fuzzClient(statusCode int, body string) *instantly.Client {
	return instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(_ *http.Request) (*http.Response, error) {
				return instantlytest.JSONResponse(statusCode, body), nil
			},
		)},
	))
}

// FuzzLeadSerialization round trips arbitrary field values through the create,
// update, and list bodies, asserting the encoding never panics and never drifts.
func FuzzLeadSerialization(f *testing.F) {
	f.Add("lead@example.com", "Jane", "example.com", 1)
	f.Add("", "", "", 0)
	f.Add("Ünïcödé", "name\r\n", "\x00.com", -3)

	f.Fuzz(func(t *testing.T, email, first, domain string, status int) {
		lossless := utf8.ValidString(email) && utf8.ValidString(first) && utf8.ValidString(domain)

		instantlytest.RequireStableRoundTrip(t, lead.CreateRequest{
			Email:            &email,
			FirstName:        &first,
			LtInterestStatus: instantly.Ptr(lead.InterestStatus(status)),
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, lead.UpdateRequest{
			FirstName: &first,
			Website:   &domain,
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, lead.ListRequest{
			Campaign: domain,
			Search:   email,
			Limit:    status,
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, lead.UpdateInterestStatusRequest{
			LeadEmail:     email,
			InterestValue: instantly.Ptr(lead.InterestStatus(status)),
		}, lossless)
	})
}

// FuzzLeadResponseDecoding feeds arbitrary bytes back as lead responses,
// asserting the client never panics and never hands back a partly decoded value.
func FuzzLeadResponseDecoding(f *testing.F) {
	f.Add(leadFixture)
	f.Add(leadFixtureNulls)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`{"status":"not a number"}`)
	f.Add(`{"email_open_count":"nope"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := lead.New(fuzzClient(http.StatusOK, body))

		got, err := svc.Get(ctx, testID)
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated lead")
		} else {
			require.NotNil(t, got)
		}

		page, err := svc.List(ctx, lead.ListRequest{})
		if err != nil {
			require.Nil(t, page)
		} else {
			require.NotNil(t, page)
		}

		require.NotPanics(t, func() {
			_, _ = svc.BulkDelete(ctx, lead.BulkDeleteRequest{})
			_, _ = svc.Move(ctx, lead.MoveRequest{})
			_ = svc.BulkAssign(ctx, lead.BulkAssignRequest{})
		})
	})
}
