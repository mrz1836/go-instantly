package inboxanalytics_test

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/mrz1836/go-instantly/inboxanalytics"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// TestListIterWalksEveryPage verifies the iterator stitches every page into a
// single sequence, in order, and carries the required test_id onto every page.
func (s *InboxAnalyticsTestSuite) TestListIterWalksEveryPage() {
	var requests atomic.Int64
	testIDs := make([]string, 0, 3)

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		testIDs = append(testIDs, req.URL.Query().Get("test_id"))

		switch req.URL.Query().Get("starting_after") {
		case "":
			_, _ = fmt.Fprint(w, analyticsPage([]string{"a1", "a2"}, "cursor-2"))
		case "cursor-2":
			_, _ = fmt.Fprint(w, analyticsPage([]string{"a3", "a4"}, "cursor-3"))
		default:
			_, _ = fmt.Fprint(w, analyticsPage([]string{"a5"}, ""))
		}
	})

	seen := make([]string, 0, 5)
	var iterErr error

	for got, err := range s.svc().ListIter(context.Background(), testID) {
		if err != nil {
			iterErr = err
			break
		}
		seen = append(seen, got.ID)
	}

	s.Require().NoError(iterErr)
	s.Equal([]string{"a1", "a2", "a3", "a4", "a5"}, seen)
	s.Equal(int64(3), requests.Load(), "one request per page, and no request after the last cursor")
	s.Equal([]string{testID, testID, testID}, testIDs, "the required test_id must be carried onto every page")
}

// TestListIterStopsOnBreak verifies breaking out of the loop issues no further
// requests.
func (s *InboxAnalyticsTestSuite) TestListIterStopsOnBreak() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = fmt.Fprint(w, analyticsPage([]string{"a1", "a2", "a3"}, "cursor-2"))
	})

	seen := make([]string, 0, 2)

	for got, err := range s.svc().ListIter(context.Background(), testID) {
		if err != nil {
			break
		}

		seen = append(seen, got.ID)
		if len(seen) == 2 {
			break
		}
	}

	s.Equal([]string{"a1", "a2"}, seen)
	s.Equal(int64(1), requests.Load(), "breaking mid-page must not fetch the next page")
}

// TestListIterStopsOnError verifies a failure on a later page is yielded with a
// nil event and ends the iteration.
func (s *InboxAnalyticsTestSuite) TestListIterStopsOnError() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, analyticsPage([]string{"a1"}, "cursor-2"))
			return
		}

		instantlytest.WriteAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "rate limited")
	})

	seen := make([]string, 0, 1)
	var iterErr error
	var errEvent *inboxanalytics.Analytics
	yields := 0

	for got, err := range s.svc().ListIter(context.Background(), testID) {
		yields++

		if err != nil {
			iterErr = err
			errEvent = got
			continue
		}

		seen = append(seen, got.ID)
	}

	s.Equal([]string{"a1"}, seen)
	s.Equal(2, yields, "iteration must end at the failure rather than retrying the page")
	s.Nil(errEvent, "an error is yielded with no event")

	instantlytest.AssertAPIError(s.T(), iterErr, http.StatusTooManyRequests)
	s.Equal(int64(2), requests.Load())
}

// TestListIterStopsOnCancellation verifies a context canceled part way through
// iteration is reported and stops any further request.
func (s *InboxAnalyticsTestSuite) TestListIterStopsOnCancellation() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = fmt.Fprint(w, analyticsPage([]string{"a1"}, "cursor-2"))
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	seen := make([]string, 0, 1)
	var iterErr error

	for got, err := range s.svc().ListIter(ctx, testID) {
		if err != nil {
			iterErr = err
			break
		}

		seen = append(seen, got.ID)
		cancel()
	}

	s.Equal([]string{"a1"}, seen)
	s.Require().Error(iterErr)
	s.Require().ErrorIs(iterErr, context.Canceled)
	s.Equal(int64(1), requests.Load(), "no page may be requested after cancellation")
}

// TestListIterEmptyFirstPage verifies an empty result set ends iteration
// immediately without yielding anything.
func (s *InboxAnalyticsTestSuite) TestListIterEmptyFirstPage() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	yields := 0
	for range s.svc().ListIter(context.Background(), testID) {
		yields++
	}

	s.Zero(yields, "an empty result set yields nothing at all")
	s.Equal(int64(1), requests.Load())
}

// TestListIterStopsOnEmptyPage verifies an API that keeps handing back a cursor
// cannot drive an unbounded number of requests.
func (s *InboxAnalyticsTestSuite) TestListIterStopsOnEmptyPage() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, analyticsPage([]string{"a1"}, "cursor-2"))
			return
		}

		_, _ = w.Write([]byte(`{"items":[],"next_starting_after":"cursor-3"}`))
	})

	seen := make([]string, 0, 1)
	var iterErr error

	for got, err := range s.svc().ListIter(context.Background(), testID) {
		if err != nil {
			iterErr = err
			break
		}
		seen = append(seen, got.ID)
	}

	s.Require().NoError(iterErr)
	s.Equal([]string{"a1"}, seen)
	s.Equal(int64(2), requests.Load(), "an empty page ends iteration even when a cursor is offered")
}

// TestListIterKeepsCallerOptions verifies the caller's filters and required
// test_id are sent on every page, and that the page cursor overrides the
// caller's.
func (s *InboxAnalyticsTestSuite) TestListIterKeepsCallerOptions() {
	cursors := make([]string, 0, 2)
	dates := make([]string, 0, 2)
	testIDs := make([]string, 0, 2)
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		cursors = append(cursors, req.URL.Query().Get("starting_after"))
		dates = append(dates, req.URL.Query().Get("date_from"))
		testIDs = append(testIDs, req.URL.Query().Get("test_id"))

		if req.URL.Query().Get("starting_after") == "caller-cursor" {
			_, _ = fmt.Fprint(w, analyticsPage([]string{"a1"}, "cursor-2"))
			return
		}

		_, _ = fmt.Fprint(w, analyticsPage([]string{"a2"}, ""))
	})

	callerOpts := []inboxanalytics.ListOption{
		inboxanalytics.WithDateFrom("2026-08-01"),
		inboxanalytics.WithStartingAfter("caller-cursor"),
	}

	seen := make([]string, 0, 2)
	for got, err := range s.svc().ListIter(context.Background(), testID, callerOpts...) {
		if err != nil {
			break
		}
		seen = append(seen, got.ID)
	}

	s.Equal([]string{"a1", "a2"}, seen)
	s.Equal(int64(2), requests.Load())
	s.Equal([]string{"caller-cursor", "cursor-2"}, cursors, "the page cursor must override the caller's")
	s.Equal([]string{"2026-08-01", "2026-08-01"}, dates, "the caller's filters must survive every page")
	s.Equal([]string{testID, testID}, testIDs, "the required test_id must survive every page")

	s.Require().Len(callerOpts, 2)
}

// analyticsPage renders one page of a list response containing the given event
// identifiers, followed by the cursor of the next page when there is one.
func analyticsPage(ids []string, nextCursor string) string {
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, fmt.Sprintf(
			`{"id":%q,"organization_id":"org-uuid-1","test_id":"test-uuid-1",`+
				`"timestamp_created":"2026-08-01T19:32:41.209Z","timestamp_created_date":"2026-08-01"}`,
			id,
		))
	}

	return instantlytest.Page(items, nextCursor)
}
