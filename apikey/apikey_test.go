package apikey_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/apikey"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// listPath is the router pattern for the api-keys collection endpoint.
const listPath = "/api/v2/api-keys"

// idPath is the router pattern for the single-api-key endpoint.
const idPath = "/api/v2/api-keys/:id"

// keyFixture is a spec-shaped API key with every documented field populated. The
// API declares no nullable fields on an API key.
const keyFixture = `{
	"id": "key-1",
	"name": "CI deploy key",
	"scopes": ["campaigns:read", "emails:read"],
	"key": "a1b2c3d4e5f6g7h8i9j0",
	"organization_id": "org-1",
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"timestamp_updated": "2026-08-01T11:00:00.000Z"
}`

// APIKeyTestSuite exercises the API Key API service against the mock router.
type APIKeyTestSuite struct {
	instantlytest.Suite
}

// TestAPIKeySuite runs the API Key API suite.
func TestAPIKeySuite(t *testing.T) {
	suite.Run(t, new(APIKeyTestSuite))
}

// TestCreate verifies the request body reaches the API and the response decodes,
// including the typed scopes.
func (s *APIKeyTestSuite) TestCreate() {
	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		body, err := instantlytest.ReadAll(req)
		s.NoError(err)

		var got apikey.CreateRequest
		s.NoError(json.Unmarshal(body, &got))
		s.Equal("CI deploy key", got.Name)
		s.Equal([]apikey.Scope{apikey.ScopeCampaignsRead, apikey.ScopeEmailsRead}, got.Scopes)

		_, _ = w.Write([]byte(keyFixture))
	})

	key, err := s.svc().Create(context.Background(), apikey.CreateRequest{
		Name:   "CI deploy key",
		Scopes: []apikey.Scope{apikey.ScopeCampaignsRead, apikey.ScopeEmailsRead},
	})

	s.Require().NoError(err)
	s.Equal("key-1", key.ID)
	s.Equal("a1b2c3d4e5f6g7h8i9j0", key.Key)
	s.Equal([]apikey.Scope{apikey.ScopeCampaignsRead, apikey.ScopeEmailsRead}, key.Scopes)
	s.Equal("org-1", key.OrganizationID)
}

// TestList verifies a page decodes and the options are sent.
func (s *APIKeyTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(listPath, req.URL.Path)
		s.Equal("50", req.URL.Query().Get("limit"))
		s.Equal("cursor-1", req.URL.Query().Get("starting_after"))

		_, _ = w.Write([]byte(instantlytest.Page([]string{keyFixture, keyFixture}, "cursor-2")))
	})

	page, err := s.svc().List(context.Background(),
		apikey.WithLimit(50),
		apikey.WithStartingAfter("cursor-1"),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)
	s.Equal("key-1", page.Items[0].ID)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string, even
// with a nil option.
func (s *APIKeyTestSuite) TestListWithoutOptions() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery, "an unfiltered list must not send an empty query string")
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background(), nil)

	s.Require().NoError(err)
	s.Require().NotNil(page)
	s.Empty(page.Items)
}

// TestListOptions verifies each documented list query parameter is rendered by
// exactly one option, under the key and value the API expects.
func (s *APIKeyTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option apikey.ListOption
		key    string
		value  string
	}{
		{"limit", apikey.WithLimit(50), "limit", "50"},
		{"starting after", apikey.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
	}

	s.Require().Len(tests, 2, "every documented list query parameter needs an option")

	for _, test := range tests {
		s.Run(test.name, func() {
			q := instantly.NewQuery()
			test.option(q)

			s.Require().Equal(1, q.Len(), "an option must render exactly one query parameter")
			s.Equal(test.value, q.Get(test.key))
		})
	}
}

// TestDelete verifies a delete returns the removed key and escapes the id.
func (s *APIKeyTestSuite) TestDelete() {
	s.Router.Delete(idPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("key-1", instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(keyFixture))
	})

	key, err := s.svc().Delete(context.Background(), "key-1")

	s.Require().NoError(err)
	s.Equal("key-1", key.ID)
}

// TestDeletePathEscape verifies an id carrying reserved characters is escaped
// into the request path rather than altering it.
func (s *APIKeyTestSuite) TestDeletePathEscape() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, keyFixture), nil
			},
		)},
	))

	_, err := apikey.New(client).Delete(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/api-keys/..%2Fadmin%3Fx=1", requestURI)
}

// TestFailures verifies every endpoint surfaces the documented API error.
func (s *APIKeyTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "create", Method: http.MethodPost, Path: listPath, Status: http.StatusBadRequest,
			Call: func() error { _, err := svc.Create(ctx, apikey.CreateRequest{}); return err },
		},
		{
			Name: "list", Method: http.MethodGet, Path: listPath, Status: http.StatusTooManyRequests,
			Call: func() error { _, err := svc.List(ctx); return err },
		},
		{
			Name: "delete", Method: http.MethodDelete, Path: idPath, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Delete(ctx, "key-1"); return err },
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor parses a valid
// timestamp and reports an error for an unparseable one.
func (s *APIKeyTestSuite) TestParsedTimestampCreated() {
	got, err := (&apikey.APIKey{TimestampCreated: "2026-08-01T10:00:00.000Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&apikey.APIKey{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// TestScopes verifies a sample of the named scope constants map to the exact
// wire strings the API documents.
func (s *APIKeyTestSuite) TestScopes() {
	s.Equal("all:all", string(apikey.ScopeAllAll))
	s.Equal("campaigns:read", string(apikey.ScopeCampaignsRead))
	s.Equal("api_keys:delete", string(apikey.ScopeAPIKeysDelete))
	s.Equal("crm_actions:read", string(apikey.ScopeCRMActionsRead))
	s.Equal("dfy_email_account_orders:create", string(apikey.ScopeDFYEmailAccountOrdersCreate))
	s.Equal("ai_sdr_replies:update", string(apikey.ScopeAISDRRepliesUpdate))
	s.Equal("block_list_entries:all", string(apikey.ScopeBlockListEntriesAll))
}

// svc builds an API Key service pointed at the suite's mock client.
func (s *APIKeyTestSuite) svc() *apikey.Service {
	return apikey.New(s.Client)
}
