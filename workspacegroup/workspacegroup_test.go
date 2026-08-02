package workspacegroup_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/workspacegroup"
)

// Router patterns and identifiers the workspace-group-member endpoints are
// exercised with. The patterns carry the full request path, including the
// /api/v2 prefix.
const (
	// listPath is the list/collection endpoint.
	listPath = "/api/v2/workspace-group-members"

	// idPattern is the router pattern for the single-member endpoints.
	idPattern = "/api/v2/workspace-group-members/:id"

	// adminPath is the admin-workspace endpoint.
	adminPath = "/api/v2/workspace-group-members/admin"

	// memberID identifies the member the single-member endpoints operate on.
	memberID = "wgm-1"
)

// memberFixture is a spec-shaped workspace group member with every documented
// field populated, including the nullable names.
const memberFixture = `{
	"id": "wgm-1",
	"admin_workspace_id": "ws-admin",
	"sub_workspace_id": "ws-sub",
	"status": "accepted",
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"timestamp_updated": "2026-08-01T11:00:00.000Z",
	"admin_workspace_name": "Agency HQ",
	"sub_workspace_name": "Client A"
}`

// memberFixtureNulls is the same member with every nullable field explicitly
// null, so an absent value stays distinguishable from a zero value.
const memberFixtureNulls = `{
	"id": "wgm-2",
	"admin_workspace_id": "ws-admin",
	"sub_workspace_id": "ws-sub-2",
	"status": "pending",
	"timestamp_created": "2026-08-01T12:00:00.000Z",
	"timestamp_updated": "2026-08-01T12:00:00.000Z",
	"admin_workspace_name": null,
	"sub_workspace_name": null
}`

// WorkspaceGroupTestSuite exercises the Workspace Group Member API service
// against the mock router.
type WorkspaceGroupTestSuite struct {
	instantlytest.Suite
}

// TestWorkspaceGroupSuite runs the Workspace Group Member API suite.
func TestWorkspaceGroupSuite(t *testing.T) {
	suite.Run(t, new(WorkspaceGroupTestSuite))
}

// TestCreate verifies the invite body reaches the API and the member decodes.
func (s *WorkspaceGroupTestSuite) TestCreate() {
	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		var received workspacegroup.CreateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("ws-sub", received.SubWorkspaceID)

		_, _ = w.Write([]byte(memberFixture))
	})

	got, err := s.svc().Create(context.Background(), workspacegroup.CreateRequest{
		SubWorkspaceID: "ws-sub",
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(memberID, got.ID)
	s.Equal(workspacegroup.StatusAccepted, got.Status)
}

// TestList verifies a page decodes, including the enum and nullable-vs-zero
// fields.
func (s *WorkspaceGroupTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("50", req.URL.Query().Get("limit"))

		_, _ = w.Write([]byte(instantlytest.Page([]string{memberFixture, memberFixtureNulls}, "cursor-2")))
	})

	page, err := s.svc().List(context.Background(), workspacegroup.WithLimit(50))

	s.Require().NoError(err)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)

	populated := page.Items[0]
	s.Equal(memberID, populated.ID)
	s.Equal(workspacegroup.StatusAccepted, populated.Status)
	s.Require().NotNil(populated.AdminWorkspaceName)
	s.Equal("Agency HQ", *populated.AdminWorkspaceName)

	// Nullable fields stay nil rather than collapsing to a zero value.
	bare := page.Items[1]
	s.Equal(workspacegroup.StatusPending, bare.Status)
	s.Nil(bare.AdminWorkspaceName)
	s.Nil(bare.SubWorkspaceName)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string, even
// with a nil option.
func (s *WorkspaceGroupTestSuite) TestListWithoutOptions() {
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
func (s *WorkspaceGroupTestSuite) TestGet() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(memberID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(memberFixture))
	})

	got, err := s.svc().Get(context.Background(), memberID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("ws-sub", got.SubWorkspaceID)
	s.Equal("ws-admin", got.AdminWorkspaceID)
}

// TestDelete verifies the removed member is returned to the caller.
func (s *WorkspaceGroupTestSuite) TestDelete() {
	s.Router.Delete(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(memberID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(memberFixture))
	})

	got, err := s.svc().Delete(context.Background(), memberID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(memberID, got.ID)
}

// TestAdmin verifies the admin workspace decodes, and that the exact route is not
// shadowed by the :id route.
func (s *WorkspaceGroupTestSuite) TestAdmin() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, _ *http.Request) {
		// Registered first, yet the exact admin route must win.
		w.WriteHeader(http.StatusInternalServerError)
	})
	s.Router.Get(adminPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(adminPath, req.URL.Path)
		_, _ = w.Write([]byte(
			`{"has_admin_workspace":true,"workspace_name":"Agency HQ","workspace_group_member_id":"wgm-1"}`,
		))
	})

	got, err := s.svc().Admin(context.Background())

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.True(got.HasAdminWorkspace)
	s.Equal("Agency HQ", got.WorkspaceName)
	s.Equal(memberID, got.WorkspaceGroupMemberID)
}

// TestPathParametersAreEscaped verifies a caller-supplied identifier cannot
// rewrite the request path.
func (s *WorkspaceGroupTestSuite) TestPathParametersAreEscaped() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, memberFixture), nil
			},
		)},
	))

	_, err := workspacegroup.New(client).Get(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/workspace-group-members/..%2Fadmin%3Fx=1", requestURI)
}

// TestListOptions verifies each documented list query parameter is rendered by
// exactly one option, under the key and value the API expects.
func (s *WorkspaceGroupTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option workspacegroup.ListOption
		key    string
		value  string
	}{
		{"limit", workspacegroup.WithLimit(50), "limit", "50"},
		{"starting after", workspacegroup.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
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

// TestFailures verifies every endpoint surfaces the documented API error.
func (s *WorkspaceGroupTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "create", Method: http.MethodPost, Path: listPath, Status: http.StatusPaymentRequired,
			Call: func() error { _, err := svc.Create(ctx, workspacegroup.CreateRequest{}); return err },
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
			Name: "delete", Method: http.MethodDelete, Path: idPattern, Status: http.StatusForbidden,
			Call: func() error { _, err := svc.Delete(ctx, "missing"); return err },
		},
		{
			Name: "admin", Method: http.MethodGet, Path: adminPath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.Admin(ctx); return err },
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor parses a valid
// timestamp and reports an error for an unparseable one.
func (s *WorkspaceGroupTestSuite) TestParsedTimestampCreated() {
	got, err := (&workspacegroup.Member{TimestampCreated: "2026-08-01T10:00:00.000Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&workspacegroup.Member{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// svc builds a Workspace Group Member service pointed at the suite's mock client.
func (s *WorkspaceGroupTestSuite) svc() *workspacegroup.Service {
	return workspacegroup.New(s.Client)
}
