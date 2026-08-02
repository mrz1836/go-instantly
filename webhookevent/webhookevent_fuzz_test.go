package webhookevent_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/webhookevent"
)

// FuzzWebhookEventResponseDecoding feeds arbitrary bytes back as webhook-event
// responses, asserting the client never panics and never hands back a partly
// decoded value.
//
// The events are read-only, so there is no request-body serialization to fuzz.
func FuzzWebhookEventResponseDecoding(f *testing.F) {
	f.Add(eventFixture)
	f.Add(eventFixtureNulls)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{"total_events":100,"successful_events":90,"failed_events":10,"success_rate":0.9,"failure_rate":0.1}`)
	f.Add(`{"items":[{"date":"2026-08-01","total_events":1,"success_rate":1}]}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"retry_count":"not a number"}`)
	f.Add(`{"status_code":"nope"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := webhookevent.New(instantlytest.FuzzClient(http.StatusOK, body))

		got, err := svc.Get(ctx, eventID)
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated event")
		} else {
			require.NotNil(t, got)
		}

		page, err := svc.List(ctx)
		if err != nil {
			require.Nil(t, page, "a decode failure must never hand back a partly populated page")
		} else {
			require.NotNil(t, page)
		}

		summary, err := svc.Summary(ctx, "", "")
		if err != nil {
			require.Nil(t, summary, "a decode failure must never hand back a partly populated summary")
		} else {
			require.NotNil(t, summary)
		}

		require.NotPanics(t, func() {
			_, _ = svc.SummaryByDate(ctx, "", "")
		})
	})
}
