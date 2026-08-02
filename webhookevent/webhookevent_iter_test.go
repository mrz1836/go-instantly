package webhookevent_test

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/webhookevent"
)

// TestListIterWalksEveryPage verifies the iterator stitches every page into a
// single sequence, in order, and stops once the cursor runs out.
func (s *WebhookEventTestSuite) TestListIterWalksEveryPage() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		switch req.URL.Query().Get("starting_after") {
		case "":
			_, _ = fmt.Fprint(w, eventPage([]string{"e1", "e2"}, "cursor-2"))
		case "cursor-2":
			_, _ = fmt.Fprint(w, eventPage([]string{"e3", "e4"}, "cursor-3"))
		default:
			_, _ = fmt.Fprint(w, eventPage([]string{"e5"}, ""))
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
	s.Equal([]string{"e1", "e2", "e3", "e4", "e5"}, seen)
	s.Equal(int64(3), requests.Load(), "one request per page, and no request after the last cursor")
}

// TestListIterStopsOnBreak verifies breaking out of the loop issues no further
// requests.
func (s *WebhookEventTestSuite) TestListIterStopsOnBreak() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = fmt.Fprint(w, eventPage([]string{"e1", "e2", "e3"}, "cursor-2"))
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

	s.Equal([]string{"e1", "e2"}, seen)
	s.Equal(int64(1), requests.Load(), "breaking mid-page must not fetch the next page")
}

// TestListIterStopsOnError verifies a failure on a later page is yielded with a
// nil event and ends the iteration.
func (s *WebhookEventTestSuite) TestListIterStopsOnError() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, eventPage([]string{"e1"}, "cursor-2"))
			return
		}

		instantlytest.WriteAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "rate limited")
	})

	seen := make([]string, 0, 1)
	var iterErr error
	var errEvent *webhookevent.Event
	yields := 0

	for got, err := range s.svc().ListIter(context.Background()) {
		yields++

		if err != nil {
			iterErr = err
			errEvent = got
			continue
		}

		seen = append(seen, got.ID)
	}

	s.Equal([]string{"e1"}, seen)
	s.Equal(2, yields, "iteration must end at the failure rather than retrying the page")
	s.Nil(errEvent, "an error is yielded with no event")

	instantlytest.AssertAPIError(s.T(), iterErr, http.StatusTooManyRequests)
	s.Equal(int64(2), requests.Load())
}

// TestListIterStopsOnCancellation verifies a context canceled part way through
// iteration is reported and stops any further request.
func (s *WebhookEventTestSuite) TestListIterStopsOnCancellation() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = fmt.Fprint(w, eventPage([]string{"e1"}, "cursor-2"))
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

	s.Equal([]string{"e1"}, seen)
	s.Require().Error(iterErr)
	s.Require().ErrorIs(iterErr, context.Canceled)
	s.Equal(int64(1), requests.Load(), "no page may be requested after cancellation")
}

// TestListIterEmptyFirstPage verifies an empty result set ends iteration
// immediately without yielding anything.
func (s *WebhookEventTestSuite) TestListIterEmptyFirstPage() {
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
func (s *WebhookEventTestSuite) TestListIterStopsOnEmptyPage() {
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, eventPage([]string{"e1"}, "cursor-2"))
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
	s.Equal([]string{"e1"}, seen)
	s.Equal(int64(2), requests.Load(), "an empty page ends iteration even when a cursor is offered")
}

// TestListIterKeepsCallerOptions verifies the caller's filters are sent on every
// page, and that the page cursor overrides any cursor the caller supplied.
func (s *WebhookEventTestSuite) TestListIterKeepsCallerOptions() {
	cursors := make([]string, 0, 2)
	successes := make([]string, 0, 2)
	var requests atomic.Int64

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		cursors = append(cursors, req.URL.Query().Get("starting_after"))
		successes = append(successes, req.URL.Query().Get("success"))

		if req.URL.Query().Get("starting_after") == "caller-cursor" {
			_, _ = fmt.Fprint(w, eventPage([]string{"e1"}, "cursor-2"))
			return
		}

		_, _ = fmt.Fprint(w, eventPage([]string{"e2"}, ""))
	})

	callerOpts := []webhookevent.ListOption{
		webhookevent.WithSuccess(false),
		webhookevent.WithStartingAfter("caller-cursor"),
	}

	seen := make([]string, 0, 2)
	for got, err := range s.svc().ListIter(context.Background(), callerOpts...) {
		if err != nil {
			break
		}
		seen = append(seen, got.ID)
	}

	s.Equal([]string{"e1", "e2"}, seen)
	s.Equal(int64(2), requests.Load())
	s.Equal([]string{"caller-cursor", "cursor-2"}, cursors, "the page cursor must override the caller's")
	s.Equal([]string{"false", "false"}, successes, "the caller's filters must survive every page")

	s.Require().Len(callerOpts, 2)
}

// eventPage renders one page of a list response containing the given event
// identifiers, followed by the cursor of the next page when there is one.
func eventPage(ids []string, nextCursor string) string {
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, fmt.Sprintf(
			`{"id":%q,"organization_id":"org-1","webhook_url":"https://example.com/hook",`+
				`"timestamp_created":"2026-08-01T10:00:00.000Z","timestamp_created_date":"2026-08-01",`+
				`"success":true,"retry_count":0,"will_retry":false}`,
			id,
		))
	}

	return instantlytest.Page(items, nextCursor)
}
