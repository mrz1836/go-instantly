package campaign_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/mrz1836/go-instantly/campaign"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// TestListIterWalksEveryPage verifies the iterator stitches pages together and
// carries the caller's filters onto every page.
func (s *CampaignTestSuite) TestListIterWalksEveryPage() {
	var requests atomic.Int64
	statuses := make([]string, 0, 2)

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		statuses = append(statuses, req.URL.Query().Get("status"))

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, campaignPage([]string{"c1", "c2"}, "cursor-2"))
			return
		}

		_, _ = fmt.Fprint(w, campaignPage([]string{"c3"}, ""))
	})

	seen := make([]string, 0, 3)
	for got, err := range s.svc().ListIter(context.Background(), campaign.WithStatus(campaign.StatusActive)) {
		s.Require().NoError(err)
		seen = append(seen, got.ID)
	}

	s.Equal([]string{"c1", "c2", "c3"}, seen)
	s.Equal(int64(2), requests.Load())
	s.Equal([]string{"1", "1"}, statuses, "the caller's filters survive every page")
}

// TestListIterStopsOnError verifies a failure ends the iteration with a nil
// campaign.
func (s *CampaignTestSuite) TestListIterStopsOnError() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, campaignPage([]string{"c1"}, "cursor-2"))
			return
		}
		instantlytest.WriteAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "slow")
	})

	seen := make([]string, 0, 1)
	var iterErr error
	for got, err := range s.svc().ListIter(context.Background()) {
		if err != nil {
			iterErr = err
			s.Nil(got)
			break
		}
		seen = append(seen, got.ID)
	}

	s.Equal([]string{"c1"}, seen)
	instantlytest.AssertAPIError(s.T(), iterErr, http.StatusTooManyRequests)
}

// campaignPage renders one page of a list response for the given campaign ids.
func campaignPage(ids []string, nextCursor string) string {
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, fmt.Sprintf(
			`{"id":%q,"name":"C","status":1,"campaign_schedule":{"schedules":[]},`+
				`"timestamp_created":"2026-08-01T10:00:00.000Z",`+
				`"timestamp_updated":"2026-08-01T11:00:00.000Z","open_tracking":true}`,
			id,
		))
	}

	if nextCursor == "" {
		return fmt.Sprintf(`{"items":[%s]}`, strings.Join(items, ","))
	}

	return fmt.Sprintf(`{"items":[%s],"next_starting_after":%q}`, strings.Join(items, ","), nextCursor)
}
