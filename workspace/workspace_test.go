package workspace_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/workspace"
)

// Router patterns the workspace endpoints are exercised with. Every endpoint
// operates on the singleton /workspaces/current, so none carry a path parameter.
const (
	// currentPath is the get/patch workspace endpoint.
	currentPath = "/api/v2/workspaces/current"

	// schedulePath is the schedule-for-removal endpoint (POST and DELETE).
	schedulePath = "/api/v2/workspaces/current/schedule-for-removal"

	// domainPath is the whitelabel-domain endpoint (POST, GET, and DELETE).
	domainPath = "/api/v2/workspaces/current/whitelabel-domain"

	// changeOwnerPath is the change-owner endpoint.
	changeOwnerPath = "/api/v2/workspaces/current/change-owner"
)

// workspaceFixture is a spec-shaped workspace with every documented field
// populated, including the nullable ones and the raw verification plan.
const workspaceFixture = `{
	"id": "ws-1",
	"name": "My Workspace",
	"owner": "user-1",
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"timestamp_updated": "2026-08-01T11:00:00.000Z",
	"org_logo_url": "https://example.com/logo.png",
	"org_client_domain": "client.example.com",
	"add_unsub_to_block": true,
	"default_opportunity_value": 500,
	"scheduled_for_removal_at": "2026-09-01T00:00:00.000Z",
	"plan_id": "plan-1",
	"plan_id_bundle": "bundle-1",
	"plan_id_crm": "crm-1",
	"plan_id_inbox_placement": "ip-1",
	"plan_id_leadfinder": "lf-1",
	"plan_id_website_visitor": "wv-1",
	"plan_id_verification": {"product_id": "v-1", "quantity": 2, "timestamp_updated": "2026-08-01"}
}`

// workspaceFixtureNulls is the same workspace with every nullable field
// explicitly null, so an absent value stays distinguishable from a zero value.
const workspaceFixtureNulls = `{
	"id": "ws-2",
	"name": "Bare Workspace",
	"owner": "user-1",
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"timestamp_updated": "2026-08-01T11:00:00.000Z",
	"org_logo_url": null,
	"org_client_domain": null,
	"add_unsub_to_block": null,
	"default_opportunity_value": null,
	"scheduled_for_removal_at": null,
	"plan_id": null,
	"plan_id_bundle": null,
	"plan_id_crm": null,
	"plan_id_inbox_placement": null,
	"plan_id_leadfinder": null,
	"plan_id_website_visitor": null
}`

// WorkspaceTestSuite exercises the Workspace API service against the mock router.
type WorkspaceTestSuite struct {
	instantlytest.Suite
}

// TestWorkspaceSuite runs the Workspace API suite.
func TestWorkspaceSuite(t *testing.T) {
	suite.Run(t, new(WorkspaceTestSuite))
}

// TestGet verifies the current workspace decodes, including its nullable fields
// and the raw verification plan.
func (s *WorkspaceTestSuite) TestGet() {
	s.Router.Get(currentPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodGet, req.Method)
		s.Equal(currentPath, req.URL.Path)
		_, _ = w.Write([]byte(workspaceFixture))
	})

	got, err := s.svc().Get(context.Background())

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("ws-1", got.ID)
	s.Equal("My Workspace", got.Name)
	s.Require().NotNil(got.OrgLogoURL)
	s.Equal("https://example.com/logo.png", *got.OrgLogoURL)
	s.Require().NotNil(got.DefaultOpportunityValue)
	s.InDelta(500, *got.DefaultOpportunityValue, 0)
	s.Require().NotNil(got.ScheduledForRemovalAt)
	s.JSONEq(`{"product_id":"v-1","quantity":2,"timestamp_updated":"2026-08-01"}`, string(got.PlanIDVerification))
}

// TestGetNullable verifies the nullable fields stay nil rather than collapsing to
// a zero value.
func (s *WorkspaceTestSuite) TestGetNullable() {
	s.Router.Get(currentPath, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(workspaceFixtureNulls))
	})

	got, err := s.svc().Get(context.Background())

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Nil(got.OrgLogoURL)
	s.Nil(got.OrgClientDomain)
	s.Nil(got.AddUnsubToBlock)
	s.Nil(got.DefaultOpportunityValue)
	s.Nil(got.ScheduledForRemovalAt)
	s.Nil(got.PlanID)
	s.Empty(got.PlanIDVerification)
}

// TestUpdate verifies the patch body is sent and the updated workspace decodes.
func (s *WorkspaceTestSuite) TestUpdate() {
	s.Router.Patch(currentPath, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("Renamed", received["name"])
		s.Equal("https://example.com/new.png", received["org_logo_url"])

		_, _ = w.Write([]byte(workspaceFixture))
	})

	got, err := s.svc().Update(context.Background(), workspace.UpdateRequest{
		Name:       "Renamed",
		OrgLogoURL: instantly.Ptr("https://example.com/new.png"),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("ws-1", got.ID)
}

// TestUpdateOmitsUnsetFields verifies an empty patch sends no field at all.
func (s *WorkspaceTestSuite) TestUpdateOmitsUnsetFields() {
	s.Router.Patch(currentPath, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Empty(received, "an unset patch field must not be sent")

		_, _ = w.Write([]byte(workspaceFixture))
	})

	got, err := s.svc().Update(context.Background(), workspace.UpdateRequest{})

	s.Require().NoError(err)
	s.Require().NotNil(got)
}

// TestScheduleRemoval verifies the schedule endpoint posts no body and returns
// the workspace.
func (s *WorkspaceTestSuite) TestScheduleRemoval() {
	s.Router.Post(schedulePath, func(w http.ResponseWriter, req *http.Request) {
		body, err := instantlytest.ReadAll(req)
		s.NoError(err)
		s.Empty(body, "scheduling removal sends no request body")

		_, _ = w.Write([]byte(workspaceFixture))
	})

	got, err := s.svc().ScheduleRemoval(context.Background())

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("ws-1", got.ID)
}

// TestCancelRemoval verifies the cancel endpoint uses DELETE and returns the
// workspace.
func (s *WorkspaceTestSuite) TestCancelRemoval() {
	s.Router.Delete(schedulePath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodDelete, req.Method)
		_, _ = w.Write([]byte(workspaceFixture))
	})

	got, err := s.svc().CancelRemoval(context.Background())

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("ws-1", got.ID)
}

// TestSetAgencyDomain verifies the domain body is sent and the workspace decodes.
func (s *WorkspaceTestSuite) TestSetAgencyDomain() {
	s.Router.Post(domainPath, func(w http.ResponseWriter, req *http.Request) {
		var received workspace.SetDomainRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("agency.example.com", received.Domain)

		_, _ = w.Write([]byte(workspaceFixture))
	})

	got, err := s.svc().SetAgencyDomain(context.Background(), workspace.SetDomainRequest{
		Domain: "agency.example.com",
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("ws-1", got.ID)
}

// TestDomainInfo verifies the whitelabel domain status decodes, including the
// verification records.
func (s *WorkspaceTestSuite) TestDomainInfo() {
	s.Router.Get(domainPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodGet, req.Method)
		_, _ = w.Write([]byte(
			`{"name":"agency.example.com","verified":false,"verification":[` +
				`{"domain":"agency.example.com","type":"CNAME","value":"target.instantly.ai","reason":"ownership"}]}`,
		))
	})

	got, err := s.svc().DomainInfo(context.Background())

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("agency.example.com", got.Name)
	s.False(got.Verified)
	s.Require().Len(got.Verification, 1)
	s.Equal("CNAME", got.Verification[0].Type)
	s.Equal("target.instantly.ai", got.Verification[0].Value)
}

// TestDeleteAgencyDomain verifies the delete endpoint uses DELETE and returns the
// workspace.
func (s *WorkspaceTestSuite) TestDeleteAgencyDomain() {
	s.Router.Delete(domainPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodDelete, req.Method)
		_, _ = w.Write([]byte(workspaceFixture))
	})

	got, err := s.svc().DeleteAgencyDomain(context.Background())

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("ws-1", got.ID)
}

// TestChangeOwner verifies the change-owner body is sent and the workspace
// decodes.
func (s *WorkspaceTestSuite) TestChangeOwner() {
	s.Router.Post(changeOwnerPath, func(w http.ResponseWriter, req *http.Request) {
		var received workspace.ChangeOwnerRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("new@example.com", received.Email)
		s.Equal("token-123", received.Sec)

		_, _ = w.Write([]byte(workspaceFixture))
	})

	got, err := s.svc().ChangeOwner(context.Background(), workspace.ChangeOwnerRequest{
		Email: "new@example.com",
		Sec:   "token-123",
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("ws-1", got.ID)
}

// TestFailures verifies every endpoint surfaces the documented API error.
func (s *WorkspaceTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "get", Method: http.MethodGet, Path: currentPath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.Get(ctx); return err },
		},
		{
			Name: "update", Method: http.MethodPatch, Path: currentPath, Status: http.StatusForbidden,
			Call: func() error { _, err := svc.Update(ctx, workspace.UpdateRequest{}); return err },
		},
		{
			Name: "scheduleRemoval", Method: http.MethodPost, Path: schedulePath, Status: http.StatusForbidden,
			Call: func() error { _, err := svc.ScheduleRemoval(ctx); return err },
		},
		{
			Name: "cancelRemoval", Method: http.MethodDelete, Path: schedulePath, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.CancelRemoval(ctx); return err },
		},
		{
			Name: "setAgencyDomain", Method: http.MethodPost, Path: domainPath, Status: http.StatusPaymentRequired,
			Call: func() error { _, err := svc.SetAgencyDomain(ctx, workspace.SetDomainRequest{}); return err },
		},
		{
			Name: "domainInfo", Method: http.MethodGet, Path: domainPath, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.DomainInfo(ctx); return err },
		},
		{
			Name: "deleteAgencyDomain", Method: http.MethodDelete, Path: domainPath, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.DeleteAgencyDomain(ctx); return err },
		},
		{
			Name: "changeOwner", Method: http.MethodPost, Path: changeOwnerPath, Status: http.StatusForbidden,
			Call: func() error { _, err := svc.ChangeOwner(ctx, workspace.ChangeOwnerRequest{}); return err },
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor parses a valid
// timestamp and reports an error for an unparseable one.
func (s *WorkspaceTestSuite) TestParsedTimestampCreated() {
	got, err := (&workspace.Workspace{TimestampCreated: "2026-08-01T10:00:00.000Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&workspace.Workspace{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// svc builds a Workspace service pointed at the suite's mock client.
func (s *WorkspaceTestSuite) svc() *workspace.Service {
	return workspace.New(s.Client)
}
