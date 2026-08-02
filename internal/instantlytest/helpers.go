package instantlytest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly"
)

// AssertAPIError asserts err carries the documented 4xx envelope with the given
// status code and a non-empty code.
func AssertAPIError(t require.TestingT, err error, statusCode int) {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}

	require.Error(t, err)

	var apiErr *instantly.APIError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, int64(statusCode), apiErr.StatusCode)
	require.NotEmpty(t, apiErr.Code)
}

// WriteAPIErrorEnvelope writes the documented 4xx error envelope to w.
func WriteAPIErrorEnvelope(w http.ResponseWriter, statusCode int, code, message string) {
	w.WriteHeader(statusCode)
	_, _ = fmt.Fprintf(w, `{"statusCode":%d,"error":%q,"message":%q}`, statusCode, code, message)
}

// FailHandler returns a handler that always answers with the documented 4xx
// error envelope for the given status code. It backs the table-driven failure
// tests, so every endpoint's failure path is exercised from one place.
func FailHandler(statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		WriteAPIErrorEnvelope(w, statusCode, http.StatusText(statusCode), "request failed")
	}
}

// FailureCase is one row of a resource suite's table-driven failure test: it
// names the case, the route to fail, and the service call to make. It replaces
// the per-endpoint *Failure clones each suite used to carry.
type FailureCase struct {
	// Name is the subtest name.
	Name string

	// Method is the HTTP method of the route to fail.
	Method string

	// Path is the router pattern of the route to fail.
	Path string

	// Status is the HTTP status the failing route returns, and the status the
	// call under test must surface as an APIError.
	Status int

	// Call performs the service call under test and returns only its error.
	Call func() error
}

// RoundTripFunc adapts a function to http.RoundTripper so transport-level
// details can be asserted without a network round trip.
type RoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the http.RoundTripper interface.
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// JSONResponse builds a canned JSON response for an intercepted transport.
func JSONResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Header:     http.Header{"Content-Type": []string{MediaTypeJSON}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// FuzzClient returns a client whose transport answers every request with the
// given status and body, so no fuzz input ever reaches a network. It is the
// single fixture the resource fuzz tests share in place of a per-package clone.
func FuzzClient(statusCode int, body string) *instantly.Client {
	return instantly.NewClient(APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: RoundTripFunc(
			func(_ *http.Request) (*http.Response, error) {
				return JSONResponse(statusCode, body), nil
			},
		)},
	))
}

// Page wraps already-rendered item JSON in the cursor-paginated list envelope
// every resource shares, appending next_starting_after only when a cursor is
// given. Each resource package keeps just its own per-item fixture and calls
// Page to build the page around it, so the envelope lives in one place.
func Page(items []string, next string) string {
	joined := strings.Join(items, ",")
	if next == "" {
		return fmt.Sprintf(`{"items":[%s]}`, joined)
	}

	return fmt.Sprintf(`{"items":[%s],"next_starting_after":%q}`, joined, next)
}

// ReadAll drains a request body so a handler can assert on its contents.
func ReadAll(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}

	return io.ReadAll(req.Body)
}

// RequireStableRoundTrip asserts a request body survives encoding and decoding
// unchanged, and that a second encoding produces identical bytes.
//
// When lossless is true it also asserts the decoded value equals the original;
// pass false for inputs the JSON encoder cannot represent exactly, such as
// strings containing invalid UTF-8.
func RequireStableRoundTrip[T any](t require.TestingT, request T, lossless bool) {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}

	encoded, err := json.Marshal(request)
	require.NoError(t, err)

	var decoded T
	require.NoError(t, json.Unmarshal(encoded, &decoded))

	reencoded, err := json.Marshal(decoded)
	require.NoError(t, err)
	require.JSONEq(t, string(encoded), string(reencoded), "a decoded body must re-encode identically")

	if lossless {
		require.Equal(t, request, decoded)
	}
}
