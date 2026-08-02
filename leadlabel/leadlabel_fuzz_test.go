package leadlabel_test

import (
	"context"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/leadlabel"
)

// FuzzLeadLabelSerialization round trips arbitrary field values through the
// create, update, and AI-reply bodies, asserting the encoding never panics and
// never drifts.
func FuzzLeadLabelSerialization(f *testing.F) {
	f.Add("Interested", "positive", "a warm lead")
	f.Add("", "", "")
	f.Add("Ünïcödé", "p\r\n", "\x00")

	f.Fuzz(func(t *testing.T, label, status, desc string) {
		lossless := utf8.ValidString(label) && utf8.ValidString(status) && utf8.ValidString(desc)

		instantlytest.RequireStableRoundTrip(t, leadlabel.CreateRequest{
			Label:               label,
			InterestStatusLabel: status,
			Description:         &desc,
			UseWithAI:           instantly.Ptr(true),
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, leadlabel.UpdateRequest{
			Label:               label,
			InterestStatusLabel: status,
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, leadlabel.AIReplyLabelRequest{
			ReplyText: desc,
		}, lossless)
	})
}

// FuzzLeadLabelResponseDecoding feeds arbitrary bytes back as lead-label
// responses, asserting the client never panics and never hands back a partly
// decoded value.
func FuzzLeadLabelResponseDecoding(f *testing.F) {
	f.Add(labelFixture)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`{"interest_status":"not a number"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := leadlabel.New(instantlytest.FuzzClient(http.StatusOK, body))

		got, err := svc.Get(ctx, testID)
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated label")
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
			_, _ = svc.TestAIReplyLabel(ctx, leadlabel.AIReplyLabelRequest{})
		})
	})
}
