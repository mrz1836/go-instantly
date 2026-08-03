package oauth_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/oauth"
)

// Router patterns for the OAuth endpoints.
const (
	googleInitPath    = "/api/v2/oauth/google/init"
	microsoftInitPath = "/api/v2/oauth/microsoft/init"
	statusPath        = "/api/v2/oauth/session/status/:sessionId"
)

// initFixture is a spec-shaped init response.
const initFixture = `{
	"auth_url": "https://accounts.google.com/o/oauth2/auth?x=1",
	"session_id": "sess-1",
	"expires_at": "2026-08-01T10:10:00.000Z"
}`

// OAuthTestSuite exercises the OAuth API service against the mock router.
type OAuthTestSuite struct {
	instantlytest.Suite
}

// TestOAuthSuite runs the OAuth API suite.
func TestOAuthSuite(t *testing.T) {
	suite.Run(t, new(OAuthTestSuite))
}

// TestInitGoogle verifies the init request sends no body and the response
// decodes.
func (s *OAuthTestSuite) TestInitGoogle() {
	s.Router.Post(googleInitPath, func(w http.ResponseWriter, req *http.Request) {
		body, err := instantlytest.ReadAll(req)
		s.NoError(err)
		s.Empty(body, "an init request carries no body")

		_, _ = w.Write([]byte(initFixture))
	})

	result, err := s.svc().InitGoogle(context.Background())

	s.Require().NoError(err)
	s.Equal("https://accounts.google.com/o/oauth2/auth?x=1", result.AuthURL)
	s.Equal("sess-1", result.SessionID)
}

// TestInitMicrosoft verifies the init request sends no body and the response
// decodes.
func (s *OAuthTestSuite) TestInitMicrosoft() {
	s.Router.Post(microsoftInitPath, func(w http.ResponseWriter, req *http.Request) {
		body, err := instantlytest.ReadAll(req)
		s.NoError(err)
		s.Empty(body, "an init request carries no body")

		_, _ = w.Write([]byte(initFixture))
	})

	result, err := s.svc().InitMicrosoft(context.Background())

	s.Require().NoError(err)
	s.Equal("sess-1", result.SessionID)
}

// TestSessionStatusSuccess verifies a successful session decodes, including the
// connected account details.
func (s *OAuthTestSuite) TestSessionStatusSuccess() {
	s.Router.Get(statusPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("sess-1", instantlytest.PathParam(req, "sessionId"))
		_, _ = w.Write([]byte(`{"status":"success","email":"user@example.com","name":"John Doe"}`))
	})

	status, err := s.svc().SessionStatus(context.Background(), "sess-1")

	s.Require().NoError(err)
	s.Equal(oauth.StatusSuccess, status.Status)
	s.Equal("user@example.com", status.Email)
	s.Equal("John Doe", status.Name)
}

// TestSessionStatusPending verifies a still-running session decodes as pending.
func (s *OAuthTestSuite) TestSessionStatusPending() {
	s.Router.Get(statusPath, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"pending"}`))
	})

	status, err := s.svc().SessionStatus(context.Background(), "sess-1")

	s.Require().NoError(err)
	s.Equal(oauth.StatusPending, status.Status)
	s.Empty(status.Email)
}

// TestSessionStatusError verifies the documented HTTP-200-embedded-error case: a
// session that ends in an error is delivered as an HTTP 200 body carrying a
// top-level error code, which the client surfaces as an *instantly.APIError.
func (s *OAuthTestSuite) TestSessionStatusError() {
	s.Router.Get(statusPath, func(w http.ResponseWriter, _ *http.Request) {
		// HTTP 200, but the body carries a top-level error code.
		_, _ = w.Write([]byte(
			`{"status":"error","error":"access_denied","error_description":"User denied access"}`,
		))
	})

	status, err := s.svc().SessionStatus(context.Background(), "sess-1")

	s.Require().Error(err)
	s.Nil(status, "an error must never hand back a partly populated status")

	var apiErr *instantly.APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal("access_denied", apiErr.Code)
	s.Equal(int64(http.StatusOK), apiErr.StatusCode, "the error arrived inside an HTTP 200 body")
}

// TestSessionStatusPathEscape verifies a session id carrying reserved characters
// is escaped into the request path rather than altering it.
func (s *OAuthTestSuite) TestSessionStatusPathEscape() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, `{"status":"pending"}`), nil
			},
		)},
	))

	_, err := oauth.New(client).SessionStatus(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/oauth/session/status/..%2Fadmin%3Fx=1", requestURI)
}

// TestFailures verifies every endpoint surfaces the documented API error.
func (s *OAuthTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "init google", Method: http.MethodPost, Path: googleInitPath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.InitGoogle(ctx); return err },
		},
		{
			Name: "init microsoft", Method: http.MethodPost, Path: microsoftInitPath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.InitMicrosoft(ctx); return err },
		},
		{
			Name: "session status", Method: http.MethodGet, Path: statusPath, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.SessionStatus(ctx, "sess-1"); return err },
		},
	})
}

// TestParsedExpiresAt verifies the RFC 3339 accessor parses a valid timestamp
// and reports an error for an unparseable one.
func (s *OAuthTestSuite) TestParsedExpiresAt() {
	got, err := (&oauth.InitResult{ExpiresAt: "2026-08-01T10:10:00.000Z"}).ParsedExpiresAt()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&oauth.InitResult{ExpiresAt: "not-a-timestamp"}).ParsedExpiresAt()
	s.Require().Error(err)
}

// svc builds an OAuth service pointed at the suite's mock client.
func (s *OAuthTestSuite) svc() *oauth.Service {
	return oauth.New(s.Client)
}
