package instantly

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

// FuzzAPIErrorDecoding feeds arbitrary bytes through both API error wire shapes,
// asserting the two-shape handling never panics and never turns a failure into a
// nil error.
func FuzzAPIErrorDecoding(f *testing.F) {
	f.Add(http.StatusUnauthorized,
		`{"statusCode":401,"error":"Unauthorized","message":"Missing Authorization header"}`, "Unauthorized")
	f.Add(http.StatusTooManyRequests, `{"statusCode":429,"error":"Too Many Requests","message":""}`, "")
	f.Add(http.StatusOK, `{"error":"ACC_AUTH_ERROR"}`, ErrCodeAccountAuthError)
	f.Add(http.StatusOK, `{"status":"success"}`, ErrCodeAccountNotFound)
	f.Add(http.StatusNotFound, `<html><body>404</body></html>`, ErrCodeAccountUnknownError)
	f.Add(http.StatusBadGateway, ``, "")
	f.Add(http.StatusOK, `null`, "  ")
	f.Add(http.StatusOK, `{"error":123}`, "\x00")
	f.Add(http.StatusInternalServerError, `{"error":{"nested":"object"}}`, "nested")

	f.Fuzz(func(t *testing.T, statusCode int, body, code string) {
		var err error
		require.NotPanics(t, func() {
			err = checkResponse(statusCode, []byte(body))
		})

		if statusCode >= http.StatusBadRequest {
			require.Error(t, err, "a failing status must never decode to a nil error")
		}

		var apiErr *APIError
		if errors.As(err, &apiErr) {
			require.NotEmpty(t, apiErr.Error(), "every API error must render a message")

			if statusCode < http.StatusBadRequest {
				require.Equal(t, int64(statusCode), apiErr.StatusCode)
				require.NotEmpty(t, apiErr.Code, "a success body is only an error when it carries a code")
			}
		}

		requireErrorCodeAlwaysFails(t, code)
		requireDestinationUntouched(t, statusCode, body)
	})
}

// FuzzQueryEncoding feeds arbitrary key/value pairs through the query builder,
// asserting the encoded path never trips the standard parser, always round trips
// idempotently, and that Path returns a bare base only when nothing was set.
func FuzzQueryEncoding(f *testing.F) {
	const base = "/api/v2/thing"

	f.Add("limit", "50", "search", "hello world")
	f.Add("", "", "", "")
	f.Add("eaccount", "a@b.com", "mode", "emode_focused")
	f.Add("a", "b\r\nc: injected", "d", "e&f=g")
	f.Add("k", "Ünïcödé", "q", "%40%2F")

	f.Fuzz(func(t *testing.T, k1, v1, k2, v2 string) {
		q := NewQuery()
		q.SetString(k1, v1)
		q.SetString(k2, v2)

		path := q.Path(base)

		if q.Len() == 0 {
			require.Equal(t, base, path, "no parameters returns the bare path")
			return
		}

		require.Equal(t, base+"?"+q.Encode(), path, "Path is base + ? + encoded parameters")

		// Encoding can never produce a query string the standard parser rejects,
		// and re-encoding the parsed result must be byte-for-byte identical.
		encoded := q.Encode()
		parsed, err := url.ParseQuery(encoded)
		require.NoError(t, err)
		require.Equal(t, encoded, parsed.Encode(), "encoding round trips through the parser")

		// The most-recently-set value for each key is the one that survives.
		require.Equal(t, q.Get(k1), parsed.Get(k1))
		require.Equal(t, q.Get(k2), parsed.Get(k2))
	})
}

// requireErrorCodeAlwaysFails asserts a well-formed error code delivered inside
// an HTTP 200 body is always surfaced as an error, and that a success body
// without one never is.
func requireErrorCodeAlwaysFails(t *testing.T, code string) {
	t.Helper()

	encoded, err := json.Marshal(map[string]string{"error": code})
	require.NoError(t, err)

	codeErr := checkResponse(http.StatusOK, encoded)
	if code == "" {
		require.NoError(t, codeErr, "an empty error field is not a failure")
		return
	}

	var apiErr *APIError
	require.ErrorAs(t, codeErr, &apiErr, "a non-empty error field at HTTP 200 must be an error")
	require.NotEmpty(t, apiErr.Code)
	require.Equal(t, int64(http.StatusOK), apiErr.StatusCode)
}

// requireDestinationUntouched drives the fuzzed body through the whole request
// path and asserts an API error never leaves a partly populated destination
// behind, whichever wire shape it arrived in.
func requireDestinationUntouched(t *testing.T, statusCode int, body string) {
	t.Helper()

	// The statuses the API documents for a typical operation. Fuzzed statuses are
	// folded onto them so the request path is exercised rather than the redirect
	// handling of an arbitrary status.
	documented := []int{
		http.StatusOK,
		http.StatusUnauthorized,
		http.StatusPaymentRequired,
		http.StatusNotFound,
		http.StatusTooManyRequests,
	}
	status := documented[((statusCode%len(documented))+len(documented))%len(documented)]

	client := fuzzClient(status, body)

	destination := map[string]any{}
	err := client.Get(context.Background(), "/api/v2/emails", &destination)

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		require.Empty(t, destination, "an API error must never populate the destination")
	}
}

// fuzzClient returns a client whose transport answers every request with the
// given status and body, so no fuzz input ever reaches a network.
func fuzzClient(statusCode int, body string) *Client {
	return NewClient(testAPIKey, WithHTTPClient(&http.Client{Transport: roundTripFunc(
		func(_ *http.Request) (*http.Response, error) {
			return jsonResponse(statusCode, body), nil
		},
	)}))
}
