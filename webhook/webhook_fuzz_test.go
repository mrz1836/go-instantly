package webhook_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/webhook"
)

// FuzzWebhookSerialization round trips arbitrary field values through the create
// and update bodies, asserting the encoding never panics and never drifts.
func FuzzWebhookSerialization(f *testing.F) {
	f.Add("https://example.com/hook", "My Webhook", "reply_received", "camp-1")
	f.Add("", "", "", "")
	f.Add("https://héllo.example/ünïcödé", "line\r\n", "\x00type", "c")

	f.Fuzz(func(t *testing.T, url, name, eventType, campaign string) {
		lossless := utf8.ValidString(url) && utf8.ValidString(name) &&
			utf8.ValidString(eventType) && utf8.ValidString(campaign)

		instantlytest.RequireStableRoundTrip(t, webhook.CreateRequest{
			TargetHookURL:       url,
			Name:                &name,
			EventType:           webhook.EventType(eventType),
			Campaign:            &campaign,
			CustomInterestValue: instantly.Ptr(1.5),
			Headers:             json.RawMessage(`{"X-Token":"abc"}`),
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, webhook.UpdateRequest{
			TargetHookURL: url,
			EventType:     webhook.EventType(eventType),
			Name:          &name,
		}, lossless)
	})
}

// FuzzWebhookResponseDecoding feeds arbitrary bytes back as webhook responses,
// asserting the client never panics and never hands back a partly decoded value.
func FuzzWebhookResponseDecoding(f *testing.F) {
	f.Add(webhookFixture)
	f.Add(webhookFixtureNulls)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{"event_types":[{"id":"x","label":"X","type":"x"}]}`)
	f.Add(`{"success":true,"status_code":200}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"status":"not a number"}`)
	f.Add(`{"custom_interest_value":"nope"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := webhook.New(instantlytest.FuzzClient(http.StatusOK, body))

		got, err := svc.Get(ctx, webhookID)
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated webhook")
		} else {
			require.NotNil(t, got)
		}

		page, err := svc.List(ctx)
		if err != nil {
			require.Nil(t, page, "a decode failure must never hand back a partly populated page")
		} else {
			require.NotNil(t, page)
		}

		result, err := svc.Test(ctx, webhookID)
		if err != nil {
			require.Nil(t, result, "a decode failure must never hand back a partly populated result")
		} else {
			require.NotNil(t, result)
		}

		require.NotPanics(t, func() {
			_, _ = svc.EventTypes(ctx)
			_, _ = svc.Create(ctx, webhook.CreateRequest{})
			_, _ = svc.Resume(ctx, webhookID)
		})
	})
}
