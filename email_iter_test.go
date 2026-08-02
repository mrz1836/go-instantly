package instantly

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

// TestListEmailsIterWalksEveryPage verifies the iterator stitches every page
// into a single sequence, in order, and stops once the cursor runs out.
func (s *InstantlyTestSuite) TestListEmailsIterWalksEveryPage() {
	var requests atomic.Int64

	s.mux.Get(testPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		switch req.URL.Query().Get("starting_after") {
		case "":
			_, _ = fmt.Fprint(w, emailPage([]string{"e1", "e2"}, "cursor-2"))
		case "cursor-2":
			_, _ = fmt.Fprint(w, emailPage([]string{"e3", "e4"}, "cursor-3"))
		default:
			// The last page carries no cursor, which ends the iteration.
			_, _ = fmt.Fprint(w, emailPage([]string{"e5"}, ""))
		}
	})

	seen := make([]string, 0, 5)
	var iterErr error

	for email, err := range s.client.ListEmailsIter(context.Background()) {
		if err != nil {
			iterErr = err
			break
		}
		seen = append(seen, email.ID)
	}

	s.Require().NoError(iterErr)
	s.Equal([]string{"e1", "e2", "e3", "e4", "e5"}, seen)
	s.Equal(int64(3), requests.Load(), "one request per page, and no request after the last cursor")
}

// TestListEmailsIterStopsOnBreak verifies breaking out of the loop issues no
// further requests, so a consumer is never billed for pages it never reads.
func (s *InstantlyTestSuite) TestListEmailsIterStopsOnBreak() {
	var requests atomic.Int64

	s.mux.Get(testPath, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = fmt.Fprint(w, emailPage([]string{"e1", "e2", "e3"}, "cursor-2"))
	})

	seen := make([]string, 0, 2)

	for email, err := range s.client.ListEmailsIter(context.Background()) {
		if err != nil {
			break
		}

		seen = append(seen, email.ID)
		if len(seen) == 2 {
			break
		}
	}

	s.Equal([]string{"e1", "e2"}, seen)
	s.Equal(int64(1), requests.Load(), "breaking mid-page must not fetch the next page")
}

// TestListEmailsIterStopsOnError verifies a failure on a later page is yielded
// with a nil email and ends the iteration.
func (s *InstantlyTestSuite) TestListEmailsIterStopsOnError() {
	var requests atomic.Int64

	s.mux.Get(testPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, emailPage([]string{"e1"}, "cursor-2"))
			return
		}

		writeAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "20 requests per minute")
	})

	seen := make([]string, 0, 1)
	var iterErr error
	var errEmail *Email
	yields := 0

	for email, err := range s.client.ListEmailsIter(context.Background()) {
		yields++

		if err != nil {
			iterErr = err
			errEmail = email
			continue
		}

		seen = append(seen, email.ID)
	}

	s.Equal([]string{"e1"}, seen)
	s.Equal(2, yields, "iteration must end at the failure rather than retrying the page")
	s.Nil(errEmail, "an error is yielded with no email")

	assertAPIError(s, iterErr, http.StatusTooManyRequests)
	s.Equal(int64(2), requests.Load())
}

// TestListEmailsIterStopsOnCancellation verifies a context canceled part way
// through iteration is reported and stops any further request.
func (s *InstantlyTestSuite) TestListEmailsIterStopsOnCancellation() {
	var requests atomic.Int64

	s.mux.Get(testPath, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = fmt.Fprint(w, emailPage([]string{"e1"}, "cursor-2"))
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	seen := make([]string, 0, 1)
	var iterErr error

	for email, err := range s.client.ListEmailsIter(ctx) {
		if err != nil {
			iterErr = err
			break
		}

		seen = append(seen, email.ID)

		// Cancel between pages: the cursor is still set, so without the check
		// the iterator would go on to request the next page.
		cancel()
	}

	s.Equal([]string{"e1"}, seen)
	s.Require().Error(iterErr)
	s.Require().ErrorIs(iterErr, context.Canceled)
	s.Equal(int64(1), requests.Load(), "no page may be requested after cancellation")
}

// TestListEmailsIterEmptyFirstPage verifies an empty result set ends iteration
// immediately without yielding anything.
func (s *InstantlyTestSuite) TestListEmailsIterEmptyFirstPage() {
	var requests atomic.Int64

	s.mux.Get(testPath, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	yields := 0
	for range s.client.ListEmailsIter(context.Background()) {
		yields++
	}

	s.Zero(yields, "an empty result set yields nothing at all")
	s.Equal(int64(1), requests.Load())
}

// TestListEmailsIterStopsOnEmptyPage verifies an API that keeps handing back a
// cursor cannot drive an unbounded number of requests.
func (s *InstantlyTestSuite) TestListEmailsIterStopsOnEmptyPage() {
	var requests atomic.Int64

	s.mux.Get(testPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, emailPage([]string{"e1"}, "cursor-2"))
			return
		}

		// An empty page that still advertises a cursor: honoring the cursor
		// here would loop forever against a rate-limited endpoint.
		_, _ = w.Write([]byte(`{"items":[],"next_starting_after":"cursor-3"}`))
	})

	seen := make([]string, 0, 1)
	var iterErr error

	for email, err := range s.client.ListEmailsIter(context.Background()) {
		if err != nil {
			iterErr = err
			break
		}
		seen = append(seen, email.ID)
	}

	s.Require().NoError(iterErr)
	s.Equal([]string{"e1"}, seen)
	s.Equal(int64(2), requests.Load(), "an empty page ends iteration even when a cursor is offered")
}

// TestListEmailsIterKeepsCallerOptions verifies the caller's filters are sent on
// every page, and that the page cursor overrides any cursor the caller supplied
// rather than re-requesting the same page forever.
func (s *InstantlyTestSuite) TestListEmailsIterKeepsCallerOptions() {
	cursors := make([]string, 0, 2)
	limits := make([]string, 0, 2)
	var requests atomic.Int64

	s.mux.Get(testPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		cursors = append(cursors, req.URL.Query().Get("starting_after"))
		limits = append(limits, req.URL.Query().Get("limit"))

		if req.URL.Query().Get("starting_after") == "caller-cursor" {
			_, _ = fmt.Fprint(w, emailPage([]string{"e1"}, "cursor-2"))
			return
		}

		_, _ = fmt.Fprint(w, emailPage([]string{"e2"}, ""))
	})

	callerOpts := []EmailListOption{
		WithEmailLimit(2),
		WithEmailStartingAfter("caller-cursor"),
	}

	seen := make([]string, 0, 2)
	for email, err := range s.client.ListEmailsIter(context.Background(), callerOpts...) {
		if err != nil {
			break
		}
		seen = append(seen, email.ID)
	}

	s.Equal([]string{"e1", "e2"}, seen)
	s.Equal(int64(2), requests.Load())
	s.Equal([]string{"caller-cursor", "cursor-2"}, cursors, "the page cursor must override the caller's")
	s.Equal([]string{"2", "2"}, limits, "the caller's filters must survive every page")

	// The caller's slice is copied before paging, so it is never mutated.
	s.Require().Len(callerOpts, 2)
}

// emailPage renders one page of a list response containing the given email
// identifiers, followed by the cursor of the next page when there is one.
func emailPage(ids []string, nextCursor string) string {
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, fmt.Sprintf(
			`{"id":%q,"timestamp_created":"2026-08-01T10:00:00.000Z",`+
				`"timestamp_email":"2026-08-01T09:59:00.000Z","message_id":"<%s@mail.example.com>",`+
				`"subject":%q,"to_address_email_list":%q,"body":{"html":"<p>Hello</p>"},`+
				`"organization_id":"org-uuid-1","eaccount":%q}`,
			id, id, testSubject, testLeadEmail, testEAccount,
		))
	}

	if nextCursor == "" {
		return fmt.Sprintf(`{"items":[%s]}`, strings.Join(items, ","))
	}

	return fmt.Sprintf(
		`{"items":[%s],"next_starting_after":%q}`, strings.Join(items, ","), nextCursor,
	)
}
