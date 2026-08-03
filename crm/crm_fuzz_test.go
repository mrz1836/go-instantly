package crm_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly/crm"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// FuzzCRMResponseDecoding feeds arbitrary bytes back as crm-actions responses,
// asserting the client never panics and never hands back a partly decoded value.
//
// The list is a GET and the delete sends no request body, so there is no
// request-body serialization to fuzz.
func FuzzCRMResponseDecoding(f *testing.F) {
	f.Add("[" + numberFixture + "]")
	f.Add(numberFixtureNoPrice)
	f.Add(`[]`)
	f.Add(`[{"id":"num-1","price":null}]`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`{"price":"not a number"}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := crm.New(instantlytest.FuzzClient(http.StatusOK, body))

		numbers, err := svc.ListPhoneNumbers(ctx)
		if err != nil {
			require.Nil(t, numbers, "a decode failure must never hand back a partly populated list")
		}

		number, err := svc.DeletePhoneNumber(ctx, "num-1")
		if err != nil {
			require.Nil(t, number, "a decode failure must never hand back a partly populated number")
		} else {
			require.NotNil(t, number)
		}
	})
}
