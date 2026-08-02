package leadlist_test

import (
	"context"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/leadlist"
)

// FuzzLeadListSerialization round trips arbitrary field values through the create
// and update bodies, asserting the encoding never panics and never drifts.
func FuzzLeadListSerialization(f *testing.F) {
	f.Add("Prospects", "owner-1")
	f.Add("", "")
	f.Add("Ünïcödé", "o\r\n")
	f.Add("name\x00", "\x00owner")

	f.Fuzz(func(t *testing.T, name, owner string) {
		lossless := utf8.ValidString(name) && utf8.ValidString(owner)

		instantlytest.RequireStableRoundTrip(t, leadlist.CreateRequest{
			Name:              name,
			HasEnrichmentTask: instantly.Ptr(true),
			OwnedBy:           &owner,
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, leadlist.UpdateRequest{
			Name:    name,
			OwnedBy: &owner,
		}, lossless)
	})
}

// FuzzLeadListResponseDecoding feeds arbitrary bytes back as lead-list responses,
// asserting the client never panics and never hands back a partly decoded value.
func FuzzLeadListResponseDecoding(f *testing.F) {
	f.Add(listFixture)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`{"total_leads":"not a number"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := leadlist.New(instantlytest.FuzzClient(http.StatusOK, body))

		got, err := svc.Get(ctx, testID)
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated list")
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
			_, _ = svc.VerificationStats(ctx, testID)
		})
	})
}
