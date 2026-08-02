package subsequence_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/subsequence"
)

// FuzzSubsequenceSerialization round trips arbitrary field values through the
// create and update bodies, asserting the encoding never panics and never drifts.
func FuzzSubsequenceSerialization(f *testing.F) {
	f.Add("Follow up", "campaign-1", "custom", 50)
	f.Add("", "", "", 0)
	f.Add("Ünïcödé", "c\r\n", "\x00", -1)

	f.Fuzz(func(t *testing.T, name, campaign, mode string, limit int) {
		lossless := utf8.ValidString(name) && utf8.ValidString(campaign) && utf8.ValidString(mode)

		instantlytest.RequireStableRoundTrip(t, subsequence.CreateRequest{
			ParentCampaign:      campaign,
			Name:                name,
			Conditions:          json.RawMessage(`{"trigger":"no_reply"}`),
			SubsequenceSchedule: json.RawMessage(`{}`),
			Sequences:           json.RawMessage(`[]`),
			DailyLimit:          instantly.Ptr(float64(limit)),
			DailyLimitMode:      subsequence.DailyLimitMode(mode),
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, subsequence.UpdateRequest{
			Name:       name,
			DailyLimit: instantly.Ptr(float64(limit)),
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, subsequence.DuplicateRequest{
			ParentCampaign: campaign,
			Name:           name,
		}, lossless)
	})
}

// FuzzSubsequenceResponseDecoding feeds arbitrary bytes back as subsequence
// responses, asserting the client never panics and never hands back a partly
// decoded value.
func FuzzSubsequenceResponseDecoding(f *testing.F) {
	f.Add(fixture)
	f.Add(fixtureNulls)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`{"status":"not a number"}`)
	f.Add(`{"conditions":42}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := subsequence.New(instantlytest.FuzzClient(http.StatusOK, body))

		got, err := svc.Get(ctx, testID)
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated subsequence")
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
			_, _ = svc.Pause(ctx, testID)
			_, _ = svc.SendingStatus(ctx, testID)
		})
	})
}
