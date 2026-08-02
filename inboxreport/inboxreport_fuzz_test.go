package inboxreport_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly/inboxreport"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// FuzzInboxReportResponseDecoding feeds arbitrary bytes back as
// inbox-placement-report responses, asserting the client never panics and never
// hands back a partly decoded value.
//
// The reports are read-only, so there is no request-body serialization to fuzz.
func FuzzInboxReportResponseDecoding(f *testing.F) {
	f.Add(reportFixture)
	f.Add(reportFixtureNulls)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{"spam_assassin_report":{"report":[{"score":"0.0"}]}}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"spam_assassin_score":"not a number"}`)
	f.Add(`{"domain_blacklist_count":"nope"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := inboxreport.New(instantlytest.FuzzClient(http.StatusOK, body))

		got, err := svc.Get(ctx, reportID)
		if err != nil {
			require.Nil(t, got, "a decode failure must never hand back a partly populated report")
		} else {
			require.NotNil(t, got)
		}

		page, err := svc.List(ctx, testID)
		if err != nil {
			require.Nil(t, page, "a decode failure must never hand back a partly populated page")
		} else {
			require.NotNil(t, page)
		}
	})
}
