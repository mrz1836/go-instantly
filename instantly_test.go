package instantly

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	// testAPIKey is the V2 API key every suite request authenticates with.
	testAPIKey = "test-api-key"

	// testAuthHeader is the exact Authorization header value the client must send.
	testAuthHeader = "Bearer " + testAPIKey

	// successBody is a minimal successful JSON response body.
	successBody = `{"status":"success"}`
)

// clientTestSuite exercises the low-level client plumbing against a stdlib mux.
//
// These tests stay in package instantly because they reach unexported client
// fields and the unexported checkResponse path; they cannot import
// internal/instantlytest, which imports instantly and would form a test cycle.
type clientTestSuite struct {
	suite.Suite

	mux    *http.ServeMux
	server *httptest.Server
	client *Client
}

// SetupTest gives every test a fresh mux and server, so route registrations
// never collide across tests.
func (s *clientTestSuite) SetupTest() {
	s.mux = http.NewServeMux()
	s.server = httptest.NewServer(s.mux)
	s.client = NewClient(testAPIKey, WithBaseURL(s.server.URL))
}

// TearDownTest shuts the per-test server down.
func (s *clientTestSuite) TearDownTest() {
	if s.server != nil {
		s.server.Close()
	}
}

// TestClientSuite runs the client plumbing suite.
func TestClientSuite(t *testing.T) {
	suite.Run(t, new(clientTestSuite))
}

// TestNewClientDefaults verifies the constructor defaults.
func (s *clientTestSuite) TestNewClientDefaults() {
	client := NewClient("some-key")

	s.Require().NotNil(client)
	s.Equal("some-key", client.apiKey)
	s.Equal(defaultBaseURL, client.baseURL)
	s.NotNil(client.httpClient)
	s.Empty(client.userAgent)
	s.Nil(client.headers)
}

// TestNewClientOptions verifies each functional option is applied, and that a
// nil option is ignored rather than panicking.
func (s *clientTestSuite) TestNewClientOptions() {
	custom := &http.Client{}
	client := NewClient(
		"some-key",
		nil,
		WithHTTPClient(custom),
		WithBaseURL("https://example.test"),
		WithUserAgent("go-instantly/test"),
		WithHTTPHeader("X-Trace", "abc"),
		WithHTTPHeader("X-Trace", "def"),
	)

	s.Same(custom, client.httpClient)
	s.Equal("https://example.test", client.baseURL)
	s.Equal("go-instantly/test", client.userAgent)
	s.Equal([]string{"abc", "def"}, client.headers.Values("X-Trace"))
}

// TestPtr verifies the pointer helper returns a pointer to its argument.
func (s *clientTestSuite) TestPtr() {
	s.True(*Ptr(true))
	s.Equal("x", *Ptr("x"))
	s.Equal(7, *Ptr(7))
}

// TestRequestHeaders verifies the bearer token and JSON headers are always sent,
// along with a configured user agent and extra header.
func (s *clientTestSuite) TestRequestHeaders() {
	s.handle(http.MethodGet, "/api/v2/headers", func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testAuthHeader, req.Header.Get("Authorization"))
		s.Equal("application/json", req.Header.Get("Accept"))
		s.Equal("application/json", req.Header.Get("Content-Type"))
		s.Equal("go-instantly/test", req.Header.Get("User-Agent"))
		s.Equal("abc", req.Header.Get("X-Trace"))
		_, _ = w.Write([]byte(successBody))
	})

	client := NewClient(
		testAPIKey,
		WithBaseURL(s.server.URL),
		WithUserAgent("go-instantly/test"),
		WithHTTPHeader("X-Trace", "abc"),
	)

	var result map[string]string
	err := client.Get(context.Background(), "/api/v2/headers", &result)

	s.Require().NoError(err)
	s.Equal("success", result["status"])
}

// TestExtraHeaderCannotOverrideMandatory verifies a caller-supplied header does
// not clobber the Authorization the client must send.
func (s *clientTestSuite) TestExtraHeaderCannotOverrideMandatory() {
	s.handle(http.MethodGet, "/api/v2/protected", func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testAuthHeader, req.Header.Get("Authorization"))
		_, _ = w.Write([]byte(successBody))
	})

	client := NewClient(
		testAPIKey,
		WithBaseURL(s.server.URL),
		WithHTTPHeader("Authorization", "Bearer forged"),
	)

	s.Require().NoError(client.Get(context.Background(), "/api/v2/protected", nil))
}

// TestPayloadRoundTrip verifies a request body reaches the server intact and the
// response decodes into the destination.
func (s *clientTestSuite) TestPayloadRoundTrip() {
	s.handle(http.MethodPost, "/api/v2/echo", func(w http.ResponseWriter, req *http.Request) {
		var received map[string]string
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("hello", received["subject"])
		_, _ = w.Write([]byte(`{"subject":"hello","id":"abc"}`))
	})

	var result map[string]string
	err := s.client.Post(
		context.Background(), "/api/v2/echo", map[string]string{"subject": "hello"}, &result,
	)

	s.Require().NoError(err)
	s.Equal("abc", result["id"])
}

// TestVerbHelpers exercises every exported verb helper against the mux.
func (s *clientTestSuite) TestVerbHelpers() {
	const path = "/api/v2/verbs"

	handler := func(expected string) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			s.Equal(expected, req.Method)
			_, _ = w.Write([]byte(successBody))
		}
	}

	s.handle(http.MethodGet, path, handler(http.MethodGet))
	s.handle(http.MethodPost, path, handler(http.MethodPost))
	s.handle(http.MethodPatch, path, handler(http.MethodPatch))
	s.handle(http.MethodPut, path, handler(http.MethodPut))
	s.handle(http.MethodDelete, path, handler(http.MethodDelete))

	ctx := context.Background()
	payload := map[string]string{"field": "value"}

	var result map[string]string
	s.Require().NoError(s.client.Get(ctx, path, &result))
	s.Require().NoError(s.client.Post(ctx, path, payload, &result))
	s.Require().NoError(s.client.Patch(ctx, path, payload, &result))
	s.Require().NoError(s.client.Put(ctx, path, payload, &result))
	s.Require().NoError(s.client.Delete(ctx, path, &result))
	s.Equal("success", result["status"])
}

// TestDoRaw verifies the raw body is returned undecoded for endpoints that
// answer with a non-JSON payload.
func (s *clientTestSuite) TestDoRaw() {
	s.handle(http.MethodGet, "/api/v2/download", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("col1,col2\r\na,b\r\n"))
	})

	body, err := s.client.DoRaw(context.Background(), http.MethodGet, "/api/v2/download", nil)

	s.Require().NoError(err)
	s.Equal("col1,col2\r\na,b\r\n", string(body))
}

// TestDoRawSurfacesErrors verifies DoRaw runs the same error handling as Do.
func (s *clientTestSuite) TestDoRawSurfacesErrors() {
	s.handle(http.MethodGet, "/api/v2/download-fail", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"statusCode":404,"error":"Not Found"}`))
	})

	body, err := s.client.DoRaw(context.Background(), http.MethodGet, "/api/v2/download-fail", nil)

	s.Require().Error(err)
	s.Nil(body, "a failed raw request must not hand back a body")

	var apiErr *APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(int64(http.StatusNotFound), apiErr.StatusCode)
}

// TestDeleteWithBody verifies Do can carry a payload on a DELETE, which a few
// endpoints require.
func (s *clientTestSuite) TestDeleteWithBody() {
	s.handle(http.MethodDelete, "/api/v2/bulk", func(w http.ResponseWriter, req *http.Request) {
		var received map[string][]string
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal([]string{"a", "b"}, received["ids"])
		_, _ = w.Write([]byte(successBody))
	})

	err := s.client.Do(
		context.Background(), http.MethodDelete, "/api/v2/bulk", map[string][]string{"ids": {"a", "b"}}, nil,
	)

	s.Require().NoError(err)
}

// TestRequestGetBodyIsReplayable verifies the payload can be replayed on a
// redirect or retry, and that a bodyless request carries no replay function.
func (s *clientTestSuite) TestRequestGetBodyIsReplayable() {
	var captured *http.Request

	client := NewClient(testAPIKey, WithHTTPClient(&http.Client{Transport: roundTripFunc(
		func(req *http.Request) (*http.Response, error) {
			captured = req
			return jsonResponse(http.StatusOK, successBody), nil
		},
	)}))

	err := client.Post(
		context.Background(), "/api/v2/replay", map[string]string{"subject": "hello"}, nil,
	)
	s.Require().NoError(err)
	s.Require().NotNil(captured)
	s.Require().NotNil(captured.GetBody)

	replayed, err := captured.GetBody()
	s.Require().NoError(err)

	body, err := io.ReadAll(replayed)
	s.Require().NoError(err)
	s.JSONEq(`{"subject":"hello"}`, string(body))

	captured = nil
	s.Require().NoError(client.Get(context.Background(), "/api/v2/replay", nil))
	s.Require().NotNil(captured)
	s.Nil(captured.GetBody)
}

// TestDefaultBaseURL verifies an empty baseURL falls back to the documented API
// host. The transport is intercepted, so no live request is ever made.
func (s *clientTestSuite) TestDefaultBaseURL() {
	var requestURL string

	client := &Client{
		apiKey: testAPIKey,
		httpClient: &http.Client{Transport: roundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURL = req.URL.String()
				return jsonResponse(http.StatusOK, successBody), nil
			},
		)},
	}

	s.Require().NoError(client.Get(context.Background(), "/api/v2/emails", nil))
	s.Equal(defaultBaseURL+"/api/v2/emails", requestURL)
}

// TestNilHTTPClientFallback verifies a nil httpClient still performs the request
// through the default HTTP client.
func (s *clientTestSuite) TestNilHTTPClientFallback() {
	s.handle(http.MethodGet, "/api/v2/no-http-client", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(successBody))
	})

	client := &Client{apiKey: testAPIKey, baseURL: s.server.URL}

	var result map[string]string
	err := client.Get(context.Background(), "/api/v2/no-http-client", &result)

	s.Require().NoError(err)
	s.Equal("success", result["status"])
}

// TestNilDestination verifies a response is discarded when no destination is
// supplied, which is how the write-only endpoints call through.
func (s *clientTestSuite) TestNilDestination() {
	s.handle(http.MethodGet, "/api/v2/nil-dst", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(successBody))
	})

	err := s.client.Get(context.Background(), "/api/v2/nil-dst", nil)
	s.Require().NoError(err)
}

// TestContextCancellation verifies a canceled context aborts the request.
func (s *clientTestSuite) TestContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var result map[string]string
	err := s.client.Get(ctx, "/api/v2/canceled", &result)

	s.Require().Error(err)
	s.Require().ErrorIs(err, context.Canceled)
}

// TestPayloadMarshalError verifies an unencodable payload fails before any
// request is made.
func (s *clientTestSuite) TestPayloadMarshalError() {
	err := s.client.Post(context.Background(), "/api/v2/marshal", make(chan int), nil)

	s.Require().Error(err)
	s.Require().ErrorContains(err, "failed to encode request payload")
}

// TestInvalidURL verifies an unusable base URL surfaces as an error.
func (s *clientTestSuite) TestInvalidURL() {
	client := NewClient(testAPIKey, WithBaseURL("ht!tp://invalid url with spaces"))

	err := client.Get(context.Background(), "/api/v2/emails", nil)
	s.Require().Error(err)
}

// TestTransportError verifies a connection failure is returned to the caller.
func (s *clientTestSuite) TestTransportError() {
	closed := httptest.NewServer(http.NotFoundHandler())
	closedURL := closed.URL
	closed.Close()

	client := NewClient(testAPIKey, WithBaseURL(closedURL))

	err := client.Get(context.Background(), "/api/v2/emails", nil)
	s.Require().Error(err)
}

// TestResponseDecodeError verifies a malformed success body is reported rather
// than silently ignored.
func (s *clientTestSuite) TestResponseDecodeError() {
	s.handle(http.MethodGet, "/api/v2/bad-json", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not json at all`))
	})

	var result map[string]string
	err := s.client.Get(context.Background(), "/api/v2/bad-json", &result)

	s.Require().Error(err)
	s.Require().ErrorContains(err, "failed to decode response with status 200")
}

// TestResponseBodyReadError verifies a body that fails mid-read surfaces the
// read error rather than being silently treated as an empty response.
func (s *clientTestSuite) TestResponseBodyReadError() {
	client := NewClient(testAPIKey, WithHTTPClient(&http.Client{Transport: roundTripFunc(
		func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     http.StatusText(http.StatusOK),
				Header:     http.Header{"Content-Type": []string{mediaTypeJSON}},
				Body:       io.NopCloser(errReader{}),
			}, nil
		},
	)}))

	var result map[string]string
	err := client.Get(context.Background(), "/api/v2/read-error", &result)

	s.Require().Error(err)
	s.Require().ErrorContains(err, "unexpected read failure")
}

// handle registers a handler for a "METHOD /path" pattern.
func (s *clientTestSuite) handle(method, path string, handler http.HandlerFunc) {
	s.mux.HandleFunc(method+" "+path, handler)
}

// errReader is an io.Reader that always fails, used to exercise the body-read
// error path.
type errReader struct{}

// Read always returns an error.
func (errReader) Read([]byte) (int, error) {
	return 0, errUnexpectedRead
}

// errUnexpectedRead is the failure errReader returns.
var errUnexpectedRead = errors.New("unexpected read failure")

// roundTripFunc adapts a function to http.RoundTripper so transport-level
// details can be asserted without a network round trip.
type roundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the http.RoundTripper interface.
func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// jsonResponse builds a canned JSON response for an intercepted transport.
func jsonResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Header:     http.Header{"Content-Type": []string{mediaTypeJSON}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
