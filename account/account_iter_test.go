package account_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/mrz1836/go-instantly/account"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// TestListIterWalksEveryPage verifies the iterator stitches pages together and
// carries the caller's filters onto every page.
func (s *AccountTestSuite) TestListIterWalksEveryPage() {
	var requests atomic.Int64
	limits := make([]string, 0, 2)

	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		limits = append(limits, req.URL.Query().Get("limit"))

		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, accountPage([]string{"a@x.com", "b@x.com"}, "cursor-2"))
			return
		}

		_, _ = fmt.Fprint(w, accountPage([]string{"c@x.com"}, ""))
	})

	seen := make([]string, 0, 3)
	for got, err := range s.svc().ListIter(context.Background(), account.WithLimit(2)) {
		s.Require().NoError(err)
		seen = append(seen, got.Email)
	}

	s.Equal([]string{"a@x.com", "b@x.com", "c@x.com"}, seen)
	s.Equal(int64(2), requests.Load())
	s.Equal([]string{"2", "2"}, limits, "the caller's filters survive every page")
}

// TestListIterStopsOnError verifies a failure ends the iteration with a nil
// account.
func (s *AccountTestSuite) TestListIterStopsOnError() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Query().Get("starting_after") == "" {
			_, _ = fmt.Fprint(w, accountPage([]string{"a@x.com"}, "cursor-2"))
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
		seen = append(seen, got.Email)
	}

	s.Equal([]string{"a@x.com"}, seen)
	instantlytest.AssertAPIError(s.T(), iterErr, http.StatusTooManyRequests)
}

// accountPage renders one page of a list response for the given account emails.
func accountPage(emails []string, nextCursor string) string {
	items := make([]string, 0, len(emails))
	for _, email := range emails {
		items = append(items, fmt.Sprintf(
			`{"email":%q,"first_name":"A","last_name":"B","organization":"org-1",`+
				`"timestamp_created":"2026-08-01T10:00:00.000Z",`+
				`"timestamp_updated":"2026-08-01T11:00:00.000Z","status":1,"warmup_status":1,`+
				`"provider_code":2,"setup_pending":false,"is_managed_account":false}`,
			email,
		))
	}

	if nextCursor == "" {
		return fmt.Sprintf(`{"items":[%s]}`, strings.Join(items, ","))
	}

	return fmt.Sprintf(`{"items":[%s],"next_starting_after":%q}`, strings.Join(items, ","), nextCursor)
}
