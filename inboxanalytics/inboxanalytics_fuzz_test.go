package inboxanalytics_test

import (
	"context"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/inboxanalytics"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// FuzzInboxAnalyticsSerialization round trips arbitrary field values through the
// three stats request bodies, asserting the encoding never panics and never
// drifts.
func FuzzInboxAnalyticsSerialization(f *testing.F) {
	f.Add("test-uuid-1", "2026-08-01", "2026-08-31", "sender@x.com", 1)
	f.Add("", "", "", "", 0)
	f.Add("Ünïcödé", "date\r\n", "\x00to", "a@b.c", 13)

	f.Fuzz(func(t *testing.T, testID, from, to, sender string, esp int) {
		lossless := utf8.ValidString(testID) && utf8.ValidString(from) &&
			utf8.ValidString(to) && utf8.ValidString(sender)

		instantlytest.RequireStableRoundTrip(t, inboxanalytics.StatsByTestIDRequest{
			TestIDs:       []string{testID},
			DateFrom:      from,
			DateTo:        to,
			SenderEmail:   sender,
			RecipientESP:  []inboxanalytics.ESP{inboxanalytics.ESP(esp)},
			RecipientGeo:  []inboxanalytics.Geo{inboxanalytics.GeoUS},
			RecipientType: []inboxanalytics.RecipientType{inboxanalytics.RecipientPersonal},
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, inboxanalytics.DeliverabilityInsightsRequest{
			TestID:       testID,
			DateFrom:     from,
			DateTo:       to,
			ShowPrevious: instantly.Ptr(esp%2 == 0),
			RecipientESP: []inboxanalytics.ESP{inboxanalytics.ESP(esp)},
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, inboxanalytics.StatsByDateRequest{
			TestID:       testID,
			DateFrom:     from,
			SenderEmail:  sender,
			RecipientGeo: []inboxanalytics.Geo{inboxanalytics.GeoItaly},
		}, lossless)
	})
}

// FuzzInboxAnalyticsResponseDecoding feeds arbitrary bytes back as
// inbox-placement-analytics responses, asserting the client never panics and
// never hands back a partly decoded value.
func FuzzInboxAnalyticsResponseDecoding(f *testing.F) {
	f.Add(analyticsFixture)
	f.Add(analyticsFixtureNulls)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`[{"test_id":"t1","count":1,"spam_count":0}]`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"record_type":"not a number"}`)
	f.Add(`{"recipient_esp":2.5}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := inboxanalytics.New(instantlytest.FuzzClient(http.StatusOK, body))

		got, err := svc.Get(ctx, eventID)
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated event")
		} else {
			require.NotNil(t, got)
		}

		page, err := svc.List(ctx, testID)
		if err != nil {
			require.Nil(t, page, "a decode failure must never hand back a partly populated page")
		} else {
			require.NotNil(t, page)
		}

		require.NotPanics(t, func() {
			_, _ = svc.StatsByTestID(ctx, inboxanalytics.StatsByTestIDRequest{})
			_, _ = svc.DeliverabilityInsights(ctx, inboxanalytics.DeliverabilityInsightsRequest{})
			_, _ = svc.StatsByDate(ctx, inboxanalytics.StatsByDateRequest{})
		})
	})
}
