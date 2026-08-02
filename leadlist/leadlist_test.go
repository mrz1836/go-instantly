package leadlist_test

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
	"github.com/mrz1836/go-instantly/leadlist"
)

const (
	listPath    = "/api/v2/lead-lists"
	idPattern   = "/api/v2/lead-lists/:id"
	statsPatt   = "/api/v2/lead-lists/:id/verification-stats"
	testID      = "list-uuid-1"
	listFixture = `{"id":"list-uuid-1","name":"Prospects","organization_id":"org-1",` +
		`"timestamp_created":"2026-08-01T10:00:00.000Z","has_enrichment_task":true,"owned_by":"user-1"}`
	listFixtureNulls = `{"id":"list-uuid-2","name":"Bare","organization_id":"org-1",` +
		`"timestamp_created":"2026-08-01T10:00:00.000Z","has_enrichment_task":null,"owned_by":null}`
)

// LeadListTestSuite exercises the Lead List API service.
type LeadListTestSuite struct {
	instantlytest.Suite
}

// TestLeadListSuite runs the Lead List API suite.
func TestLeadListSuite(t *testing.T) {
	suite.Run(t, new(LeadListTestSuite))
}

// TestCreate verifies the create body reaches the API and the list decodes.
func (s *LeadListTestSuite) TestCreate() {
	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		var received leadlist.CreateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("Prospects", received.Name)

		_, _ = w.Write([]byte(listFixture))
	})

	got, err := s.svc().Create(context.Background(), leadlist.CreateRequest{
		Name:              "Prospects",
		HasEnrichmentTask: instantly.Ptr(true),
	})

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
	s.Require().NotNil(got.HasEnrichmentTask)
	s.True(*got.HasEnrichmentTask)
}

// TestCreateFailure verifies a rejected create returns no list.
func (s *LeadListTestSuite) TestCreateFailure() {
	s.Router.Post(listPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusUnauthorized, "Unauthorized", "bad key")
	})

	got, err := s.svc().Create(context.Background(), leadlist.CreateRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusUnauthorized)
	s.Nil(got)
}

// TestList verifies a page decodes, including nullable-vs-zero fields.
func (s *LeadListTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("25", req.URL.Query().Get("limit"))
		s.Equal("true", req.URL.Query().Get("has_enrichment_task"))

		_, _ = fmt.Fprintf(w, `{"items":[%s,%s],"next_starting_after":"cursor-2"}`, listFixture, listFixtureNulls)
	})

	page, err := s.svc().List(context.Background(),
		leadlist.WithLimit(25),
		leadlist.WithHasEnrichmentTask(true),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)
	s.Nil(page.Items[1].HasEnrichmentTask)
	s.Nil(page.Items[1].OwnedBy)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string.
func (s *LeadListTestSuite) TestListWithoutOptions() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery)
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background(), nil)

	s.Require().NoError(err)
	s.Empty(page.Items)
}

// TestListFailure verifies a failed list returns no page.
func (s *LeadListTestSuite) TestListFailure() {
	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "slow")
	})

	page, err := s.svc().List(context.Background())

	instantlytest.AssertAPIError(s.T(), err, http.StatusTooManyRequests)
	s.Nil(page)
}

// TestGet verifies a single list decodes.
func (s *LeadListTestSuite) TestGet() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(listFixture))
	})

	got, err := s.svc().Get(context.Background(), testID)

	s.Require().NoError(err)
	s.Equal("Prospects", got.Name)
}

// TestGetFailure verifies a missing list returns no value.
func (s *LeadListTestSuite) TestGetFailure() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no list")
	})

	got, err := s.svc().Get(context.Background(), "missing")

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestUpdate verifies the patch body is sent and the list decodes.
func (s *LeadListTestSuite) TestUpdate() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("Renamed", received["name"])

		_, _ = w.Write([]byte(listFixture))
	})

	got, err := s.svc().Update(context.Background(), testID, leadlist.UpdateRequest{Name: "Renamed"})

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestUpdateOmitsUnsetFields verifies an empty patch sends no field.
func (s *LeadListTestSuite) TestUpdateOmitsUnsetFields() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Empty(received)

		_, _ = w.Write([]byte(listFixture))
	})

	got, err := s.svc().Update(context.Background(), testID, leadlist.UpdateRequest{})

	s.Require().NoError(err)
	s.NotNil(got)
}

// TestUpdateFailure verifies a failed patch returns no value.
func (s *LeadListTestSuite) TestUpdateFailure() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no list")
	})

	got, err := s.svc().Update(context.Background(), "missing", leadlist.UpdateRequest{})

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestDelete verifies the deleted list is returned to the caller.
func (s *LeadListTestSuite) TestDelete() {
	s.Router.Delete(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(listFixture))
	})

	got, err := s.svc().Delete(context.Background(), testID)

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestDeleteFailure verifies a failed delete returns no value.
func (s *LeadListTestSuite) TestDeleteFailure() {
	s.Router.Delete(idPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no list")
	})

	got, err := s.svc().Delete(context.Background(), "missing")

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestVerificationStats verifies the stats decode, including the raw breakdown.
func (s *LeadListTestSuite) TestVerificationStats() {
	s.Router.Get(statsPatt, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(`{"total_leads":100,"stats":{"verified":80,"invalid":20}}`))
	})

	got, err := s.svc().VerificationStats(context.Background(), testID)

	s.Require().NoError(err)
	s.Equal(int64(100), got.TotalLeads)
	s.JSONEq(`{"verified":80,"invalid":20}`, string(got.Stats))
}

// TestVerificationStatsFailure verifies a failed request returns no value.
func (s *LeadListTestSuite) TestVerificationStatsFailure() {
	s.Router.Get(statsPatt, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no list")
	})

	got, err := s.svc().VerificationStats(context.Background(), "missing")

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(got)
}

// TestListIter verifies the iterator stitches pages together and stops on error.
func (s *LeadListTestSuite) TestListIter() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, listPage([]string{"a", "b"}, "cursor-2"))
			return
		}
		_, _ = fmt.Fprint(w, listPage([]string{"c"}, ""))
	})

	seen := make([]string, 0, 3)
	for got, err := range s.svc().ListIter(context.Background()) {
		s.Require().NoError(err)
		seen = append(seen, got.ID)
	}

	s.Equal([]string{"a", "b", "c"}, seen)
	s.Equal(int64(2), requests.Load())
}

// TestListIterStopsOnError verifies a failure ends the iteration.
func (s *LeadListTestSuite) TestListIterStopsOnError() {
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
func (s *LeadListTestSuite) TestPathParametersAreEscaped() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, listFixture), nil
			},
		)},
	))

	_, err := leadlist.New(client).Get(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/lead-lists/..%2Fadmin%3Fx=1", requestURI)
}

// TestListOptions verifies each documented list query parameter renders correctly.
func (s *LeadListTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option leadlist.ListOption
		key    string
		value  string
	}{
		{"limit", leadlist.WithLimit(50), "limit", "50"},
		{"starting after", leadlist.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"search", leadlist.WithSearch("prospects"), "search", "prospects"},
		{"has enrichment task", leadlist.WithHasEnrichmentTask(false), "has_enrichment_task", "false"},
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

// svc builds a Lead List service pointed at the suite's mock client.
func (s *LeadListTestSuite) svc() *leadlist.Service {
	return leadlist.New(s.Client)
}

// listPage renders one page of a list response for the given list ids.
func listPage(ids []string, nextCursor string) string {
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, fmt.Sprintf(
			`{"id":%q,"name":"L","organization_id":"org-1","timestamp_created":"2026-08-01T10:00:00.000Z"}`,
			id,
		))
	}

	if nextCursor == "" {
		return fmt.Sprintf(`{"items":[%s]}`, strings.Join(items, ","))
	}

	return fmt.Sprintf(`{"items":[%s],"next_starting_after":%q}`, strings.Join(items, ","), nextCursor)
}
