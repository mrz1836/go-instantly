package workspacemember_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/workspacemember"
)

// Router patterns and identifiers the workspace-member endpoints are exercised
// with. The patterns carry the full request path, including the /api/v2 prefix.
const (
	// listPath is the list/collection endpoint.
	listPath = "/api/v2/workspace-members"

	// idPattern is the router pattern for the single-member endpoints.
	idPattern = "/api/v2/workspace-members/:id"

	// memberID identifies the member the single-member endpoints operate on.
	memberID = "wm-1"
)

// memberFixture is a spec-shaped workspace member with every documented field
// populated, including the nullable ones and the permissions.
const memberFixture = `{
	"id": "wm-1",
	"email": "member@example.com",
	"user_id": "user-1",
	"workspace_id": "ws-1",
	"role": "admin",
	"accepted": true,
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"name": {"first": "Jane", "last": "Doe"},
	"nickname": "JD",
	"user_email": "jane@example.com",
	"issuer_id": "user-0",
	"permissions": ["campaigns.view", "unibox.all"]
}`

// memberFixtureNulls is the same member with every nullable field explicitly
// null, so an absent value stays distinguishable from a zero value.
const memberFixtureNulls = `{
	"id": "wm-2",
	"email": "pending@example.com",
	"user_id": "user-2",
	"workspace_id": "ws-1",
	"role": "view",
	"accepted": false,
	"timestamp_created": "2026-08-01T11:00:00.000Z",
	"name": {"first": "", "last": ""},
	"nickname": null,
	"user_email": null,
	"issuer_id": null,
	"permissions": null
}`

// WorkspaceMemberTestSuite exercises the Workspace Member API service against the
// mock router.
type WorkspaceMemberTestSuite struct {
	instantlytest.Suite
}

// TestWorkspaceMemberSuite runs the Workspace Member API suite.
func TestWorkspaceMemberSuite(t *testing.T) {
	suite.Run(t, new(WorkspaceMemberTestSuite))
}

// TestCreate verifies the invite body reaches the API and the member decodes.
func (s *WorkspaceMemberTestSuite) TestCreate() {
	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		var received workspacemember.CreateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("member@example.com", received.Email)
		s.Equal(workspacemember.RoleAdmin, received.Role)
		s.Equal([]workspacemember.Permission{workspacemember.PermissionUniboxAll}, received.Permissions)

		_, _ = w.Write([]byte(memberFixture))
	})

	got, err := s.svc().Create(context.Background(), workspacemember.CreateRequest{
		Email:       "member@example.com",
		Role:        workspacemember.RoleAdmin,
		Permissions: []workspacemember.Permission{workspacemember.PermissionUniboxAll},
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(memberID, got.ID)
	s.Equal(workspacemember.RoleAdmin, got.Role)
}

// TestList verifies a page decodes, including the enum, name, permissions, and
// nullable-vs-zero fields.
func (s *WorkspaceMemberTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("50", req.URL.Query().Get("limit"))
		s.Equal("true", req.URL.Query().Get("accepted"))

		_, _ = w.Write([]byte(instantlytest.Page([]string{memberFixture, memberFixtureNulls}, "cursor-2")))
	})

	page, err := s.svc().List(context.Background(),
		workspacemember.WithLimit(50),
		workspacemember.WithAccepted(true),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)

	populated := page.Items[0]
	s.Equal(memberID, populated.ID)
	s.Equal(workspacemember.RoleAdmin, populated.Role)
	s.True(populated.Accepted)
	s.Equal("Jane", populated.Name.First)
	s.Equal(
		[]workspacemember.Permission{workspacemember.PermissionCampaignsView, workspacemember.PermissionUniboxAll},
		populated.Permissions,
	)
	s.Require().NotNil(populated.Nickname)
	s.Equal("JD", *populated.Nickname)

	// Nullable fields stay nil rather than collapsing to a zero value.
	bare := page.Items[1]
	s.False(bare.Accepted)
	s.Nil(bare.Nickname)
	s.Nil(bare.UserEmail)
	s.Nil(bare.IssuerID)
	s.Empty(bare.Permissions)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string, even
// with a nil option.
func (s *WorkspaceMemberTestSuite) TestListWithoutOptions() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery, "an unfiltered list must not send an empty query string")
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background(), nil)

	s.Require().NoError(err)
	s.Require().NotNil(page)
	s.Empty(page.Items)
}

// TestGet verifies a single member decodes.
func (s *WorkspaceMemberTestSuite) TestGet() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(memberID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(memberFixture))
	})

	got, err := s.svc().Get(context.Background(), memberID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("member@example.com", got.Email)
	s.Equal("Doe", got.Name.Last)
}

// TestUpdate verifies the patch body is sent and the updated member decodes.
func (s *WorkspaceMemberTestSuite) TestUpdate() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(memberID, instantlytest.PathParam(req, "id"))

		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("editor", received["role"])
		s.Equal("Boss", received["nickname"])

		_, _ = w.Write([]byte(memberFixture))
	})

	got, err := s.svc().Update(context.Background(), memberID, workspacemember.UpdateRequest{
		Role:     workspacemember.RoleEditor,
		Nickname: instantly.Ptr("Boss"),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(memberID, got.ID)
}

// TestUpdateOmitsUnsetFields verifies an empty patch sends no field at all.
func (s *WorkspaceMemberTestSuite) TestUpdateOmitsUnsetFields() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Empty(received, "an unset patch field must not be sent")

		_, _ = w.Write([]byte(memberFixture))
	})

	got, err := s.svc().Update(context.Background(), memberID, workspacemember.UpdateRequest{})

	s.Require().NoError(err)
	s.Require().NotNil(got)
}

// TestDelete verifies the removed member is returned to the caller.
func (s *WorkspaceMemberTestSuite) TestDelete() {
	s.Router.Delete(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(memberID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(memberFixture))
	})

	got, err := s.svc().Delete(context.Background(), memberID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(memberID, got.ID)
}

// TestPathParametersAreEscaped verifies a caller-supplied identifier cannot
// rewrite the request path.
func (s *WorkspaceMemberTestSuite) TestPathParametersAreEscaped() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, memberFixture), nil
			},
		)},
	))

	_, err := workspacemember.New(client).Get(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/workspace-members/..%2Fadmin%3Fx=1", requestURI)
}

// TestListOptions verifies each documented list query parameter is rendered by
// exactly one option, under the key and value the API expects.
func (s *WorkspaceMemberTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option workspacemember.ListOption
		key    string
		value  string
	}{
		{"limit", workspacemember.WithLimit(50), "limit", "50"},
		{"starting after", workspacemember.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"accepted", workspacemember.WithAccepted(false), "accepted", "false"},
		{"search", workspacemember.WithSearch("jane"), "search", "jane"},
	}

	s.Require().Len(tests, 4, "every documented list query parameter needs an option")

	for _, test := range tests {
		s.Run(test.name, func() {
			q := instantly.NewQuery()
			test.option(q)

			s.Require().Equal(1, q.Len(), "an option must render exactly one query parameter")
			s.Equal(test.value, q.Get(test.key))
		})
	}
}

// TestFailures verifies every endpoint surfaces the documented API error.
func (s *WorkspaceMemberTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "create", Method: http.MethodPost, Path: listPath, Status: http.StatusPaymentRequired,
			Call: func() error { _, err := svc.Create(ctx, workspacemember.CreateRequest{}); return err },
		},
		{
			Name: "list", Method: http.MethodGet, Path: listPath, Status: http.StatusTooManyRequests,
			Call: func() error { _, err := svc.List(ctx); return err },
		},
		{
			Name: "get", Method: http.MethodGet, Path: idPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Get(ctx, "missing"); return err },
		},
		{
			Name: "update", Method: http.MethodPatch, Path: idPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Update(ctx, "missing", workspacemember.UpdateRequest{}); return err },
		},
		{
			Name: "delete", Method: http.MethodDelete, Path: idPattern, Status: http.StatusForbidden,
			Call: func() error { _, err := svc.Delete(ctx, "missing"); return err },
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor parses a valid
// timestamp and reports an error for an unparseable one.
func (s *WorkspaceMemberTestSuite) TestParsedTimestampCreated() {
	got, err := (&workspacemember.Member{TimestampCreated: "2026-08-01T10:00:00.000Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&workspacemember.Member{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// svc builds a Workspace Member service pointed at the suite's mock client.
func (s *WorkspaceMemberTestSuite) svc() *workspacemember.Service {
	return workspacemember.New(s.Client)
}
