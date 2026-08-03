package dfy_test

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/mrz1836/go-instantly/dfy"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// TestListIterWalksEveryPage verifies the order iterator stitches every page
// into a single sequence, in order, and stops once the cursor runs out.
func (s *DFYTestSuite) TestListIterWalksEveryPage() {
	var requests atomic.Int64

	s.Router.Get(ordersPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		switch req.URL.Query().Get("starting_after") {
		case "":
			_, _ = fmt.Fprint(w, orderPage([]string{"d1", "d2"}, "cursor-2"))
		case "cursor-2":
			_, _ = fmt.Fprint(w, orderPage([]string{"d3", "d4"}, "cursor-3"))
		default:
			_, _ = fmt.Fprint(w, orderPage([]string{"d5"}, ""))
		}
	})

	seen := make([]string, 0, 5)
	var iterErr error

	for got, err := range s.svc().ListIter(context.Background()) {
		if err != nil {
			iterErr = err
			break
		}
		seen = append(seen, got.Domain)
	}

	s.Require().NoError(iterErr)
	s.Equal([]string{"d1", "d2", "d3", "d4", "d5"}, seen)
	s.Equal(int64(3), requests.Load(), "one request per page, and no request after the last cursor")
}

// TestListIterStopsOnBreak verifies breaking out of the loop issues no further
// requests.
func (s *DFYTestSuite) TestListIterStopsOnBreak() {
	var requests atomic.Int64

	s.Router.Get(ordersPath, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = fmt.Fprint(w, orderPage([]string{"d1", "d2", "d3"}, "cursor-2"))
	})

	seen := make([]string, 0, 2)

	for got, err := range s.svc().ListIter(context.Background()) {
		if err != nil {
			break
		}

		seen = append(seen, got.Domain)
		if len(seen) == 2 {
			break
		}
	}

	s.Equal([]string{"d1", "d2"}, seen)
	s.Equal(int64(1), requests.Load(), "breaking mid-page must not fetch the next page")
}

// TestListIterStopsOnError verifies a failure on a later page is yielded with a
// nil order and ends the iteration.
func (s *DFYTestSuite) TestListIterStopsOnError() {
	var requests atomic.Int64

	s.Router.Get(ordersPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, orderPage([]string{"d1"}, "cursor-2"))
			return
		}

		instantlytest.WriteAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "rate limited")
	})

	seen := make([]string, 0, 1)
	var iterErr error
	var errOrder *dfy.Order
	yields := 0

	for got, err := range s.svc().ListIter(context.Background()) {
		yields++

		if err != nil {
			iterErr = err
			errOrder = got
			continue
		}

		seen = append(seen, got.Domain)
	}

	s.Equal([]string{"d1"}, seen)
	s.Equal(2, yields, "iteration must end at the failure rather than retrying the page")
	s.Nil(errOrder, "an error is yielded with no order")

	instantlytest.AssertAPIError(s.T(), iterErr, http.StatusTooManyRequests)
	s.Equal(int64(2), requests.Load())
}

// TestListIterStopsOnCancellation verifies a context canceled part way through
// iteration is reported and stops any further request.
func (s *DFYTestSuite) TestListIterStopsOnCancellation() {
	var requests atomic.Int64

	s.Router.Get(ordersPath, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = fmt.Fprint(w, orderPage([]string{"d1"}, "cursor-2"))
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

		seen = append(seen, got.Domain)
		cancel()
	}

	s.Equal([]string{"d1"}, seen)
	s.Require().Error(iterErr)
	s.Require().ErrorIs(iterErr, context.Canceled)
	s.Equal(int64(1), requests.Load(), "no page may be requested after cancellation")
}

// TestListIterEmptyFirstPage verifies an empty result set ends iteration
// immediately without yielding anything.
func (s *DFYTestSuite) TestListIterEmptyFirstPage() {
	var requests atomic.Int64

	s.Router.Get(ordersPath, func(w http.ResponseWriter, _ *http.Request) {
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
func (s *DFYTestSuite) TestListIterStopsOnEmptyPage() {
	var requests atomic.Int64

	s.Router.Get(ordersPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, orderPage([]string{"d1"}, "cursor-2"))
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
		seen = append(seen, got.Domain)
	}

	s.Require().NoError(iterErr)
	s.Equal([]string{"d1"}, seen)
	s.Equal(int64(2), requests.Load(), "an empty page ends iteration even when a cursor is offered")
}

// TestListIterKeepsCallerOptions verifies the caller's filters are sent on every
// page, and that the page cursor overrides any cursor the caller supplied.
func (s *DFYTestSuite) TestListIterKeepsCallerOptions() {
	cursors := make([]string, 0, 2)
	limits := make([]string, 0, 2)
	var requests atomic.Int64

	s.Router.Get(ordersPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		cursors = append(cursors, req.URL.Query().Get("starting_after"))
		limits = append(limits, req.URL.Query().Get("limit"))

		if req.URL.Query().Get("starting_after") == "caller-cursor" {
			_, _ = fmt.Fprint(w, orderPage([]string{"d1"}, "cursor-2"))
			return
		}

		_, _ = fmt.Fprint(w, orderPage([]string{"d2"}, ""))
	})

	callerOpts := []dfy.ListOption{
		dfy.WithLimit(10),
		dfy.WithStartingAfter("caller-cursor"),
	}

	seen := make([]string, 0, 2)
	for got, err := range s.svc().ListIter(context.Background(), callerOpts...) {
		if err != nil {
			break
		}
		seen = append(seen, got.Domain)
	}

	s.Equal([]string{"d1", "d2"}, seen)
	s.Equal(int64(2), requests.Load())
	s.Equal([]string{"caller-cursor", "cursor-2"}, cursors, "the page cursor must override the caller's")
	s.Equal([]string{"10", "10"}, limits, "the caller's filters must survive every page")

	s.Require().Len(callerOpts, 2)
}

// TestListAccountsIterWalksPages verifies the account iterator stitches every
// page and threads the bound withPasswords choice onto each request.
func (s *DFYTestSuite) TestListAccountsIterWalksPages() {
	cursors := make([]string, 0, 2)
	passwordFlags := make([]string, 0, 2)
	var requests atomic.Int64

	s.Router.Get(accountsPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		cursors = append(cursors, req.URL.Query().Get("starting_after"))
		passwordFlags = append(passwordFlags, req.URL.Query().Get("with_passwords"))

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, accountPage([]string{"a1"}, "cursor-2"))
			return
		}

		_, _ = fmt.Fprint(w, accountPage([]string{"a2"}, ""))
	})

	seen := make([]string, 0, 2)
	for got, err := range s.svc().ListAccountsIter(context.Background(), true) {
		if err != nil {
			break
		}
		seen = append(seen, got.ID)
	}

	s.Equal([]string{"a1", "a2"}, seen)
	s.Equal(int64(2), requests.Load())
	s.Equal([]string{"", "cursor-2"}, cursors)
	s.Equal([]string{"true", "true"}, passwordFlags, "the bound withPasswords choice must survive every page")
}

// TestListAccountsIterKeepsCallerOptions verifies the caller's filters ride on
// every account page alongside the bound withPasswords choice, and the page
// cursor overrides any cursor the caller supplied.
func (s *DFYTestSuite) TestListAccountsIterKeepsCallerOptions() {
	cursors := make([]string, 0, 2)
	limits := make([]string, 0, 2)
	passwordFlags := make([]string, 0, 2)
	var requests atomic.Int64

	s.Router.Get(accountsPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		cursors = append(cursors, req.URL.Query().Get("starting_after"))
		limits = append(limits, req.URL.Query().Get("limit"))
		passwordFlags = append(passwordFlags, req.URL.Query().Get("with_passwords"))

		if req.URL.Query().Get("starting_after") == "caller-cursor" {
			_, _ = fmt.Fprint(w, accountPage([]string{"a1"}, "cursor-2"))
			return
		}

		_, _ = fmt.Fprint(w, accountPage([]string{"a2"}, ""))
	})

	callerOpts := []dfy.ListOption{
		dfy.WithLimit(15),
		dfy.WithStartingAfter("caller-cursor"),
	}

	seen := make([]string, 0, 2)
	for got, err := range s.svc().ListAccountsIter(context.Background(), false, callerOpts...) {
		if err != nil {
			break
		}
		seen = append(seen, got.ID)
	}

	s.Equal([]string{"a1", "a2"}, seen)
	s.Equal(int64(2), requests.Load())
	s.Equal([]string{"caller-cursor", "cursor-2"}, cursors, "the page cursor must override the caller's")
	s.Equal([]string{"15", "15"}, limits, "the caller's filters must survive every page")
	s.Equal([]string{"false", "false"}, passwordFlags, "the bound withPasswords choice must survive every page")

	s.Require().Len(callerOpts, 2)
}

// TestListAccountsIterStopsOnError verifies a failure on a later account page is
// yielded with a nil account and ends the iteration.
func (s *DFYTestSuite) TestListAccountsIterStopsOnError() {
	var requests atomic.Int64

	s.Router.Get(accountsPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, accountPage([]string{"a1"}, "cursor-2"))
			return
		}

		instantlytest.WriteAPIErrorEnvelope(w, http.StatusTooManyRequests, "Too Many Requests", "rate limited")
	})

	seen := make([]string, 0, 1)
	var iterErr error
	var errAccount *dfy.OrderedAccount
	yields := 0

	for got, err := range s.svc().ListAccountsIter(context.Background(), true) {
		yields++

		if err != nil {
			iterErr = err
			errAccount = got
			continue
		}

		seen = append(seen, got.ID)
	}

	s.Equal([]string{"a1"}, seen)
	s.Equal(2, yields, "iteration must end at the failure rather than retrying the page")
	s.Nil(errAccount, "an error is yielded with no account")

	instantlytest.AssertAPIError(s.T(), iterErr, http.StatusTooManyRequests)
	s.Equal(int64(2), requests.Load())
}

// orderPage renders one page of orders whose domains are the given identifiers.
func orderPage(domains []string, nextCursor string) string {
	items := make([]string, 0, len(domains))
	for _, domain := range domains {
		items = append(items, fmt.Sprintf(
			`{"workspace_id":"ws-1","domain":%q,"timestamp_created":"2026-08-01T10:00:00.000Z"}`,
			domain,
		))
	}

	return instantlytest.Page(items, nextCursor)
}

// accountPage renders one page of ordered accounts with the given identifiers.
func accountPage(ids []string, nextCursor string) string {
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, fmt.Sprintf(
			`{"id":%q,"domain":"example.com","email":"user@example.com","email_provider":1,`+
				`"first_name":"John","last_name":"Doe","is_pre_warmed_up":false,`+
				`"timestamp_cancelled":"","timestamp_created":"2026-08-01T10:00:00.000Z"}`,
			id,
		))
	}

	return instantlytest.Page(items, nextCursor)
}
