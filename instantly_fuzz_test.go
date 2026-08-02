package instantly

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
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

	// The statuses the API documents for every email operation. Fuzzed statuses
	// are folded onto them so the request path is exercised rather than the
	// redirect handling of an arbitrary status.
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
	err := client.get(context.Background(), "/api/v2/emails", &destination)

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		require.Empty(t, destination, "an API error must never populate the destination")
	}
}

// fuzzClient returns a client whose transport answers every request with the
// given status and body, so no fuzz input ever reaches a network.
func fuzzClient(statusCode int, body string) *Client {
	client := NewClient(testAPIKey)
	client.HTTPClient = &http.Client{Transport: roundTripFunc(
		func(_ *http.Request) (*http.Response, error) {
			return jsonResponse(statusCode, body), nil
		},
	)}

	return client
}
