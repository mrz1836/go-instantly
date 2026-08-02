package workspacebilling_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/workspacebilling"
)

// FuzzWorkspaceBillingResponseDecoding feeds arbitrary bytes back as
// workspace-billing responses, asserting the client never panics and never hands
// back a partly decoded value.
//
// The billing details are read-only, so there is no request-body serialization
// to fuzz.
func FuzzWorkspaceBillingResponseDecoding(f *testing.F) {
	f.Add(planFixture)
	f.Add(subscriptionFixture)
	f.Add(`{"all_subs_cancelled":true,"subscriptions":[]}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"all_subs_cancelled":"not a bool"}`)
	f.Add(`{"subscriptions":"not an array"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := workspacebilling.New(instantlytest.FuzzClient(http.StatusOK, body))

		plan, err := svc.PlanDetails(ctx)
		if err != nil {
			require.Nil(t, plan, "a decode failure must never hand back partly populated plan details")
		} else {
			require.NotNil(t, plan)
		}

		subs, err := svc.SubscriptionDetails(ctx)
		if err != nil {
			require.Nil(t, subs, "a decode failure must never hand back partly populated subscription details")
		} else {
			require.NotNil(t, subs)
		}
	})
}
