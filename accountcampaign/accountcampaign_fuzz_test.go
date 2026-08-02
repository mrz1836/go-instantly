package accountcampaign_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly/accountcampaign"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// FuzzAccountCampaignResponseDecoding feeds arbitrary bytes back as mapping
// responses, asserting the client never panics and never hands back a partly
// decoded value. The endpoint is read-only, so there is no request body to fuzz.
func FuzzAccountCampaignResponseDecoding(f *testing.F) {
	f.Add(`{"items":[{"campaign_id":"c1","campaign_name":"C","status":1,` +
		`"timestamp_created":"2026-08-01T10:00:00.000Z"}],"next_starting_after":"cursor-2"}`)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`{"status":"not a number"}`)

	f.Fuzz(func(t *testing.T, body string) {
		svc := accountcampaign.New(instantlytest.FuzzClient(http.StatusOK, body))

		page, err := svc.List(context.Background(), testEmail)
		if err != nil {
			require.Nil(t, page, "a decode failure must never hand back a partly populated page")
		} else {
			require.NotNil(t, page)
		}
	})
}
