package oauth_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/oauth"
)

// FuzzOAuthResponseDecoding feeds arbitrary bytes back as OAuth responses,
// asserting the client never panics and never hands back a partly decoded value.
//
// The init endpoints send a nil body and the status endpoint is a GET, so there
// is no request-body serialization to fuzz. A body carrying a top-level error is
// converted to an *instantly.APIError, so the result is nil in that case too.
func FuzzOAuthResponseDecoding(f *testing.F) {
	f.Add(initFixture)
	f.Add(`{"status":"success","email":"user@example.com"}`)
	f.Add(`{"status":"pending"}`)
	f.Add(`{"status":"error","error":"access_denied"}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"status":123}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := oauth.New(instantlytest.FuzzClient(http.StatusOK, body))

		google, err := svc.InitGoogle(ctx)
		if err != nil {
			require.Nil(t, google, "a decode failure must never hand back a partly populated result")
		} else {
			require.NotNil(t, google)
		}

		microsoft, err := svc.InitMicrosoft(ctx)
		if err != nil {
			require.Nil(t, microsoft, "a decode failure must never hand back a partly populated result")
		} else {
			require.NotNil(t, microsoft)
		}

		status, err := svc.SessionStatus(ctx, "sess-1")
		if err != nil {
			require.Nil(t, status, "a decode failure must never hand back a partly populated status")
		} else {
			require.NotNil(t, status)
		}
	})
}
