package lead_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/lead"
)

// Router patterns and identifiers the lead endpoints are exercised with.
const (
	collectionPath = "/api/v2/leads"
	listPath       = "/api/v2/leads/list"
	addPath        = "/api/v2/leads/add"
	bulkAssignPath = "/api/v2/leads/bulk-assign"
	mergePath      = "/api/v2/leads/merge"
	interestPath   = "/api/v2/leads/update-interest-status"
	subRemovePath  = "/api/v2/leads/subsequence/remove"
	subMovePath    = "/api/v2/leads/subsequence/move"
	movePath       = "/api/v2/leads/move"
	idPattern      = "/api/v2/leads/:id"

	testID = "lead-uuid-1"
)

// leadFixture is a spec-shaped lead with the required fields plus a
// representative set of populated optional fields.
const leadFixture = `{
	"id": "lead-uuid-1",
	"organization": "org-1",
	"company_domain": "example.com",
	"status": 1,
	"lt_interest_status": 1,
	"verification_status": 1,
	"enrichment_status": 1,
	"esg_code": 0,
	"esp_code": 2,
	"email_open_count": 3,
	"email_reply_count": 1,
	"email_click_count": 2,
	"upload_method": "manual",
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"timestamp_updated": "2026-08-01T11:00:00.000Z",
	"email": "lead@example.com",
	"first_name": "Jane",
	"campaign": "campaign-1",
	"is_website_visitor": true,
	"email_opened_step": 1,
	"status_summary": {"sent": 3},
	"payload": {"custom": "value"}
}`

// leadFixtureNulls has the nullable fields explicitly null.
const leadFixtureNulls = `{
	"id": "lead-uuid-2",
	"organization": "org-1",
	"company_domain": "example.org",
	"status": 2,
	"lt_interest_status": 0,
	"verification_status": 1,
	"enrichment_status": 0,
	"esg_code": 0,
	"esp_code": 0,
	"email_open_count": 0,
	"email_reply_count": 0,
	"email_click_count": 0,
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"timestamp_updated": "2026-08-01T11:00:00.000Z",
	"email": null,
	"first_name": null,
	"campaign": null,
	"is_website_visitor": null,
	"email_opened_step": null,
	"status_summary": {}
}`

// LeadTestSuite exercises the Lead API service against the mock router.
type LeadTestSuite struct {
	instantlytest.Suite
}

// TestLeadSuite runs the Lead API suite.
func TestLeadSuite(t *testing.T) {
	suite.Run(t, new(LeadTestSuite))
}

// TestCreate verifies the create body reaches the API and the lead decodes.
func (s *LeadTestSuite) TestCreate() {
	s.Router.Post(collectionPath, func(w http.ResponseWriter, req *http.Request) {
		var received lead.CreateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		if s.NotNil(received.Email) {
			s.Equal("lead@example.com", *received.Email)
		}

		_, _ = w.Write([]byte(leadFixture))
	})

	got, err := s.svc().Create(context.Background(), lead.CreateRequest{
		Email:            instantly.Ptr("lead@example.com"),
		FirstName:        instantly.Ptr("Jane"),
		LtInterestStatus: instantly.Ptr(lead.InterestInterested),
	})

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
	s.Equal(lead.StatusActive, got.Status)
	s.Equal(lead.InterestInterested, got.LtInterestStatus)
}

// TestList verifies the POST-list body carries the filters and the page decodes.
func (s *LeadTestSuite) TestList() {
	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodPost, req.Method, "listing leads is a POST")

		var received lead.ListRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("campaign-1", received.Campaign)
		s.Equal(50, received.Limit)

		_, _ = fmt.Fprintf(w, `{"items":[%s,%s],"next_starting_after":"cursor-2"}`, leadFixture, leadFixtureNulls)
	})

	page, err := s.svc().List(context.Background(), lead.ListRequest{Campaign: "campaign-1", Limit: 50})

	s.Require().NoError(err)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)

	populated := page.Items[0]
	s.Require().NotNil(populated.Email)
	s.Equal("lead@example.com", *populated.Email)
	s.InDelta(3, populated.EmailOpenCount, 0)
	s.JSONEq(`{"custom":"value"}`, string(populated.Payload))

	// Nullable fields stay nil rather than collapsing to a zero value.
	bare := page.Items[1]
	s.Nil(bare.Email)
	s.Nil(bare.Campaign)
	s.Nil(bare.IsWebsiteVisitor)
	s.Nil(bare.EmailOpenedStep)
}

// TestGet verifies a single lead decodes.
func (s *LeadTestSuite) TestGet() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(leadFixture))
	})

	got, err := s.svc().Get(context.Background(), testID)

	s.Require().NoError(err)
	s.Equal("example.com", got.CompanyDomain)
	s.JSONEq(`{"sent":3}`, string(got.StatusSummary))
}

// TestUpdate verifies the patch body is sent and the lead decodes.
func (s *LeadTestSuite) TestUpdate() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("Janet", received["first_name"])

		_, _ = w.Write([]byte(leadFixture))
	})

	got, err := s.svc().Update(context.Background(), testID, lead.UpdateRequest{
		FirstName: instantly.Ptr("Janet"),
	})

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestUpdateOmitsUnsetFields verifies an empty patch sends no field.
func (s *LeadTestSuite) TestUpdateOmitsUnsetFields() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Empty(received)

		_, _ = w.Write([]byte(leadFixture))
	})

	got, err := s.svc().Update(context.Background(), testID, lead.UpdateRequest{})

	s.Require().NoError(err)
	s.NotNil(got)
}

// TestDelete verifies the deleted lead is returned to the caller.
func (s *LeadTestSuite) TestDelete() {
	s.Router.Delete(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(leadFixture))
	})

	got, err := s.svc().Delete(context.Background(), testID)

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestListIter verifies the iterator stitches pages together, carries the
// caller's filters, and overrides the cursor on each page.
func (s *LeadTestSuite) TestListIter() {
	var requests atomic.Int64
	campaigns := make([]string, 0, 2)
	cursors := make([]string, 0, 2)

	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		var received lead.ListRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		campaigns = append(campaigns, received.Campaign)
		cursors = append(cursors, received.StartingAfter)

		if received.StartingAfter == "" {
			_, _ = fmt.Fprint(w, leadPage([]string{"l1", "l2"}, "cursor-2"))
			return
		}

		_, _ = fmt.Fprint(w, leadPage([]string{"l3"}, ""))
	})

	seen := make([]string, 0, 3)
	for got, err := range s.svc().ListIter(context.Background(), lead.ListRequest{Campaign: "campaign-1"}) {
		s.Require().NoError(err)
		seen = append(seen, got.ID)
	}

	s.Equal([]string{"l1", "l2", "l3"}, seen)
	s.Equal(int64(2), requests.Load())
	s.Equal([]string{"campaign-1", "campaign-1"}, campaigns, "the caller's filters survive every page")
	s.Equal([]string{"", "cursor-2"}, cursors, "the page cursor is sent in the body")
}

// TestListIterStopsOnError verifies a failure ends the iteration with a nil lead.
func (s *LeadTestSuite) TestListIterStopsOnError() {
	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		var received lead.ListRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))

		if received.StartingAfter == "" {
			_, _ = fmt.Fprint(w, leadPage([]string{"l1"}, "cursor-2"))
			return
		}
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "slow")
	})

	seen := make([]string, 0, 1)
	var iterErr error
	for got, err := range s.svc().ListIter(context.Background(), lead.ListRequest{}) {
		if err != nil {
			iterErr = err
			s.Nil(got)
			break
		}
		seen = append(seen, got.ID)
	}

	s.Equal([]string{"l1"}, seen)
	instantlytest.AssertAPIError(s.T(), iterErr, http.StatusTooManyRequests)
}

// TestPathParametersAreEscaped verifies a caller-supplied id cannot rewrite the
// request path.
func (s *LeadTestSuite) TestPathParametersAreEscaped() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, leadFixture), nil
			},
		)},
	))

	_, err := lead.New(client).Get(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/leads/..%2Fadmin%3Fx=1", requestURI)
}

// TestFailures verifies every endpoint surfaces the documented API error.
func (s *LeadTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "create", Method: http.MethodPost, Path: collectionPath, Status: http.StatusPaymentRequired,
			Call: func() error { _, err := svc.Create(ctx, lead.CreateRequest{}); return err },
		},
		{
			Name: "list", Method: http.MethodPost, Path: listPath, Status: http.StatusTooManyRequests,
			Call: func() error { _, err := svc.List(ctx, lead.ListRequest{}); return err },
		},
		{
			Name: "get", Method: http.MethodGet, Path: idPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Get(ctx, "missing"); return err },
		},
		{
			Name: "update", Method: http.MethodPatch, Path: idPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Update(ctx, "missing", lead.UpdateRequest{}); return err },
		},
		{
			Name: "delete", Method: http.MethodDelete, Path: idPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Delete(ctx, "missing"); return err },
		},
		{
			Name: "bulkDelete", Method: http.MethodDelete, Path: collectionPath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.BulkDelete(ctx, lead.BulkDeleteRequest{}); return err },
		},
		{
			Name: "bulkAdd", Method: http.MethodPost, Path: addPath, Status: http.StatusPaymentRequired,
			Call: func() error { _, err := svc.BulkAdd(ctx, lead.BulkAddRequest{}); return err },
		},
		{
			Name: "bulkAssign", Method: http.MethodPost, Path: bulkAssignPath, Status: http.StatusForbidden,
			Call: func() error { return svc.BulkAssign(ctx, lead.BulkAssignRequest{}) },
		},
		{
			Name: "merge", Method: http.MethodPost, Path: mergePath, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Merge(ctx, lead.MergeRequest{}); return err },
		},
		{
			Name: "updateInterestStatus", Method: http.MethodPost, Path: interestPath, Status: http.StatusNotFound,
			Call: func() error { return svc.UpdateInterestStatus(ctx, lead.UpdateInterestStatusRequest{}) },
		},
		{
			Name: "removeFromSubsequence", Method: http.MethodPost, Path: subRemovePath, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.RemoveFromSubsequence(ctx, lead.SubsequenceRemoveRequest{}); return err },
		},
		{
			Name: "moveToSubsequence", Method: http.MethodPost, Path: subMovePath, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.MoveToSubsequence(ctx, lead.SubsequenceMoveRequest{}); return err },
		},
		{
			Name: "move", Method: http.MethodPost, Path: movePath, Status: http.StatusForbidden,
			Call: func() error { _, err := svc.Move(ctx, lead.MoveRequest{}); return err },
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor parses a valid
// timestamp and reports an error for an unparseable one.
func (s *LeadTestSuite) TestParsedTimestampCreated() {
	got, err := (&lead.Lead{TimestampCreated: "2026-08-01T10:00:00.000Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&lead.Lead{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// svc builds a Lead service pointed at the suite's mock client.
func (s *LeadTestSuite) svc() *lead.Service {
	return lead.New(s.Client)
}

// leadPage renders one page of a list response for the given lead ids.
func leadPage(ids []string, nextCursor string) string {
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, fmt.Sprintf(
			`{"id":%q,"organization":"org-1","company_domain":"example.com","status":1,`+
				`"lt_interest_status":1,"verification_status":1,"enrichment_status":1,"esg_code":0,`+
				`"esp_code":2,"email_open_count":0,"email_reply_count":0,"email_click_count":0,`+
				`"timestamp_created":"2026-08-01T10:00:00.000Z",`+
				`"timestamp_updated":"2026-08-01T11:00:00.000Z","status_summary":{}}`,
			id,
		))
	}

	return instantlytest.Page(items, nextCursor)
}
