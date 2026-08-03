package backgroundjob_test

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/mrz1836/go-instantly/backgroundjob"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// TestListIterWalksEveryPage verifies the iterator stitches every page into a
// single sequence, in order, and stops once the cursor runs out.
func (s *BackgroundJobTestSuite) TestListIterWalksEveryPage() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		switch req.URL.Query().Get("starting_after") {
		case "":
			_, _ = fmt.Fprint(w, jobPage([]string{"j1", "j2"}, "cursor-2"))
		case "cursor-2":
			_, _ = fmt.Fprint(w, jobPage([]string{"j3", "j4"}, "cursor-3"))
		default:
			_, _ = fmt.Fprint(w, jobPage([]string{"j5"}, ""))
		}
	})

	seen := make([]string, 0, 5)
	var iterErr error

	for got, err := range s.svc().ListIter(context.Background()) {
		if err != nil {
			iterErr = err
			break
		}
		seen = append(seen, got.ID)
	}

	s.Require().NoError(iterErr)
	s.Equal([]string{"j1", "j2", "j3", "j4", "j5"}, seen)
	s.Equal(int64(3), requests.Load(), "one request per page, and no request after the last cursor")
}

// TestListIterStopsOnBreak verifies breaking out of the loop issues no further
// requests.
func (s *BackgroundJobTestSuite) TestListIterStopsOnBreak() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = fmt.Fprint(w, jobPage([]string{"j1", "j2", "j3"}, "cursor-2"))
	})

	seen := make([]string, 0, 2)

	for got, err := range s.svc().ListIter(context.Background()) {
		if err != nil {
			break
		}

		seen = append(seen, got.ID)
		if len(seen) == 2 {
			break
		}
	}

	s.Equal([]string{"j1", "j2"}, seen)
	s.Equal(int64(1), requests.Load(), "breaking mid-page must not fetch the next page")
}

// TestListIterStopsOnError verifies a failure on a later page is yielded with a
// nil job and ends the iteration.
func (s *BackgroundJobTestSuite) TestListIterStopsOnError() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, jobPage([]string{"j1"}, "cursor-2"))
			return
		}

		instantlytest.WriteAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "rate limited")
	})

	seen := make([]string, 0, 1)
	var iterErr error
	var errJob *backgroundjob.Job
	yields := 0

	for got, err := range s.svc().ListIter(context.Background()) {
		yields++

		if err != nil {
			iterErr = err
			errJob = got
			continue
		}

		seen = append(seen, got.ID)
	}

	s.Equal([]string{"j1"}, seen)
	s.Equal(2, yields, "iteration must end at the failure rather than retrying the page")
	s.Nil(errJob, "an error is yielded with no job")

	instantlytest.AssertAPIError(s.T(), iterErr, http.StatusTooManyRequests)
	s.Equal(int64(2), requests.Load())
}

// TestListIterStopsOnCancellation verifies a context canceled part way through
// iteration is reported and stops any further request.
func (s *BackgroundJobTestSuite) TestListIterStopsOnCancellation() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = fmt.Fprint(w, jobPage([]string{"j1"}, "cursor-2"))
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	seen := make([]string, 0, 1)
	var iterErr error

	for got, err := range s.svc().ListIter(ctx) {
		if err != nil {
			iterErr = err
			break
		}

		seen = append(seen, got.ID)
		cancel()
	}

	s.Equal([]string{"j1"}, seen)
	s.Require().Error(iterErr)
	s.Require().ErrorIs(iterErr, context.Canceled)
	s.Equal(int64(1), requests.Load(), "no page may be requested after cancellation")
}

// TestListIterEmptyFirstPage verifies an empty result set ends iteration
// immediately without yielding anything.
func (s *BackgroundJobTestSuite) TestListIterEmptyFirstPage() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	yields := 0
	for range s.svc().ListIter(context.Background()) {
		yields++
	}

	s.Zero(yields, "an empty result set yields nothing at all")
	s.Equal(int64(1), requests.Load())
}

// TestListIterStopsOnEmptyPage verifies an API that keeps handing back a cursor
// cannot drive an unbounded number of requests.
func (s *BackgroundJobTestSuite) TestListIterStopsOnEmptyPage() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, jobPage([]string{"j1"}, "cursor-2"))
			return
		}

		_, _ = w.Write([]byte(`{"items":[],"next_starting_after":"cursor-3"}`))
	})

	seen := make([]string, 0, 1)
	var iterErr error

	for got, err := range s.svc().ListIter(context.Background()) {
		if err != nil {
			iterErr = err
			break
		}
		seen = append(seen, got.ID)
	}

	s.Require().NoError(iterErr)
	s.Equal([]string{"j1"}, seen)
	s.Equal(int64(2), requests.Load(), "an empty page ends iteration even when a cursor is offered")
}

// TestListIterKeepsCallerOptions verifies the caller's filters are sent on every
// page, and that the page cursor overrides any cursor the caller supplied.
func (s *BackgroundJobTestSuite) TestListIterKeepsCallerOptions() {
	cursors := make([]string, 0, 2)
	statuses := make([]string, 0, 2)
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		cursors = append(cursors, req.URL.Query().Get("starting_after"))
		statuses = append(statuses, req.URL.Query().Get("status"))

		if req.URL.Query().Get("starting_after") == "caller-cursor" {
			_, _ = fmt.Fprint(w, jobPage([]string{"j1"}, "cursor-2"))
			return
		}

		_, _ = fmt.Fprint(w, jobPage([]string{"j2"}, ""))
	})

	callerOpts := []backgroundjob.ListOption{
		backgroundjob.WithStatus("pending"),
		backgroundjob.WithStartingAfter("caller-cursor"),
	}

	seen := make([]string, 0, 2)
	for got, err := range s.svc().ListIter(context.Background(), callerOpts...) {
		if err != nil {
			break
		}
		seen = append(seen, got.ID)
	}

	s.Equal([]string{"j1", "j2"}, seen)
	s.Equal(int64(2), requests.Load())
	s.Equal([]string{"caller-cursor", "cursor-2"}, cursors, "the page cursor must override the caller's")
	s.Equal([]string{"pending", "pending"}, statuses, "the caller's filters must survive every page")

	s.Require().Len(callerOpts, 2)
}

// jobPage renders one page of a list response containing the given background
// job identifiers, followed by the cursor of the next page when there is one.
func jobPage(ids []string, nextCursor string) string {
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, fmt.Sprintf(
			`{"id":%q,"workspace_id":"ws-1","type":"move-leads","progress":10,`+
				`"status":"pending","created_at":"2026-08-01T10:00:00.000Z",`+
				`"updated_at":"2026-08-01T10:05:00.000Z"}`,
			id,
		))
	}

	return instantlytest.Page(items, nextCursor)
}
