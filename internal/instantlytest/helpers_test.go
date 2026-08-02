package instantlytest_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

func TestJSONResponse(t *testing.T) {
	t.Parallel()

	res := instantlytest.JSONResponse(http.StatusTeapot, `{"a":1}`)

	require.Equal(t, http.StatusTeapot, res.StatusCode)
	require.Equal(t, instantlytest.MediaTypeJSON, res.Header.Get("Content-Type"))

	body, err := instantlytest.ReadAll(&http.Request{Body: res.Body})
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	require.JSONEq(t, `{"a":1}`, string(body))
}

func TestRoundTripFunc(t *testing.T) {
	t.Parallel()

	var seen *http.Request
	rt := instantlytest.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		seen = req

		return instantlytest.JSONResponse(http.StatusOK, `{}`), nil
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v2/x", nil)
	res, err := rt.RoundTrip(req)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	require.Same(t, req, seen, "the transport receives the request unchanged")
}

func TestReadAll(t *testing.T) {
	t.Parallel()

	t.Run("reads a body", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader("payload"))
		body, err := instantlytest.ReadAll(req)
		require.NoError(t, err)
		require.Equal(t, "payload", string(body))
	})

	t.Run("a nil body reads as no bytes", func(t *testing.T) {
		t.Parallel()

		body, err := instantlytest.ReadAll(&http.Request{})
		require.NoError(t, err)
		require.Nil(t, body)
	})
}

func TestFuzzClient(t *testing.T) {
	t.Parallel()

	client := instantlytest.FuzzClient(http.StatusOK, `{"id":"x1"}`)

	var out map[string]string
	require.NoError(t, client.Get(context.Background(), "/api/v2/anything", &out))
	require.Equal(t, "x1", out["id"])
}

func TestPage(t *testing.T) {
	t.Parallel()

	require.JSONEq(t, `{"items":[{"id":"a"},{"id":"b"}]}`, instantlytest.Page(
		[]string{`{"id":"a"}`, `{"id":"b"}`}, "",
	))
	require.JSONEq(t, `{"items":[{"id":"a"}],"next_starting_after":"c2"}`, instantlytest.Page(
		[]string{`{"id":"a"}`}, "c2",
	))
}

func TestWriteAPIErrorEnvelope(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	instantlytest.WriteAPIErrorEnvelope(rec, http.StatusNotFound, "Not Found", "missing")

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.JSONEq(t, `{"statusCode":404,"error":"Not Found","message":"missing"}`, rec.Body.String())
}

func TestFailHandler(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	instantlytest.FailHandler(http.StatusTooManyRequests)(rec, req)

	require.Equal(t, http.StatusTooManyRequests, rec.Code)
	require.JSONEq(
		t,
		`{"statusCode":429,"error":"Too Many Requests","message":"request failed"}`,
		rec.Body.String(),
	)
}

func TestAssertAPIError(t *testing.T) {
	t.Parallel()

	err := instantlytest.FuzzClient(
		http.StatusUnauthorized, `{"statusCode":401,"error":"Unauthorized"}`,
	).Get(context.Background(), "/api/v2/x", nil)

	instantlytest.AssertAPIError(t, err, http.StatusUnauthorized)
}

func TestRequireStableRoundTrip(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	instantlytest.RequireStableRoundTrip(t, payload{Name: "x", Count: 3}, true)
	// lossless=false skips the equality assertion for inputs the encoder cannot
	// represent exactly; the round-trip stability is still checked.
	instantlytest.RequireStableRoundTrip(t, payload{Name: "y", Count: 0}, false)
}

// nonHelperT is a require.TestingT without a Helper method, exercising the
// helpers' optional-Helper path.
type nonHelperT struct{ failed bool }

func (n *nonHelperT) Errorf(string, ...any) { n.failed = true }
func (n *nonHelperT) FailNow()              { panic("failnow") }

func TestHelpersWithoutHelperMethod(t *testing.T) {
	t.Parallel()

	err := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return instantlytest.JSONResponse(http.StatusNotFound, `{"statusCode":404,"error":"Not Found"}`), nil
		})},
	)).Get(context.Background(), "/api/v2/x", nil)

	tt := &nonHelperT{}
	instantlytest.AssertAPIError(tt, err, http.StatusNotFound)
	require.False(t, tt.failed, "a genuine API error must satisfy AssertAPIError")

	instantlytest.RequireStableRoundTrip(tt, map[string]int{"a": 1}, true)
	require.False(t, tt.failed, "a stable value must satisfy RequireStableRoundTrip")
}
