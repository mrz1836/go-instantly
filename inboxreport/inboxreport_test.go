package inboxreport_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/inboxreport"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// Router patterns and identifiers the inbox-placement-report endpoints are
// exercised with. The patterns carry the full request path, including the
// /api/v2 prefix.
const (
	// listPath is the list/collection endpoint.
	listPath = "/api/v2/inbox-placement-reports"

	// idPattern is the router pattern for the single-report endpoint.
	idPattern = "/api/v2/inbox-placement-reports/:id"

	// testID identifies the test the reports belong to.
	testID = "test-uuid-1"

	// reportID identifies the single report.
	reportID = "report-uuid-1"
)

// reportFixture is a spec-shaped report with every documented field populated,
// including the nullable counts and the nested reports. The SpamAssassin per-rule
// score arrives as a string, which the raw payload keeps intact.
const reportFixture = `{
	"id": "report-uuid-1",
	"organization_id": "org-uuid-1",
	"test_id": "test-uuid-1",
	"timestamp_created": "2026-08-01T19:32:41.209Z",
	"timestamp_created_date": "2026-08-01",
	"domain": "example.com",
	"domain_ip": "203.0.113.10",
	"spam_assassin_score": 1.5,
	"domain_blacklist_count": 0,
	"domain_ip_blacklist_count": 1,
	"blacklist_report": {"listed": ["spamhaus"]},
	"spam_assassin_report": {
		"is_spam": false,
		"spam_score": 0,
		"report": [{"description": "BODY", "name": "HTML_MESSAGE", "score": "0.0"}]
	}
}`

// reportFixtureNulls is the same report with the nullable counts explicitly null,
// so an absent value stays distinguishable from a zero value.
const reportFixtureNulls = `{
	"id": "report-uuid-2",
	"organization_id": "org-uuid-1",
	"test_id": "test-uuid-1",
	"timestamp_created": "2026-08-01T20:00:00.000Z",
	"timestamp_created_date": "2026-08-01",
	"domain": "bare.example.com",
	"domain_ip": "203.0.113.11",
	"spam_assassin_score": 0,
	"domain_blacklist_count": null,
	"domain_ip_blacklist_count": null
}`

// InboxReportTestSuite exercises the Inbox Placement Report API service against
// the mock router.
type InboxReportTestSuite struct {
	instantlytest.Suite
}

// TestInboxReportSuite runs the Inbox Placement Report API suite.
func TestInboxReportSuite(t *testing.T) {
	suite.Run(t, new(InboxReportTestSuite))
}

// TestList verifies the required test_id filter and options are sent, and that a
// page decodes including nullable-vs-zero fields and the raw nested reports.
func (s *InboxReportTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodGet, req.Method)
		s.Equal(listPath, req.URL.Path)
		s.Equal(testID, req.URL.Query().Get("test_id"))
		s.Equal("true", req.URL.Query().Get("skip_spam_assassin_report"))

		_, _ = w.Write([]byte(instantlytest.Page([]string{reportFixture, reportFixtureNulls}, "cursor-2")))
	})

	page, err := s.svc().List(context.Background(), testID,
		inboxreport.WithSkipSpamAssassinReport(true),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)

	populated := page.Items[0]
	s.Equal(reportID, populated.ID)
	s.Equal("example.com", populated.Domain)
	s.InDelta(1.5, populated.SpamAssassinScore, 0)
	s.Require().NotNil(populated.DomainBlacklistCount)
	s.InDelta(0, *populated.DomainBlacklistCount, 0)
	s.Require().NotNil(populated.DomainIPBlacklistCount)
	s.InDelta(1, *populated.DomainIPBlacklistCount, 0)
	s.JSONEq(`{"listed":["spamhaus"]}`, string(populated.BlacklistReport))
	// The SpamAssassin per-rule score is a string, preserved verbatim.
	s.JSONEq(
		`{"is_spam":false,"spam_score":0,"report":[{"description":"BODY","name":"HTML_MESSAGE","score":"0.0"}]}`,
		string(populated.SpamAssassinReport),
	)

	// The nullable counts stay nil rather than collapsing to a zero value.
	bare := page.Items[1]
	s.Nil(bare.DomainBlacklistCount)
	s.Nil(bare.DomainIPBlacklistCount)
	s.Empty(bare.BlacklistReport)
}

// TestListSendsOnlyTestID verifies an options-free list sends the required
// test_id and nothing else.
func (s *InboxReportTestSuite) TestListSendsOnlyTestID() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, req.URL.Query().Get("test_id"))
		s.Len(req.URL.Query(), 1, "an options-free list sends only test_id")

		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background(), testID, nil)

	s.Require().NoError(err)
	s.Require().NotNil(page)
	s.Empty(page.Items)
}

// TestGet verifies a single report decodes.
func (s *InboxReportTestSuite) TestGet() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(reportID, instantlytest.PathParam(req, "id"))

		_, _ = w.Write([]byte(reportFixture))
	})

	got, err := s.svc().Get(context.Background(), reportID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(reportID, got.ID)
	s.Equal(testID, got.TestID)
	s.Equal("203.0.113.10", got.DomainIP)
}

// TestPathParametersAreEscaped verifies a caller-supplied identifier cannot
// rewrite the request path.
func (s *InboxReportTestSuite) TestPathParametersAreEscaped() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, reportFixture), nil
			},
		)},
	))

	_, err := inboxreport.New(client).Get(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/inbox-placement-reports/..%2Fadmin%3Fx=1", requestURI)
}

// TestListOptions verifies each documented list query parameter is rendered by
// exactly one option, under the key and value the API expects.
func (s *InboxReportTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option inboxreport.ListOption
		key    string
		value  string
	}{
		{"limit", inboxreport.WithLimit(50), "limit", "50"},
		{"starting after", inboxreport.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"date from", inboxreport.WithDateFrom("2026-08-01"), "date_from", "2026-08-01"},
		{"date to", inboxreport.WithDateTo("2026-08-31"), "date_to", "2026-08-31"},
		{
			"skip spam assassin report",
			inboxreport.WithSkipSpamAssassinReport(true), "skip_spam_assassin_report", "true",
		},
		{"skip blacklist report", inboxreport.WithSkipBlacklistReport(false), "skip_blacklist_report", "false"},
	}

	s.Require().Len(tests, 6, "every documented list query parameter needs an option")

	for _, test := range tests {
		s.Run(test.name, func() {
			q := instantly.NewQuery()
			test.option(q)

			s.Require().Equal(1, q.Len(), "an option must render exactly one query parameter")
			s.Equal(test.value, q.Get(test.key))
		})
	}
}

// TestFailures verifies every endpoint surfaces the documented API error.
func (s *InboxReportTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "list", Method: http.MethodGet, Path: listPath, Status: http.StatusTooManyRequests,
			Call: func() error { _, err := svc.List(ctx, testID); return err },
		},
		{
			Name: "get", Method: http.MethodGet, Path: idPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Get(ctx, "missing"); return err },
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor parses a valid
// timestamp and reports an error for an unparseable one.
func (s *InboxReportTestSuite) TestParsedTimestampCreated() {
	got, err := (&inboxreport.Report{TimestampCreated: "2026-08-01T19:32:41.209Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&inboxreport.Report{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// svc builds an Inbox Placement Report service pointed at the suite's mock
// client.
func (s *InboxReportTestSuite) svc() *inboxreport.Service {
	return inboxreport.New(s.Client)
}
