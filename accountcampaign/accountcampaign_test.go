package accountcampaign_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/accountcampaign"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

const (
	mappingPattern = "/api/v2/account-campaign-mappings/:email"
	testEmail      = "sender@example.com"
)

// MappingTestSuite exercises the account-campaign-mappings service.
type MappingTestSuite struct {
	instantlytest.Suite
}

// TestMappingSuite runs the account-campaign-mappings suite.
func TestMappingSuite(t *testing.T) {
	suite.Run(t, new(MappingTestSuite))
}

// TestList verifies a page decodes and the query and path are sent correctly.
func (s *MappingTestSuite) TestList() {
	s.Router.Get(mappingPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testEmail, instantlytest.PathParam(req, "email"))
		s.Equal("25", req.URL.Query().Get("limit"))

		_, _ = w.Write([]byte(
			`{"items":[{"campaign_id":"c1","campaign_name":"Launch","status":1,` +
				`"timestamp_created":"2026-08-01T10:00:00.000Z"}],"next_starting_after":"cursor-2"}`,
		))
	})

	page, err := accountcampaign.New(s.Client).List(
		context.Background(), testEmail, accountcampaign.WithLimit(25),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 1)
	s.Equal("c1", page.Items[0].CampaignID)
	s.Equal("Launch", page.Items[0].CampaignName)
	s.Equal(int64(1), page.Items[0].Status)
	s.Equal("cursor-2", page.NextStartingAfter)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string.
func (s *MappingTestSuite) TestListWithoutOptions() {
	s.Router.Get(mappingPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery)
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := accountcampaign.New(s.Client).List(context.Background(), testEmail, nil)

	s.Require().NoError(err)
	s.Empty(page.Items)
}

// TestListFailure verifies a failed list returns no page.
func (s *MappingTestSuite) TestListFailure() {
	s.Router.Get(mappingPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusNotFound, "Not Found", "no account")
	})

	page, err := accountcampaign.New(s.Client).List(context.Background(), "missing@example.com")

	instantlytest.AssertAPIError(s.T(), err, http.StatusNotFound)
	s.Nil(page)
}

// TestListOptions verifies each documented query parameter renders correctly.
func (s *MappingTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option accountcampaign.ListOption
		key    string
		value  string
	}{
		{"limit", accountcampaign.WithLimit(50), "limit", "50"},
		{"starting after", accountcampaign.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
	}

	s.Require().Len(tests, 2)

	for _, test := range tests {
		s.Run(test.name, func() {
			q := instantly.NewQuery()
			test.option(q)

			s.Require().Equal(1, q.Len())
			s.Equal(test.value, q.Get(test.key))
		})
	}
}

// TestPathParametersAreEscaped verifies a caller-supplied email cannot rewrite
// the request path.
func (s *MappingTestSuite) TestPathParametersAreEscaped() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, `{"items":[]}`), nil
			},
		)},
	))

	_, err := accountcampaign.New(client).List(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/account-campaign-mappings/..%2Fadmin%3Fx=1", requestURI)
}

// TestListIter verifies the iterator stitches pages together and stops on the
// empty cursor.
func (s *MappingTestSuite) TestListIter() {
	var requests atomic.Int64

	s.Router.Get(mappingPattern, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, mappingPage([]string{"c1", "c2"}, "cursor-2"))
			return
		}

		_, _ = fmt.Fprint(w, mappingPage([]string{"c3"}, ""))
	})

	seen := make([]string, 0, 3)
	for got, err := range accountcampaign.New(s.Client).ListIter(context.Background(), testEmail) {
		s.Require().NoError(err)
		seen = append(seen, got.CampaignID)
	}

	s.Equal([]string{"c1", "c2", "c3"}, seen)
	s.Equal(int64(2), requests.Load())
}

// TestListIterStopsOnError verifies a failure ends the iteration.
func (s *MappingTestSuite) TestListIterStopsOnError() {
	s.Router.Get(mappingPattern, func(w http.ResponseWriter, _ *http.Request) {
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "slow")
	})

	var iterErr error
	for got, err := range accountcampaign.New(s.Client).ListIter(context.Background(), testEmail) {
		if err != nil {
			iterErr = err
			s.Nil(got)
			break
		}
	}

	instantlytest.AssertAPIError(s.T(), iterErr, http.StatusTooManyRequests)
}

// mappingPage renders one page of mappings for the given campaign identifiers.
func mappingPage(ids []string, nextCursor string) string {
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, fmt.Sprintf(
			`{"campaign_id":%q,"campaign_name":"C","status":1,`+
				`"timestamp_created":"2026-08-01T10:00:00.000Z"}`, id,
		))
	}

	if nextCursor == "" {
		return fmt.Sprintf(`{"items":[%s]}`, strings.Join(items, ","))
	}

	return fmt.Sprintf(`{"items":[%s],"next_starting_after":%q}`, strings.Join(items, ","), nextCursor)
}
