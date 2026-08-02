package campaign_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mrz1836/go-instantly/campaign"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// BenchmarkCampaignDecode measures decoding a fully populated campaign response,
// standing in for the resource decode every read performs.
func BenchmarkCampaignDecode(b *testing.B) {
	body := []byte(campaignFixture)

	for b.Loop() {
		var c campaign.Campaign
		if err := json.Unmarshal(body, &c); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCampaignGet measures a full read: request build, canned transport,
// response check, and decode into the resource model.
func BenchmarkCampaignGet(b *testing.B) {
	svc := campaign.New(instantlytest.FuzzClient(http.StatusOK, campaignFixture))
	ctx := context.Background()

	for b.Loop() {
		if _, err := svc.Get(ctx, testID); err != nil {
			b.Fatal(err)
		}
	}
}
