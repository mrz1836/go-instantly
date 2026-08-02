package instantly

import (
	"context"
	"encoding/json"
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

	// mediaTypeJSON is the media type used for both Accept and Content-Type.
	mediaTypeJSON = "application/json"

	// successBody is a minimal successful JSON response body.
	successBody = `{"status":"success"}`
)

// InstantlyTestSuite bootstraps an in-process API server for the whole package.
//
// Every test routes through the mock router rather than the live Instantly API:
// no test in this repository is permitted to reach api.instantly.ai.
type InstantlyTestSuite struct {
	suite.Suite

	mux    *TestRouter
	server *httptest.Server
	client *Client
}

// SetupSuite starts the mock API server and points a client at it.
func (s *InstantlyTestSuite) SetupSuite() {
	s.mux = NewTestRouter()
	s.server = httptest.NewServer(s.mux)

	s.client = NewClient(testAPIKey)
	s.client.BaseURL = s.server.URL
}

// TearDownSuite shuts the mock API server down.
func (s *InstantlyTestSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
}

// TestInstantlySuite runs the package suite.
func TestInstantlySuite(t *testing.T) {
	suite.Run(t, new(InstantlyTestSuite))
}

// TestNewClient verifies the constructor defaults.
func (s *InstantlyTestSuite) TestNewClient() {
	client := NewClient("some-key")

	s.Require().NotNil(client)
	s.Equal("some-key", client.APIKey)
	s.Equal(defaultBaseURL, client.BaseURL)
	s.NotNil(client.HTTPClient)
}

// TestRequestHeaders verifies the bearer token and JSON headers are always sent.
func (s *InstantlyTestSuite) TestRequestHeaders() {
	s.mux.Get("/api/v2/headers", func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testAuthHeader, req.Header.Get("Authorization"))
		s.Equal("application/json", req.Header.Get("Accept"))
		s.Equal("application/json", req.Header.Get("Content-Type"))
		_, _ = w.Write([]byte(successBody))
	})

	var result map[string]string
	err := s.client.get(context.Background(), "/api/v2/headers", &result)

	s.Require().NoError(err)
	s.Equal("success", result["status"])
}

// TestPayloadRoundTrip verifies a request body reaches the server intact and the
// response decodes into the destination.
func (s *InstantlyTestSuite) TestPayloadRoundTrip() {
	s.mux.Post("/api/v2/echo", func(w http.ResponseWriter, req *http.Request) {
		var received map[string]string
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("hello", received["subject"])
		_, _ = w.Write([]byte(`{"subject":"hello","id":"abc"}`))
	})

	var result map[string]string
	err := s.client.post(
		context.Background(), "/api/v2/echo", map[string]string{"subject": "hello"}, &result,
	)

	s.Require().NoError(err)
	s.Equal("abc", result["id"])
}

// TestVerbHelpers exercises every private verb helper against the mock router.
func (s *InstantlyTestSuite) TestVerbHelpers() {
	const path = "/api/v2/verbs"

	handler := func(expected string) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			s.Equal(expected, req.Method)
			_, _ = w.Write([]byte(successBody))
		}
	}

	s.mux.Get(path, handler(http.MethodGet))
	s.mux.Post(path, handler(http.MethodPost))
	s.mux.Patch(path, handler(http.MethodPatch))
	s.mux.Put(path, handler(http.MethodPut))
	s.mux.Delete(path, handler(http.MethodDelete))

	ctx := context.Background()
	payload := map[string]string{"field": "value"}

	var result map[string]string
	s.Require().NoError(s.client.get(ctx, path, &result))
	s.Require().NoError(s.client.post(ctx, path, payload, &result))
	s.Require().NoError(s.client.patch(ctx, path, payload, &result))
	s.Require().NoError(s.client.put(ctx, path, payload, &result))
	s.Require().NoError(s.client.delete(ctx, path, &result))
	s.Equal("success", result["status"])
}

// TestRequestGetBodyIsReplayable verifies the payload can be replayed on a
// redirect or retry, and that a bodyless request carries no replay function.
func (s *InstantlyTestSuite) TestRequestGetBodyIsReplayable() {
	var captured *http.Request

	client := NewClient(testAPIKey)
	client.BaseURL = s.server.URL
	client.HTTPClient = &http.Client{Transport: roundTripFunc(
		func(req *http.Request) (*http.Response, error) {
			captured = req
			return jsonResponse(http.StatusOK, successBody), nil
		},
	)}

	err := client.post(
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
	s.Require().NoError(client.get(context.Background(), "/api/v2/replay", nil))
	s.Require().NotNil(captured)
	s.Nil(captured.GetBody)
}

// TestDefaultBaseURL verifies an empty BaseURL falls back to the documented API
// host. The transport is intercepted, so no live request is ever made.
func (s *InstantlyTestSuite) TestDefaultBaseURL() {
	var requestURL string

	client := &Client{
		APIKey: testAPIKey,
		HTTPClient: &http.Client{Transport: roundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURL = req.URL.String()
				return jsonResponse(http.StatusOK, successBody), nil
			},
		)},
	}

	s.Require().NoError(client.get(context.Background(), "/api/v2/emails", nil))
	s.Equal(defaultBaseURL+"/api/v2/emails", requestURL)
}

// TestNilHTTPClientFallback verifies a zero-value HTTPClient still performs the
// request through the default HTTP client.
func (s *InstantlyTestSuite) TestNilHTTPClientFallback() {
	s.mux.Get("/api/v2/no-http-client", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(successBody))
	})

	client := &Client{APIKey: testAPIKey, BaseURL: s.server.URL}

	var result map[string]string
	err := client.get(context.Background(), "/api/v2/no-http-client", &result)

	s.Require().NoError(err)
	s.Equal("success", result["status"])
}

// TestNilDestination verifies a response is discarded when no destination is
// supplied, which is how the write-only endpoints call through.
func (s *InstantlyTestSuite) TestNilDestination() {
	s.mux.Get("/api/v2/nil-dst", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(successBody))
	})

	err := s.client.get(context.Background(), "/api/v2/nil-dst", nil)
	s.Require().NoError(err)
}

// TestContextCancellation verifies a canceled context aborts the request.
func (s *InstantlyTestSuite) TestContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var result map[string]string
	err := s.client.get(ctx, "/api/v2/canceled", &result)

	s.Require().Error(err)
	s.Require().ErrorIs(err, context.Canceled)
}

// TestPayloadMarshalError verifies an unencodable payload fails before any
// request is made.
func (s *InstantlyTestSuite) TestPayloadMarshalError() {
	err := s.client.post(context.Background(), "/api/v2/marshal", make(chan int), nil)

	s.Require().Error(err)
	s.Require().ErrorContains(err, "failed to encode request payload")
}

// TestInvalidURL verifies an unusable base URL surfaces as an error.
func (s *InstantlyTestSuite) TestInvalidURL() {
	client := NewClient(testAPIKey)
	client.BaseURL = "ht!tp://invalid url with spaces"

	err := client.get(context.Background(), "/api/v2/emails", nil)
	s.Require().Error(err)
}

// TestTransportError verifies a connection failure is returned to the caller.
func (s *InstantlyTestSuite) TestTransportError() {
	closed := httptest.NewServer(http.NotFoundHandler())
	closedURL := closed.URL
	closed.Close()

	client := NewClient(testAPIKey)
	client.BaseURL = closedURL

	err := client.get(context.Background(), "/api/v2/emails", nil)
	s.Require().Error(err)
}

// TestResponseDecodeError verifies a malformed success body is reported rather
// than silently ignored.
func (s *InstantlyTestSuite) TestResponseDecodeError() {
	s.mux.Get("/api/v2/bad-json", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not json at all`))
	})

	var result map[string]string
	err := s.client.get(context.Background(), "/api/v2/bad-json", &result)

	s.Require().Error(err)
	s.Require().ErrorContains(err, "failed to decode response with status 200")
}

// TestRouterPathParameters verifies the mock router extracts :param segments.
func (s *InstantlyTestSuite) TestRouterPathParameters() {
	s.mux.Get("/api/v2/emails/:id/detail", func(w http.ResponseWriter, req *http.Request) {
		s.Equal("email-123", GetPathParam(req, "id"))
		s.Empty(GetPathParam(req, "missing"))
		_, _ = w.Write([]byte(successBody))
	})

	err := s.client.get(context.Background(), "/api/v2/emails/email-123/detail", nil)
	s.Require().NoError(err)
}

// TestRouterReplacesDuplicateRoutes verifies re-registering a method and pattern
// replaces the previous handler instead of shadowing it.
func (s *InstantlyTestSuite) TestRouterReplacesDuplicateRoutes() {
	const path = "/api/v2/replaced"

	s.mux.Get(path, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"first"}`))
	})
	s.mux.Get(path, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"second"}`))
	})

	var result map[string]string
	err := s.client.get(context.Background(), path, &result)

	s.Require().NoError(err)
	s.Equal("second", result["status"])
}

// TestRouterUnknownRouteIsNotFound verifies unregistered paths fall through to a
// 404, which the client reports as an error.
func (s *InstantlyTestSuite) TestRouterUnknownRouteIsNotFound() {
	err := s.client.get(context.Background(), "/api/v2/never-registered", nil)

	s.Require().Error(err)
	s.Require().ErrorContains(err, "404")
}

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
