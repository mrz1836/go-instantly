package backgroundjob_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly/backgroundjob"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// FuzzBackgroundJobResponseDecoding feeds arbitrary bytes back as background-job
// responses, asserting the client never panics and never hands back a partly
// decoded value.
//
// Both endpoints are read-only GETs, so there is no request-body serialization
// to fuzz.
func FuzzBackgroundJobResponseDecoding(f *testing.F) {
	f.Add(jobFixture)
	f.Add(jobFixtureNulls)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{"items":[{"id":"j1"}],"next_starting_after":"cursor-2"}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"progress":"not a number"}`)
	f.Add(`{"data":"not an object"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := backgroundjob.New(instantlytest.FuzzClient(http.StatusOK, body))

		page, err := svc.List(ctx)
		if err != nil {
			require.Nil(t, page, "a decode failure must never hand back a partly populated page")
		} else {
			require.NotNil(t, page)
		}

		job, err := svc.Get(ctx, "job-1")
		if err != nil {
			require.Nil(t, job, "a decode failure must never hand back a partly populated job")
		} else {
			require.NotNil(t, job)
		}
	})
}
