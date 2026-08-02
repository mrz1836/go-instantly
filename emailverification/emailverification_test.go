package emailverification_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/emailverification"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// Router patterns and identifiers the email-verification endpoints are exercised
// with. The patterns carry the full request path, including the /api/v2 prefix.
const (
	// createPath is the create-verification endpoint.
	createPath = "/api/v2/email-verification"

	// checkPattern is the router pattern for the check-verification endpoint.
	checkPattern = "/api/v2/email-verification/:email"

	// testEmail is the address the endpoints verify.
	testEmail = "example@example.com"
)

// verifiedFixture is a completed verification with every documented field
// populated, including the nullable ones.
const verifiedFixture = `{
	"email": "example@example.com",
	"verification_status": "verified",
	"status": "success",
	"catch_all": true,
	"credits": 100,
	"credits_used": 1
}`

// pendingFixture is an asynchronous verification still in progress, with the
// nullable fields explicitly null and the mixed catch_all reported as "pending".
const pendingFixture = `{
	"email": "example@example.com",
	"verification_status": "pending",
	"status": null,
	"catch_all": "pending",
	"credits": null,
	"credits_used": null
}`

// EmailVerificationTestSuite exercises the Email Verification API service against
// the mock router.
type EmailVerificationTestSuite struct {
	instantlytest.Suite
}

// TestEmailVerificationSuite runs the Email Verification API suite.
func TestEmailVerificationSuite(t *testing.T) {
	suite.Run(t, new(EmailVerificationTestSuite))
}

// TestCreate verifies the request body reaches the API and a completed
// verification decodes, including its nullable fields and the raw catch_all.
func (s *EmailVerificationTestSuite) TestCreate() {
	s.Router.Post(createPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodPost, req.Method)
		s.Equal(createPath, req.URL.Path)

		var received emailverification.CreateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(testEmail, received.Email)
		s.Equal("https://example.com/webhook", received.WebhookURL)

		_, _ = w.Write([]byte(verifiedFixture))
	})

	got, err := s.svc().Create(context.Background(), emailverification.CreateRequest{
		Email:      testEmail,
		WebhookURL: "https://example.com/webhook",
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(testEmail, got.Email)
	s.Equal(emailverification.StatusVerified, got.VerificationStatus)
	s.Require().NotNil(got.RequestStatus)
	s.Equal(emailverification.RequestStatusSuccess, *got.RequestStatus)
	s.JSONEq(`true`, string(got.CatchAll))
	s.Require().NotNil(got.Credits)
	s.InDelta(100, *got.Credits, 0)
	s.Require().NotNil(got.CreditsUsed)
	s.InDelta(1, *got.CreditsUsed, 0)
}

// TestCreateOmitsWebhookURL verifies an unset webhook URL is left out of the body
// entirely rather than sent as an empty string.
func (s *EmailVerificationTestSuite) TestCreateOmitsWebhookURL() {
	s.Router.Post(createPath, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.NotContains(received, "webhook_url")

		_, _ = w.Write([]byte(verifiedFixture))
	})

	got, err := s.svc().Create(context.Background(), emailverification.CreateRequest{Email: testEmail})

	s.Require().NoError(err)
	s.Require().NotNil(got)
}

// TestCreatePending verifies an asynchronous verification returns the pending
// status inside an HTTP 200 rather than as an error, with its nullable fields
// staying nil and the mixed catch_all preserved verbatim.
func (s *EmailVerificationTestSuite) TestCreatePending() {
	s.Router.Post(createPath, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(pendingFixture))
	})

	got, err := s.svc().Create(context.Background(), emailverification.CreateRequest{Email: testEmail})

	s.Require().NoError(err, "a pending status is a normal result, not an error")
	s.Require().NotNil(got)
	s.Equal(emailverification.StatusPending, got.VerificationStatus)

	// The nullable fields stay nil rather than collapsing to a zero value.
	s.Nil(got.RequestStatus)
	s.Nil(got.Credits)
	s.Nil(got.CreditsUsed)

	// The mixed true/false/"pending" catch_all is preserved as sent.
	s.JSONEq(`"pending"`, string(got.CatchAll))
}

// TestCheck verifies the address is escaped onto the path and the verification
// decodes.
func (s *EmailVerificationTestSuite) TestCheck() {
	s.Router.Get(checkPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodGet, req.Method)
		s.Equal(testEmail, instantlytest.PathParam(req, "email"))

		_, _ = w.Write([]byte(pendingFixture))
	})

	got, err := s.svc().Check(context.Background(), testEmail)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(testEmail, got.Email)
	s.Equal(emailverification.StatusPending, got.VerificationStatus)
}

// TestPathParametersAreEscaped verifies a caller-supplied address cannot rewrite
// the request path. The transport is intercepted so the raw path can be asserted
// before a server decodes it.
func (s *EmailVerificationTestSuite) TestPathParametersAreEscaped() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, verifiedFixture), nil
			},
		)},
	))

	_, err := emailverification.New(client).Check(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/email-verification/..%2Fadmin%3Fx=1", requestURI)
}

// TestFailures verifies every endpoint surfaces the documented API error.
func (s *EmailVerificationTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "create", Method: http.MethodPost, Path: createPath, Status: http.StatusPaymentRequired,
			Call: func() error { _, err := svc.Create(ctx, emailverification.CreateRequest{}); return err },
		},
		{
			Name: "check", Method: http.MethodGet, Path: checkPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Check(ctx, "missing"); return err },
		},
	})
}

// svc builds an Email Verification service pointed at the suite's mock client.
func (s *EmailVerificationTestSuite) svc() *emailverification.Service {
	return emailverification.New(s.Client)
}
