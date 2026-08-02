package account_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/account"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// Router patterns and identifiers the account endpoints are exercised with.
const (
	listPath        = "/api/v2/accounts"
	emailPattern    = "/api/v2/accounts/:email"
	pausePattern    = "/api/v2/accounts/:email/pause"
	resumePattern   = "/api/v2/accounts/:email/resume"
	markFixedPatt   = "/api/v2/accounts/:email/mark-fixed"
	bulkPausePath   = "/api/v2/accounts/pause"
	movePath        = "/api/v2/accounts/move"
	enableWarmup    = "/api/v2/accounts/warmup/enable"
	disableWarmup   = "/api/v2/accounts/warmup/disable"
	warmupAnalytics = "/api/v2/accounts/warmup-analytics"
	dailyAnalytics  = "/api/v2/accounts/analytics/daily"
	ctdStatusPath   = "/api/v2/accounts/ctd/status"
	vitalsPath      = "/api/v2/accounts/test/vitals"

	testEmail = "sender@example.com"
)

// accountFixture is a spec-shaped account with the required fields plus a
// representative set of populated nullable fields.
const accountFixture = `{
	"email": "sender@example.com",
	"first_name": "Jon",
	"last_name": "Doe",
	"organization": "org-uuid-1",
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"timestamp_updated": "2026-08-01T11:00:00.000Z",
	"status": 1,
	"warmup_status": 1,
	"provider_code": 2,
	"setup_pending": false,
	"is_managed_account": false,
	"sending_gap": 10,
	"signature": "Best, Jon",
	"reply_to": "replies@example.com",
	"daily_limit": 100,
	"stat_warmup_score": 98,
	"enable_slow_ramp": true,
	"autofix_failed": null,
	"tracking_domain_name": "track.example.com",
	"status_message": {"code": "OK"},
	"warmup": {"limit": 20, "reply_rate": 30, "increment": "1", "advanced": {"x": 1}}
}`

// accountFixtureNulls has every nullable field explicitly null.
const accountFixtureNulls = `{
	"email": "bare@example.com",
	"first_name": "Bare",
	"last_name": "Account",
	"organization": "org-uuid-1",
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"timestamp_updated": "2026-08-01T11:00:00.000Z",
	"status": 2,
	"warmup_status": 0,
	"provider_code": 1,
	"setup_pending": true,
	"is_managed_account": false,
	"signature": null,
	"reply_to": null,
	"added_by": null,
	"modified_by": null,
	"daily_limit": null,
	"stat_warmup_score": null,
	"enable_slow_ramp": null,
	"autofix_failed": null,
	"tracking_domain_name": null
}`

// AccountTestSuite exercises the Account API service against the mock router.
type AccountTestSuite struct {
	instantlytest.Suite
}

// TestAccountSuite runs the Account API suite.
func TestAccountSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}

// TestCreate verifies the create body reaches the API and the account decodes.
func (s *AccountTestSuite) TestCreate() {
	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		var received account.CreateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(testEmail, received.Email)
		s.Equal(account.ProviderGoogle, received.ProviderCode)
		s.Equal("imap.example.com", received.IMAPHost)
		s.Equal(int64(993), received.IMAPPort)

		_, _ = w.Write([]byte(accountFixture))
	})

	got, err := s.svc().Create(context.Background(), account.CreateRequest{
		Email:        testEmail,
		FirstName:    "Jon",
		LastName:     "Doe",
		ProviderCode: account.ProviderGoogle,
		IMAPUsername: testEmail,
		IMAPPassword: "secret",
		IMAPHost:     "imap.example.com",
		IMAPPort:     993,
		SMTPUsername: testEmail,
		SMTPPassword: "secret",
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		DailyLimit:   instantly.Ptr(100.0),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(testEmail, got.Email)
	s.Equal(account.StatusActive, got.Status)
	s.Equal(account.ProviderGoogle, got.ProviderCode)
}

// TestCreateFailure verifies a rejected create returns no account.
func (s *AccountTestSuite) TestCreateFailure() {
	s.Router.Post(listPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusPaymentRequired, "Payment Required", "limit")
	})

	got, err := s.svc().Create(context.Background(), account.CreateRequest{Email: testEmail})

	instantlytest.AssertAPIError(s.T(), err, http.StatusPaymentRequired)
	s.Nil(got)
}

// TestList verifies a page decodes, including nullable-vs-zero fields and tags.
func (s *AccountTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("50", req.URL.Query().Get("limit"))
		s.Equal("1", req.URL.Query().Get("status"))
		s.Equal("true", req.URL.Query().Get("include_tags"))

		_, _ = w.Write([]byte(
			`{"items":[` + withTags(accountFixture) + `,` + accountFixtureNulls +
				`],"next_starting_after":"cursor-2"}`,
		))
	})

	page, err := s.svc().List(context.Background(),
		account.WithLimit(50),
		account.WithStatus(account.StatusActive),
		account.WithIncludeTags(true),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)

	populated := page.Items[0]
	s.Require().NotNil(populated.DailyLimit)
	s.InDelta(100, *populated.DailyLimit, 0)
	s.Require().NotNil(populated.Warmup)
	s.Equal(account.WarmupIncrement1, populated.Warmup.Increment)
	s.Require().Len(populated.Tags, 1)
	s.Equal("Important", populated.Tags[0].Label)

	// Nullable fields stay nil rather than collapsing to a zero value.
	bare := page.Items[1]
	s.Nil(bare.DailyLimit)
	s.Nil(bare.Signature)
	s.Nil(bare.EnableSlowRamp)
	s.Nil(bare.Warmup)
	s.Empty(bare.Tags)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string.
func (s *AccountTestSuite) TestListWithoutOptions() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery)
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background(), nil)

	s.Require().NoError(err)
	s.Empty(page.Items)
}

// TestListFailure verifies a failed list returns no page.
func (s *AccountTestSuite) TestListFailure() {
	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "slow down")
	})

	page, err := s.svc().List(context.Background())

	instantlytest.AssertAPIError(s.T(), err, http.StatusTooManyRequests)
	s.Nil(page)
}

// TestGet verifies a single account decodes.
func (s *AccountTestSuite) TestGet() {
	s.Router.Get(emailPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testEmail, instantlytest.PathParam(req, "email"))
		_, _ = w.Write([]byte(accountFixture))
	})

	got, err := s.svc().Get(context.Background(), testEmail)

	s.Require().NoError(err)
	s.Equal(testEmail, got.Email)
	s.Require().NotNil(got.Signature)
	s.Equal("Best, Jon", *got.Signature)
	s.JSONEq(`{"code":"OK"}`, string(got.StatusMessage))
}

// TestGetFailure verifies a missing account returns no value.
func (s *AccountTestSuite) TestGetFailure() {
	s.Router.Get(emailPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no account")
	})

	got, err := s.svc().Get(context.Background(), "missing@example.com")

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestUpdate verifies the patch body is sent and the account decodes.
func (s *AccountTestSuite) TestUpdate() {
	s.Router.Patch(emailPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testEmail, instantlytest.PathParam(req, "email"))

		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.InDelta(200, received["daily_limit"], 0)
		s.Equal(true, received["enable_slow_ramp"])

		_, _ = w.Write([]byte(accountFixture))
	})

	got, err := s.svc().Update(context.Background(), testEmail, account.UpdateRequest{
		DailyLimit:     instantly.Ptr(200.0),
		EnableSlowRamp: instantly.Ptr(true),
	})

	s.Require().NoError(err)
	s.Equal(testEmail, got.Email)
}

// TestUpdateOmitsUnsetFields verifies an empty patch sends no field.
func (s *AccountTestSuite) TestUpdateOmitsUnsetFields() {
	s.Router.Patch(emailPattern, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Empty(received, "an unset patch field must not be sent")

		_, _ = w.Write([]byte(accountFixture))
	})

	got, err := s.svc().Update(context.Background(), testEmail, account.UpdateRequest{})

	s.Require().NoError(err)
	s.NotNil(got)
}

// TestUpdateFailure verifies a failed patch returns no value.
func (s *AccountTestSuite) TestUpdateFailure() {
	s.Router.Patch(emailPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no account")
	})

	got, err := s.svc().Update(context.Background(), "missing@example.com", account.UpdateRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestDelete verifies the deleted account is returned to the caller.
func (s *AccountTestSuite) TestDelete() {
	s.Router.Delete(emailPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testEmail, instantlytest.PathParam(req, "email"))
		_, _ = w.Write([]byte(accountFixture))
	})

	got, err := s.svc().Delete(context.Background(), testEmail)

	s.Require().NoError(err)
	s.Equal(testEmail, got.Email)
}

// TestDeleteFailure verifies a failed delete returns no value.
func (s *AccountTestSuite) TestDeleteFailure() {
	s.Router.Delete(emailPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no account")
	})

	got, err := s.svc().Delete(context.Background(), "missing@example.com")

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestPathParametersAreEscaped verifies a caller-supplied email cannot rewrite
// the request path.
func (s *AccountTestSuite) TestPathParametersAreEscaped() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, accountFixture), nil
			},
		)},
	))

	_, err := account.New(client).Get(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/accounts/..%2Fadmin%3Fx=1", requestURI)
}

// TestListOptions verifies each documented list query parameter is rendered by
// exactly one option, under the key and value the API expects.
func (s *AccountTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option account.ListOption
		key    string
		value  string
	}{
		{"limit", account.WithLimit(50), "limit", "50"},
		{"starting after", account.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"search", account.WithSearch("gmail.com"), "search", "gmail.com"},
		{"status", account.WithStatus(account.StatusPaused), "status", "2"},
		{"provider code", account.WithProviderCode(account.ProviderMicrosoft), "provider_code", "3"},
		{"tag ids", account.WithTagIDs("a,b"), "tag_ids", "a,b"},
		{"tag ids all", account.WithTagIDsAll("a,b"), "tag_ids_all", "a,b"},
		{"include tags", account.WithIncludeTags(true), "include_tags", "true"},
		{"filter", account.WithFilter(account.FilterPaused), "filter", "ACC_FILTER_PAUSED"},
		{"sort by", account.WithSortBy(account.SortByEmail), "sort_by", "email"},
		{"sort order", account.WithSortOrder(instantly.SortOrderAsc), "sort_order", "asc"},
		{"skip", account.WithSkip(20), "skip", "20"},
	}

	s.Require().Len(tests, 12, "every documented list query parameter needs an option")

	for _, test := range tests {
		s.Run(test.name, func() {
			q := instantly.NewQuery()
			test.option(q)

			s.Require().Equal(1, q.Len(), "an option must render exactly one query parameter")
			s.Equal(test.value, q.Get(test.key))
		})
	}
}

// svc builds an Account service pointed at the suite's mock client.
func (s *AccountTestSuite) svc() *account.Service {
	return account.New(s.Client)
}

// withTags injects a tags array into an account fixture so a list response can
// exercise the include_tags path.
func withTags(fixture string) string {
	return fixture[:len(fixture)-1] + `,"tags":[{"id":"tag-1","label":"Important","description":null}]}`
}
