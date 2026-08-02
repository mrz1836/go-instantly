package leadlabel_test

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
	"github.com/mrz1836/go-instantly/leadlabel"
)

const (
	listPath     = "/api/v2/lead-labels"
	idPattern    = "/api/v2/lead-labels/:id"
	aiReplyPath  = "/api/v2/lead-labels/ai-reply-label"
	testID       = "label-uuid-1"
	labelFixture = `{"id":"label-uuid-1","label":"Interested","interest_status_label":"positive",` +
		`"interest_status":1,"created_by":"user-1","organization_id":"org-1",` +
		`"timestamp_created":"2026-08-01T10:00:00.000Z","description":"Warm lead","use_with_ai":true}`
	labelFixtureNulls = `{"id":"label-uuid-2","label":"Cold","interest_status_label":"negative",` +
		`"interest_status":-1,"created_by":"user-1","organization_id":"org-1",` +
		`"timestamp_created":"2026-08-01T10:00:00.000Z","description":null,"use_with_ai":null}`
)

// LeadLabelTestSuite exercises the Lead Label API service.
type LeadLabelTestSuite struct {
	instantlytest.Suite
}

// TestLeadLabelSuite runs the Lead Label API suite.
func TestLeadLabelSuite(t *testing.T) {
	suite.Run(t, new(LeadLabelTestSuite))
}

// TestCreate verifies the create body reaches the API and the label decodes.
func (s *LeadLabelTestSuite) TestCreate() {
	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		var received leadlabel.CreateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("Interested", received.Label)
		s.Equal("positive", received.InterestStatusLabel)

		_, _ = w.Write([]byte(labelFixture))
	})

	got, err := s.svc().Create(context.Background(), leadlabel.CreateRequest{
		Label:               "Interested",
		InterestStatusLabel: "positive",
		UseWithAI:           instantly.Ptr(true),
	})

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
	s.Equal(int64(1), got.InterestStatus)
}

// TestList verifies a page decodes, including nullable-vs-zero fields.
func (s *LeadLabelTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("25", req.URL.Query().Get("limit"))
		s.Equal("positive", req.URL.Query().Get("interest_status"))

		_, _ = fmt.Fprintf(w, `{"items":[%s,%s],"next_starting_after":"cursor-2"}`, labelFixture, labelFixtureNulls)
	})

	page, err := s.svc().List(context.Background(),
		leadlabel.WithLimit(25),
		leadlabel.WithInterestStatus(leadlabel.InterestPositive),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)
	s.Require().NotNil(page.Items[0].Description)
	s.Equal("Warm lead", *page.Items[0].Description)
	s.Nil(page.Items[1].Description)
	s.Nil(page.Items[1].UseWithAI)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string.
func (s *LeadLabelTestSuite) TestListWithoutOptions() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery)
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background(), nil)

	s.Require().NoError(err)
	s.Empty(page.Items)
}

// TestGet verifies a single label decodes.
func (s *LeadLabelTestSuite) TestGet() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(labelFixture))
	})

	got, err := s.svc().Get(context.Background(), testID)

	s.Require().NoError(err)
	s.Equal("Interested", got.Label)
}

// TestUpdate verifies the patch body is sent and the label decodes.
func (s *LeadLabelTestSuite) TestUpdate() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("Renamed", received["label"])

		_, _ = w.Write([]byte(labelFixture))
	})

	got, err := s.svc().Update(context.Background(), testID, leadlabel.UpdateRequest{Label: "Renamed"})

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestUpdateOmitsUnsetFields verifies an empty patch sends no field.
func (s *LeadLabelTestSuite) TestUpdateOmitsUnsetFields() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Empty(received)

		_, _ = w.Write([]byte(labelFixture))
	})

	got, err := s.svc().Update(context.Background(), testID, leadlabel.UpdateRequest{})

	s.Require().NoError(err)
	s.NotNil(got)
}

// TestDelete verifies the deleted label is returned to the caller.
func (s *LeadLabelTestSuite) TestDelete() {
	s.Router.Delete(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(labelFixture))
	})

	got, err := s.svc().Delete(context.Background(), testID)

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestTestAIReplyLabel verifies the reply text is sent and the raw result decodes.
func (s *LeadLabelTestSuite) TestTestAIReplyLabel() {
	s.Router.Post(aiReplyPath, func(w http.ResponseWriter, req *http.Request) {
		var received leadlabel.AIReplyLabelRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("Sounds great!", received.ReplyText)

		_, _ = w.Write([]byte(`{"result":{"label":"Interested"},"custom_labels_considered":["Interested"]}`))
	})

	got, err := s.svc().TestAIReplyLabel(context.Background(), leadlabel.AIReplyLabelRequest{
		ReplyText: "Sounds great!",
	})

	s.Require().NoError(err)
	s.JSONEq(`{"label":"Interested"}`, string(got.Result))
	s.JSONEq(`["Interested"]`, string(got.CustomLabelsConsidered))
}

// TestListIter verifies the iterator stitches pages together and stops on error.
func (s *LeadLabelTestSuite) TestListIter() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, labelPage([]string{"l1", "l2"}, "cursor-2"))
			return
		}
		_, _ = fmt.Fprint(w, labelPage([]string{"l3"}, ""))
	})

	seen := make([]string, 0, 3)
	for got, err := range s.svc().ListIter(context.Background(), leadlabel.WithLimit(2)) {
		s.Require().NoError(err)
		seen = append(seen, got.ID)
	}

	s.Equal([]string{"l1", "l2", "l3"}, seen)
	s.Equal(int64(2), requests.Load())
}

// TestListIterStopsOnError verifies a failure ends the iteration.
func (s *LeadLabelTestSuite) TestListIterStopsOnError() {
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
func (s *LeadLabelTestSuite) TestPathParametersAreEscaped() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, labelFixture), nil
			},
		)},
	))

	_, err := leadlabel.New(client).Get(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/lead-labels/..%2Fadmin%3Fx=1", requestURI)
}

// TestListOptions verifies each documented list query parameter renders correctly.
func (s *LeadLabelTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option leadlabel.ListOption
		key    string
		value  string
	}{
		{"limit", leadlabel.WithLimit(50), "limit", "50"},
		{"starting after", leadlabel.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"search", leadlabel.WithSearch("warm"), "search", "warm"},
		{"interest status", leadlabel.WithInterestStatus(leadlabel.InterestNeutral), "interest_status", "neutral"},
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

// TestFailures verifies every endpoint surfaces the documented API error.
func (s *LeadLabelTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "create", Method: http.MethodPost, Path: listPath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.Create(ctx, leadlabel.CreateRequest{}); return err },
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
			Call: func() error { _, err := svc.Update(ctx, "missing", leadlabel.UpdateRequest{}); return err },
		},
		{
			Name: "delete", Method: http.MethodDelete, Path: idPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Delete(ctx, "missing"); return err },
		},
		{
			Name: "testAIReplyLabel", Method: http.MethodPost, Path: aiReplyPath, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.TestAIReplyLabel(ctx, leadlabel.AIReplyLabelRequest{}); return err },
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor parses a valid
// timestamp and reports an error for an unparseable one.
func (s *LeadLabelTestSuite) TestParsedTimestampCreated() {
	got, err := (&leadlabel.LeadLabel{TimestampCreated: "2026-08-01T10:00:00.000Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&leadlabel.LeadLabel{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// svc builds a Lead Label service pointed at the suite's mock client.
func (s *LeadLabelTestSuite) svc() *leadlabel.Service {
	return leadlabel.New(s.Client)
}

// labelPage renders one page of a list response for the given label ids.
func labelPage(ids []string, nextCursor string) string {
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, fmt.Sprintf(
			`{"id":%q,"label":"L","interest_status_label":"positive","interest_status":1,`+
				`"created_by":"user-1","organization_id":"org-1","timestamp_created":"2026-08-01T10:00:00.000Z"}`,
			id,
		))
	}

	return instantlytest.Page(items, nextCursor)
}
