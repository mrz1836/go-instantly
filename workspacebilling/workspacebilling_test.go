package workspacebilling_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/workspacebilling"
)

// Router patterns the workspace-billing endpoints are exercised with. The
// patterns carry the full request path, including the /api/v2 prefix.
const (
	// planPath is the plan-details endpoint.
	planPath = "/api/v2/workspace-billing/plan-details"

	// subscriptionPath is the subscription-details endpoint.
	subscriptionPath = "/api/v2/workspace-billing/subscription-details"
)

// planFixture is a spec-shaped plan-details response with the raw addon and
// subscription payloads preserved.
const planFixture = `{
	"organization_id": "org-1",
	"organization_name": "Acme",
	"addons": {"verification": {"quantity": 5}},
	"subscriptions": {"main": {"plan_id": "plan-1", "status": "active"}}
}`

// subscriptionFixture is a spec-shaped subscription-details response with a raw
// subscription entry preserved.
const subscriptionFixture = `{
	"all_subs_cancelled": false,
	"subscriptions": [{"id": "sub-1", "status": "active", "quantity": 3}]
}`

// WorkspaceBillingTestSuite exercises the Workspace Billing API service against
// the mock router.
type WorkspaceBillingTestSuite struct {
	instantlytest.Suite
}

// TestWorkspaceBillingSuite runs the Workspace Billing API suite.
func TestWorkspaceBillingSuite(t *testing.T) {
	suite.Run(t, new(WorkspaceBillingTestSuite))
}

// TestPlanDetails verifies the plan details decode, with the raw payloads
// preserved verbatim.
func (s *WorkspaceBillingTestSuite) TestPlanDetails() {
	s.Router.Get(planPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodGet, req.Method)
		s.Equal(planPath, req.URL.Path)
		_, _ = w.Write([]byte(planFixture))
	})

	got, err := s.svc().PlanDetails(context.Background())

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("org-1", got.OrganizationID)
	s.Equal("Acme", got.OrganizationName)
	s.JSONEq(`{"verification":{"quantity":5}}`, string(got.Addons))
	s.JSONEq(`{"main":{"plan_id":"plan-1","status":"active"}}`, string(got.Subscriptions))
}

// TestSubscriptionDetails verifies the subscription details decode, with each
// subscription preserved verbatim.
func (s *WorkspaceBillingTestSuite) TestSubscriptionDetails() {
	s.Router.Get(subscriptionPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(subscriptionPath, req.URL.Path)
		_, _ = w.Write([]byte(subscriptionFixture))
	})

	got, err := s.svc().SubscriptionDetails(context.Background())

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.False(got.AllSubsCancelled)
	s.Require().Len(got.Subscriptions, 1)
	s.JSONEq(`{"id":"sub-1","status":"active","quantity":3}`, string(got.Subscriptions[0]))
}

// TestFailures verifies every endpoint surfaces the documented API error.
func (s *WorkspaceBillingTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "planDetails", Method: http.MethodGet, Path: planPath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.PlanDetails(ctx); return err },
		},
		{
			Name: "subscriptionDetails", Method: http.MethodGet, Path: subscriptionPath, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.SubscriptionDetails(ctx); return err },
		},
	})
}

// svc builds a Workspace Billing service pointed at the suite's mock client.
func (s *WorkspaceBillingTestSuite) svc() *workspacebilling.Service {
	return workspacebilling.New(s.Client)
}
