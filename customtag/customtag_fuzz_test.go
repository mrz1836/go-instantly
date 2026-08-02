package customtag_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/customtag"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// FuzzCustomTagSerialization round trips arbitrary field values through the
// create, update, and toggle bodies, asserting the encoding never panics and
// never drifts.
func FuzzCustomTagSerialization(f *testing.F) {
	f.Add("VIP", "High-value leads", "tag-1", 1)
	f.Add("", "", "", 0)
	f.Add("Ünïcödé", "line\r\n", "\x00id", 2)

	f.Fuzz(func(t *testing.T, label, description, id string, kind int) {
		lossless := utf8.ValidString(label) && utf8.ValidString(description) && utf8.ValidString(id)

		instantlytest.RequireStableRoundTrip(t, customtag.CreateRequest{
			Label:       label,
			Description: &description,
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, customtag.UpdateRequest{
			Label:       label,
			Description: instantly.Ptr(description),
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, customtag.ToggleRequest{
			TagIDs:       []string{id},
			ResourceType: customtag.ResourceType(kind%2 + 1),
			Assign:       kind%2 == 0,
			ResourceIDs:  []string{label},
			SelectedAll:  instantly.Ptr(kind%2 == 1),
			Search:       description,
			Filter:       json.RawMessage(`"ACC_FILTER_PAUSED"`),
		}, lossless)
	})
}

// FuzzCustomTagResponseDecoding feeds arbitrary bytes back as custom-tag
// responses, asserting the client never panics and never hands back a partly
// decoded value.
func FuzzCustomTagResponseDecoding(f *testing.F) {
	f.Add(tagFixture)
	f.Add(tagFixtureNulls)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{"success":true}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"description":123}`)
	f.Add(`{"success":"nope"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := customtag.New(instantlytest.FuzzClient(http.StatusOK, body))

		got, err := svc.Get(ctx, tagID)
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated tag")
		} else {
			require.NotNil(t, got)
		}

		page, err := svc.List(ctx)
		if err != nil {
			require.Nil(t, page, "a decode failure must never hand back a partly populated page")
		} else {
			require.NotNil(t, page)
		}

		result, err := svc.Toggle(ctx, customtag.ToggleRequest{})
		if err != nil {
			require.Nil(t, result, "a decode failure must never hand back a partly populated result")
		} else {
			require.NotNil(t, result)
		}

		require.NotPanics(t, func() {
			_, _ = svc.Create(ctx, customtag.CreateRequest{})
			_, _ = svc.Delete(ctx, tagID)
		})
	})
}
