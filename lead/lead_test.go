package lead_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

// TestCreateFailure verifies a rejected create returns no lead.
func (s *LeadTestSuite) TestCreateFailure() {
	s.Router.Post(collectionPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusPaymentRequired, "Payment Required", "limit")
	})

	got, err := s.svc().Create(context.Background(), lead.CreateRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusPaymentRequired)
	s.Nil(got)
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

// TestListFailure verifies a failed list returns no page.
func (s *LeadTestSuite) TestListFailure() {
	s.Router.Post(listPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "slow")
	})

	page, err := s.svc().List(context.Background(), lead.ListRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusTooManyRequests)
	s.Nil(page)
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

// TestGetFailure verifies a missing lead returns no value.
func (s *LeadTestSuite) TestGetFailure() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no lead")
	})

	got, err := s.svc().Get(context.Background(), "missing")

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
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

// TestUpdateFailure verifies a failed patch returns no value.
func (s *LeadTestSuite) TestUpdateFailure() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no lead")
	})

	got, err := s.svc().Update(context.Background(), "missing", lead.UpdateRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
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

// TestDeleteFailure verifies a failed delete returns no value.
func (s *LeadTestSuite) TestDeleteFailure() {
	s.Router.Delete(idPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no lead")
	})

	got, err := s.svc().Delete(context.Background(), "missing")

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
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

	if nextCursor == "" {
		return fmt.Sprintf(`{"items":[%s]}`, strings.Join(items, ","))
	}

	return fmt.Sprintf(`{"items":[%s],"next_starting_after":%q}`, strings.Join(items, ","), nextCursor)
}
