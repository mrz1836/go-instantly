package inboxtest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/inboxtest"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// FuzzInboxTestSerialization round trips arbitrary field values through the
// create and update bodies, asserting the encoding never panics and never drifts.
func FuzzInboxTestSerialization(f *testing.F) {
	f.Add("My Test", "Subject", "Body", "john@doe.com", 1)
	f.Add("", "", "", "", 0)
	f.Add("Ünïcödé", "Re:\r\n", "\x00body", "a@b.c", 2)

	f.Fuzz(func(t *testing.T, name, subject, body, email string, mode int) {
		lossless := utf8.ValidString(name) && utf8.ValidString(subject) &&
			utf8.ValidString(body) && utf8.ValidString(email)

		instantlytest.RequireStableRoundTrip(t, inboxtest.CreateRequest{
			Name:           name,
			Type:           inboxtest.Type(mode%2 + 1),
			SendingMethod:  inboxtest.SendingMethod(mode%2 + 1),
			EmailSubject:   subject,
			EmailBody:      body,
			Emails:         []string{email},
			DeliveryMode:   instantly.Ptr(inboxtest.DeliveryMode(mode%2 + 1)),
			RunImmediately: instantly.Ptr(mode%2 == 0),
			Schedule:       json.RawMessage(`{"timezone":"Etc/GMT+12"}`),
			RecipientsLabels: []inboxtest.RecipientLabel{
				{ESP: name, Region: subject, SubRegion: body, Type: email},
			},
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, inboxtest.UpdateRequest{
			Name:        name,
			Status:      instantly.Ptr(inboxtest.Status(mode%3 + 1)),
			Automations: json.RawMessage(`[{"when":{"condition":"placement_goes_below"}}]`),
		}, lossless)
	})
}

// FuzzInboxTestResponseDecoding feeds arbitrary bytes back as inbox-placement-test
// responses, asserting the client never panics and never hands back a partly
// decoded value.
func FuzzInboxTestResponseDecoding(f *testing.F) {
	f.Add(testFixture)
	f.Add(testFixtureNulls)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`[{"esp":"Google","region":"NA","sub_region":"US","type":"Professional"}]`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"type":"not a number"}`)
	f.Add(`{"status":2.5}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := inboxtest.New(instantlytest.FuzzClient(http.StatusOK, body))

		got, err := svc.Get(ctx, testID)
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated test")
		} else {
			require.NotNil(t, got)
		}

		page, err := svc.List(ctx)
		if err != nil {
			require.Nil(t, page, "a decode failure must never hand back a partly populated page")
		} else {
			require.NotNil(t, page)
		}

		require.NotPanics(t, func() {
			_, _ = svc.ESPOptions(ctx)
			_, _ = svc.Create(ctx, inboxtest.CreateRequest{Name: "x"})
			_, _ = svc.Delete(ctx, testID)
		})
	})
}
