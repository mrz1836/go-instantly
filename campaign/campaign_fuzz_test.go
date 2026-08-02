package campaign_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/campaign"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// FuzzCampaignSerialization round trips arbitrary field values through the
// create and update bodies, asserting the encoding never panics and never drifts.
func FuzzCampaignSerialization(f *testing.F) {
	f.Add("Launch", "2026-08-01", "sender@example.com", 100)
	f.Add("", "", "", 0)
	f.Add("Ünïcödé ✉", "date\r\n", "a@b.c\x00", -5)

	f.Fuzz(func(t *testing.T, name, startDate, sender string, limit int) {
		lossless := utf8.ValidString(name) && utf8.ValidString(startDate) && utf8.ValidString(sender)

		instantlytest.RequireStableRoundTrip(t, campaign.CreateRequest{
			Name: name,
			CampaignSchedule: campaign.Schedule{
				StartDate: &startDate,
				Schedules: []campaign.ScheduleItem{{Name: name, Timezone: "UTC"}},
			},
			EmailList:  []string{sender},
			DailyLimit: instantly.Ptr(float64(limit)),
			Sequences:  json.RawMessage(`[{"steps":[]}]`),
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, campaign.UpdateRequest{
			Name:        name,
			StopOnReply: instantly.Ptr(true),
			BCCList:     []string{sender},
		}, lossless)
	})
}

// FuzzCampaignResponseDecoding feeds arbitrary bytes back as campaign responses,
// asserting the client never panics and never hands back a partly decoded value.
func FuzzCampaignResponseDecoding(f *testing.F) {
	f.Add(campaignFixture)
	f.Add(campaignFixtureNulls)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`{"status":"not a number"}`)
	f.Add(`{"campaign_schedule":42}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := campaign.New(instantlytest.FuzzClient(http.StatusOK, body))

		got, err := svc.Get(ctx, testID)
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated campaign")
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
			_, _ = svc.CountLaunched(ctx)
			_, _ = svc.Analytics(ctx)
			_, _ = svc.SendingStatus(ctx, testID)
			_ = svc.Share(ctx, testID)
		})
	})
}
