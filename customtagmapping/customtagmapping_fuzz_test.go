package customtagmapping_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly/customtagmapping"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// FuzzCustomTagMappingResponseDecoding feeds arbitrary bytes back as
// custom-tag-mapping responses, asserting the client never panics and never hands
// back a partly decoded value.
//
// The mappings are read-only, so there is no request-body serialization to fuzz.
func FuzzCustomTagMappingResponseDecoding(f *testing.F) {
	f.Add(mappingFixture)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{"items":[{"id":"m1"}],"next_starting_after":"cursor-2"}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"resource_type":"not a number"}`)
	f.Add(`{"resource_type":2.5}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := customtagmapping.New(instantlytest.FuzzClient(http.StatusOK, body))

		page, err := svc.List(ctx)
		if err != nil {
			require.Nil(t, page, "a decode failure must never hand back a partly populated page")
		} else {
			require.NotNil(t, page)
		}
	})
}
