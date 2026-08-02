package blocklist_test

import (
	"context"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly/blocklist"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// FuzzBlockListSerialization round trips arbitrary field values through the
// request bodies, asserting the encoding never panics and never drifts.
func FuzzBlockListSerialization(f *testing.F) {
	f.Add("spam.example.com", "bl-1")
	f.Add("", "")
	f.Add("Ünïcödé\r\n", "\x00id")

	f.Fuzz(func(t *testing.T, value, id string) {
		lossless := utf8.ValidString(value) && utf8.ValidString(id)

		instantlytest.RequireStableRoundTrip(t, blocklist.CreateRequest{BLValue: value}, lossless)
		instantlytest.RequireStableRoundTrip(t, blocklist.UpdateRequest{BLValue: value}, lossless)
		instantlytest.RequireStableRoundTrip(t, blocklist.BulkCreateRequest{BLValues: []string{value}}, lossless)
		instantlytest.RequireStableRoundTrip(t, blocklist.BulkDeleteRequest{IDs: []string{id}}, lossless)
	})
}

// FuzzBlockListResponseDecoding feeds arbitrary bytes back as block-list
// responses, asserting the client never panics and never hands back a partly
// decoded value.
func FuzzBlockListResponseDecoding(f *testing.F) {
	f.Add(entryFixture)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{"items":[],"valid_count":0,"invalid_count":0}`)
	f.Add(`[]`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`{"is_domain":"not a bool"}`)
	f.Add(`{"valid_count":"nope"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := blocklist.New(instantlytest.FuzzClient(http.StatusOK, body))

		got, err := svc.Get(ctx, entryID)
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated entry")
		} else {
			require.NotNil(t, got)
		}

		page, err := svc.List(ctx)
		if err != nil {
			require.Nil(t, page, "a decode failure must never hand back a partly populated page")
		} else {
			require.NotNil(t, page)
		}

		bulk, err := svc.BulkCreate(ctx, blocklist.BulkCreateRequest{})
		if err != nil {
			require.Nil(t, bulk, "a decode failure must never hand back a partly populated result")
		} else {
			require.NotNil(t, bulk)
		}

		require.NotPanics(t, func() {
			_, _ = svc.DeleteAll(ctx, false, "")
			_, _ = svc.BulkDelete(ctx, blocklist.BulkDeleteRequest{})
			_, _ = svc.Download(ctx, false, "")
		})
	})
}
