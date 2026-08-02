package inboxreport_test

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/mrz1836/go-instantly/inboxreport"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// TestListIterWalksEveryPage verifies the iterator stitches every page into a
// single sequence, in order, and carries the required test_id onto every page.
func (s *InboxReportTestSuite) TestListIterWalksEveryPage() {
	var requests atomic.Int64
	testIDs := make([]string, 0, 3)

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		testIDs = append(testIDs, req.URL.Query().Get("test_id"))

		switch req.URL.Query().Get("starting_after") {
		case "":
			_, _ = fmt.Fprint(w, reportPage([]string{"r1", "r2"}, "cursor-2"))
		case "cursor-2":
			_, _ = fmt.Fprint(w, reportPage([]string{"r3", "r4"}, "cursor-3"))
		default:
			_, _ = fmt.Fprint(w, reportPage([]string{"r5"}, ""))
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
	s.Equal([]string{"r1", "r2", "r3", "r4", "r5"}, seen)
	s.Equal(int64(3), requests.Load(), "one request per page, and no request after the last cursor")
	s.Equal([]string{testID, testID, testID}, testIDs, "the required test_id must be carried onto every page")
}

// TestListIterStopsOnBreak verifies breaking out of the loop issues no further
// requests.
func (s *InboxReportTestSuite) TestListIterStopsOnBreak() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = fmt.Fprint(w, reportPage([]string{"r1", "r2", "r3"}, "cursor-2"))
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

	s.Equal([]string{"r1", "r2"}, seen)
	s.Equal(int64(1), requests.Load(), "breaking mid-page must not fetch the next page")
}

// TestListIterStopsOnError verifies a failure on a later page is yielded with a
// nil report and ends the iteration.
func (s *InboxReportTestSuite) TestListIterStopsOnError() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, reportPage([]string{"r1"}, "cursor-2"))
			return
		}

		instantlytest.WriteAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "rate limited")
	})

	seen := make([]string, 0, 1)
	var iterErr error
	var errReport *inboxreport.Report
	yields := 0

	for got, err := range s.svc().ListIter(context.Background(), testID) {
		yields++

		if err != nil {
			iterErr = err
			errReport = got
			continue
		}

		seen = append(seen, got.ID)
	}

	s.Equal([]string{"r1"}, seen)
	s.Equal(2, yields, "iteration must end at the failure rather than retrying the page")
	s.Nil(errReport, "an error is yielded with no report")

	instantlytest.AssertAPIError(s.T(), iterErr, http.StatusTooManyRequests)
	s.Equal(int64(2), requests.Load())
}

// TestListIterStopsOnCancellation verifies a context canceled part way through
// iteration is reported and stops any further request.
func (s *InboxReportTestSuite) TestListIterStopsOnCancellation() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = fmt.Fprint(w, reportPage([]string{"r1"}, "cursor-2"))
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

	s.Equal([]string{"r1"}, seen)
	s.Require().Error(iterErr)
	s.Require().ErrorIs(iterErr, context.Canceled)
	s.Equal(int64(1), requests.Load(), "no page may be requested after cancellation")
}

// TestListIterEmptyFirstPage verifies an empty result set ends iteration
// immediately without yielding anything.
func (s *InboxReportTestSuite) TestListIterEmptyFirstPage() {
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
func (s *InboxReportTestSuite) TestListIterStopsOnEmptyPage() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, reportPage([]string{"r1"}, "cursor-2"))
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
	s.Equal([]string{"r1"}, seen)
	s.Equal(int64(2), requests.Load(), "an empty page ends iteration even when a cursor is offered")
}

// TestListIterKeepsCallerOptions verifies the caller's filters and required
// test_id are sent on every page, and that the page cursor overrides the
// caller's.
func (s *InboxReportTestSuite) TestListIterKeepsCallerOptions() {
	cursors := make([]string, 0, 2)
	skips := make([]string, 0, 2)
	testIDs := make([]string, 0, 2)
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		cursors = append(cursors, req.URL.Query().Get("starting_after"))
		skips = append(skips, req.URL.Query().Get("skip_blacklist_report"))
		testIDs = append(testIDs, req.URL.Query().Get("test_id"))

		if req.URL.Query().Get("starting_after") == "caller-cursor" {
			_, _ = fmt.Fprint(w, reportPage([]string{"r1"}, "cursor-2"))
			return
		}

		_, _ = fmt.Fprint(w, reportPage([]string{"r2"}, ""))
	})

	callerOpts := []inboxreport.ListOption{
		inboxreport.WithSkipBlacklistReport(true),
		inboxreport.WithStartingAfter("caller-cursor"),
	}

	seen := make([]string, 0, 2)
	for got, err := range s.svc().ListIter(context.Background(), testID, callerOpts...) {
		if err != nil {
			break
		}
		seen = append(seen, got.ID)
	}

	s.Equal([]string{"r1", "r2"}, seen)
	s.Equal(int64(2), requests.Load())
	s.Equal([]string{"caller-cursor", "cursor-2"}, cursors, "the page cursor must override the caller's")
	s.Equal([]string{"true", "true"}, skips, "the caller's filters must survive every page")
	s.Equal([]string{testID, testID}, testIDs, "the required test_id must survive every page")

	s.Require().Len(callerOpts, 2)
}

// reportPage renders one page of a list response containing the given report
// identifiers, followed by the cursor of the next page when there is one.
func reportPage(ids []string, nextCursor string) string {
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, fmt.Sprintf(
			`{"id":%q,"organization_id":"org-uuid-1","test_id":"test-uuid-1",`+
				`"timestamp_created":"2026-08-01T19:32:41.209Z","timestamp_created_date":"2026-08-01",`+
				`"domain":"example.com","domain_ip":"203.0.113.10","spam_assassin_score":0}`,
			id,
		))
	}

	return instantlytest.Page(items, nextCursor)
}
