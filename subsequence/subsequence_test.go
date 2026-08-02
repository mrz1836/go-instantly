package subsequence_test

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
	"github.com/mrz1836/go-instantly/subsequence"
)

// Router patterns and identifiers the subsequence endpoints are exercised with.
const (
	listPath      = "/api/v2/subsequences"
	idPattern     = "/api/v2/subsequences/:id"
	duplicatePatt = "/api/v2/subsequences/:id/duplicate"
	pausePatt     = "/api/v2/subsequences/:id/pause"
	resumePatt    = "/api/v2/subsequences/:id/resume"
	sendingPatt   = "/api/v2/subsequences/:id/sending-status"

	testID = "subseq-uuid-1"
)

// fixture is a spec-shaped subsequence with the required fields plus a
// representative set of populated optional fields.
const fixture = `{
	"id": "subseq-uuid-1",
	"name": "Follow up",
	"parent_campaign": "campaign-1",
	"workspace": "ws-1",
	"status": 1,
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"timestamp_leads_updated": "2026-08-01T11:00:00.000Z",
	"daily_limit": 50,
	"daily_limit_mode": "custom",
	"ignore_account_daily_limit": true,
	"conditions": {"trigger": "no_reply"},
	"subsequence_schedule": {"schedules": []},
	"sequences": [{"steps": []}]
}`

// fixtureNulls has the nullable daily_limit explicitly null.
const fixtureNulls = `{
	"id": "subseq-uuid-2",
	"name": "Bare",
	"parent_campaign": "campaign-1",
	"workspace": "ws-1",
	"status": 0,
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"timestamp_leads_updated": "2026-08-01T11:00:00.000Z",
	"daily_limit": null,
	"conditions": {},
	"subsequence_schedule": {},
	"sequences": []
}`

// SubsequenceTestSuite exercises the Campaign Subsequence API service.
type SubsequenceTestSuite struct {
	instantlytest.Suite
}

// TestSubsequenceSuite runs the Campaign Subsequence API suite.
func TestSubsequenceSuite(t *testing.T) {
	suite.Run(t, new(SubsequenceTestSuite))
}

// TestCreate verifies the create body reaches the API and the subsequence decodes.
func (s *SubsequenceTestSuite) TestCreate() {
	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		var received subsequence.CreateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("campaign-1", received.ParentCampaign)
		s.Equal("Follow up", received.Name)
		s.JSONEq(`{"trigger":"no_reply"}`, string(received.Conditions))

		_, _ = w.Write([]byte(fixture))
	})

	got, err := s.svc().Create(context.Background(), subsequence.CreateRequest{
		ParentCampaign:      "campaign-1",
		Name:                "Follow up",
		Conditions:          json.RawMessage(`{"trigger":"no_reply"}`),
		SubsequenceSchedule: json.RawMessage(`{"schedules":[]}`),
		Sequences:           json.RawMessage(`[{"steps":[]}]`),
		DailyLimitMode:      subsequence.DailyLimitCustom,
	})

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
	s.Equal(subsequence.StatusActive, got.Status)
	s.Equal(subsequence.DailyLimitCustom, got.DailyLimitMode)
}

// TestCreateFailure verifies a rejected create returns no subsequence.
func (s *SubsequenceTestSuite) TestCreateFailure() {
	s.Router.Post(listPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusPaymentRequired, "Payment Required", "limit")
	})

	got, err := s.svc().Create(context.Background(), subsequence.CreateRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusPaymentRequired)
	s.Nil(got)
}

// TestList verifies a page decodes, including nullable-vs-zero fields.
func (s *SubsequenceTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("25", req.URL.Query().Get("limit"))
		s.Equal("campaign-1", req.URL.Query().Get("parent_campaign"))

		_, _ = fmt.Fprintf(w, `{"items":[%s,%s],"next_starting_after":"cursor-2"}`, fixture, fixtureNulls)
	})

	page, err := s.svc().List(context.Background(),
		subsequence.WithLimit(25),
		subsequence.WithParentCampaign("campaign-1"),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)
	s.Require().NotNil(page.Items[0].DailyLimit)
	s.InDelta(50, *page.Items[0].DailyLimit, 0)
	s.Nil(page.Items[1].DailyLimit)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string.
func (s *SubsequenceTestSuite) TestListWithoutOptions() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery)
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background(), nil)

	s.Require().NoError(err)
	s.Empty(page.Items)
}

// TestListFailure verifies a failed list returns no page.
func (s *SubsequenceTestSuite) TestListFailure() {
	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "slow")
	})

	page, err := s.svc().List(context.Background())

	instantlytest.AssertAPIError(s.T(), err, http.StatusTooManyRequests)
	s.Nil(page)
}

// TestGet verifies a single subsequence decodes, including the raw payloads.
func (s *SubsequenceTestSuite) TestGet() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(fixture))
	})

	got, err := s.svc().Get(context.Background(), testID)

	s.Require().NoError(err)
	s.Equal("Follow up", got.Name)
	s.JSONEq(`{"trigger":"no_reply"}`, string(got.Conditions))
	s.JSONEq(`[{"steps":[]}]`, string(got.Sequences))
}

// TestGetFailure verifies a missing subsequence returns no value.
func (s *SubsequenceTestSuite) TestGetFailure() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no subsequence")
	})

	got, err := s.svc().Get(context.Background(), "missing")

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestUpdate verifies the patch body is sent and the subsequence decodes.
func (s *SubsequenceTestSuite) TestUpdate() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))

		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("Renamed", received["name"])
		s.Equal("unlimited", received["daily_limit_mode"])

		_, _ = w.Write([]byte(fixture))
	})

	got, err := s.svc().Update(context.Background(), testID, subsequence.UpdateRequest{
		Name:           "Renamed",
		DailyLimitMode: subsequence.DailyLimitUnlimited,
	})

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestUpdateOmitsUnsetFields verifies an empty patch sends no field.
func (s *SubsequenceTestSuite) TestUpdateOmitsUnsetFields() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Empty(received)

		_, _ = w.Write([]byte(fixture))
	})

	got, err := s.svc().Update(context.Background(), testID, subsequence.UpdateRequest{})

	s.Require().NoError(err)
	s.NotNil(got)
}

// TestUpdateFailure verifies a failed patch returns no value.
func (s *SubsequenceTestSuite) TestUpdateFailure() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no subsequence")
	})

	got, err := s.svc().Update(context.Background(), "missing", subsequence.UpdateRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestDelete verifies the deleted subsequence is returned to the caller.
func (s *SubsequenceTestSuite) TestDelete() {
	s.Router.Delete(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(fixture))
	})

	got, err := s.svc().Delete(context.Background(), testID)

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestDeleteFailure verifies a failed delete returns no value.
func (s *SubsequenceTestSuite) TestDeleteFailure() {
	s.Router.Delete(idPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no subsequence")
	})

	got, err := s.svc().Delete(context.Background(), "missing")

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestDuplicate verifies the duplicate body is sent and the new subsequence decodes.
func (s *SubsequenceTestSuite) TestDuplicate() {
	s.Router.Post(duplicatePatt, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))

		var received subsequence.DuplicateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("campaign-2", received.ParentCampaign)
		s.Equal("Copy", received.Name)

		_, _ = w.Write([]byte(fixture))
	})

	got, err := s.svc().Duplicate(context.Background(), testID, subsequence.DuplicateRequest{
		ParentCampaign: "campaign-2",
		Name:           "Copy",
	})

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestDuplicateFailure verifies a failed duplicate returns no subsequence.
func (s *SubsequenceTestSuite) TestDuplicateFailure() {
	s.Router.Post(duplicatePatt, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no subsequence")
	})

	got, err := s.svc().Duplicate(context.Background(), "missing", subsequence.DuplicateRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestPauseResume verifies the lifecycle actions POST with no body and return
// the subsequence.
func (s *SubsequenceTestSuite) TestPauseResume() {
	tests := []struct {
		name    string
		pattern string
		call    func(svc *subsequence.Service) (*subsequence.Subsequence, error)
	}{
		{"pause", pausePatt, func(svc *subsequence.Service) (*subsequence.Subsequence, error) {
			return svc.Pause(context.Background(), testID)
		}},
		{"resume", resumePatt, func(svc *subsequence.Service) (*subsequence.Subsequence, error) {
			return svc.Resume(context.Background(), testID)
		}},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.Router.Post(test.pattern, func(w http.ResponseWriter, req *http.Request) {
				s.Equal(testID, instantlytest.PathParam(req, "id"))

				body, err := instantlytest.ReadAll(req)
				s.NoError(err)
				s.Empty(body, "a lifecycle action sends no request body")

				_, _ = w.Write([]byte(fixture))
			})

			got, err := test.call(s.svc())

			s.Require().NoError(err)
			s.Equal(testID, got.ID)
		})
	}
}

// TestPauseFailure verifies a failed lifecycle action returns no subsequence.
func (s *SubsequenceTestSuite) TestPauseFailure() {
	s.Router.Post(pausePatt, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no subsequence")
	})

	got, err := s.svc().Pause(context.Background(), "missing")

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestSendingStatus verifies the raw summary and diagnostics are preserved.
func (s *SubsequenceTestSuite) TestSendingStatus() {
	s.Router.Get(sendingPatt, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(`{"summary":{"sending":true},"diagnostics":[{"code":"OK"}]}`))
	})

	got, err := s.svc().SendingStatus(context.Background(), testID)

	s.Require().NoError(err)
	s.JSONEq(`{"sending":true}`, string(got.Summary))
	s.JSONEq(`[{"code":"OK"}]`, string(got.Diagnostics))
}

// TestSendingStatusFailure verifies a failed status returns no value.
func (s *SubsequenceTestSuite) TestSendingStatusFailure() {
	s.Router.Get(sendingPatt, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no subsequence")
	})

	got, err := s.svc().SendingStatus(context.Background(), "missing")

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestListIter verifies the iterator stitches pages together and stops on error.
func (s *SubsequenceTestSuite) TestListIter() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, page([]string{"a", "b"}, "cursor-2"))
			return
		}
		_, _ = fmt.Fprint(w, page([]string{"c"}, ""))
	})

	seen := make([]string, 0, 3)
	for got, err := range s.svc().ListIter(context.Background(), subsequence.WithParentCampaign("campaign-1")) {
		s.Require().NoError(err)
		seen = append(seen, got.ID)
	}

	s.Equal([]string{"a", "b", "c"}, seen)
	s.Equal(int64(2), requests.Load())
}

// TestListIterStopsOnError verifies a failure ends the iteration.
func (s *SubsequenceTestSuite) TestListIterStopsOnError() {
	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "slow")
	})

	var iterErr error
	for got, err := range s.svc().ListIter(context.Background()) {
		if err != nil {
			iterErr = err
			s.Nil(got)
			break
		}
	}

	instantlytest.AssertAPIError(s.T(), iterErr, http.StatusTooManyRequests)
}

// TestPathParametersAreEscaped verifies a caller-supplied id cannot rewrite the
// request path.
func (s *SubsequenceTestSuite) TestPathParametersAreEscaped() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, fixture), nil
			},
		)},
	))

	_, err := subsequence.New(client).Get(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/subsequences/..%2Fadmin%3Fx=1", requestURI)
}

// TestListOptions verifies each documented list query parameter renders correctly.
func (s *SubsequenceTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option subsequence.ListOption
		key    string
		value  string
	}{
		{"limit", subsequence.WithLimit(50), "limit", "50"},
		{"starting after", subsequence.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"parent campaign", subsequence.WithParentCampaign("campaign-1"), "parent_campaign", "campaign-1"},
		{"search", subsequence.WithSearch("follow"), "search", "follow"},
	}

	s.Require().Len(tests, 4)

	for _, test := range tests {
		s.Run(test.name, func() {
			q := instantly.NewQuery()
			test.option(q)

			s.Require().Equal(1, q.Len())
			s.Equal(test.value, q.Get(test.key))
		})
	}
}

// svc builds a Subsequence service pointed at the suite's mock client.
func (s *SubsequenceTestSuite) svc() *subsequence.Service {
	return subsequence.New(s.Client)
}

// page renders one page of a list response for the given subsequence ids.
func page(ids []string, nextCursor string) string {
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, fmt.Sprintf(
			`{"id":%q,"name":"S","parent_campaign":"campaign-1","workspace":"ws-1","status":1,`+
				`"timestamp_created":"2026-08-01T10:00:00.000Z",`+
				`"timestamp_leads_updated":"2026-08-01T11:00:00.000Z",`+
				`"conditions":{},"subsequence_schedule":{},"sequences":[]}`,
			id,
		))
	}

	if nextCursor == "" {
		return fmt.Sprintf(`{"items":[%s]}`, strings.Join(items, ","))
	}

	return fmt.Sprintf(`{"items":[%s],"next_starting_after":%q}`, strings.Join(items, ","), nextCursor)
}
