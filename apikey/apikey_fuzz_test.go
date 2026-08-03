package apikey_test

import (
	"context"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly/apikey"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// FuzzAPIKeySerialization asserts a create request survives a JSON round trip
// unchanged, so no scope or name is dropped or corrupted on the wire.
func FuzzAPIKeySerialization(f *testing.F) {
	f.Add("CI deploy key", "campaigns:read")
	f.Add("", "all:all")
	f.Add("weird\x00name", "not_a_real:scope")
	f.Add("emoji 🚀", "emails:delete")

	f.Fuzz(func(t *testing.T, name, scope string) {
		req := apikey.CreateRequest{
			Name:   name,
			Scopes: []apikey.Scope{apikey.Scope(scope)},
		}

		// Encoding coerces invalid UTF-8 to the replacement character, so exact
		// equality is only asserted for input the encoder can represent.
		lossless := utf8.ValidString(name) && utf8.ValidString(scope)

		instantlytest.RequireStableRoundTrip(t, req, lossless)
	})
}

// FuzzAPIKeyResponseDecoding feeds arbitrary bytes back as api-key responses,
// asserting the client never panics and never hands back a partly decoded value.
func FuzzAPIKeyResponseDecoding(f *testing.F) {
	f.Add(keyFixture)
	f.Add(`{"items":[],"next_starting_after":""}`)
	f.Add(`{"items":[{"id":"k1","scopes":["all:all"]}],"next_starting_after":"cursor-2"}`)
	f.Add(`{"scopes":"not an array"}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := apikey.New(instantlytest.FuzzClient(http.StatusOK, body))

		page, err := svc.List(ctx)
		if err != nil {
			require.Nil(t, page, "a decode failure must never hand back a partly populated page")
		} else {
			require.NotNil(t, page)
		}

		key, err := svc.Delete(ctx, "key-1")
		if err != nil {
			require.Nil(t, key, "a decode failure must never hand back a partly populated key")
		} else {
			require.NotNil(t, key)
		}
	})
}
