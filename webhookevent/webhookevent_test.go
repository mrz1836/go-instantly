package webhookevent_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/webhookevent"
)

// Router patterns and identifiers the webhook-event endpoints are exercised
// with. The patterns carry the full request path, including the /api/v2 prefix.
const (
	// listPath is the list/collection endpoint.
	listPath = "/api/v2/webhook-events"

	// idPattern is the router pattern for the single-event endpoint.
	idPattern = "/api/v2/webhook-events/:id"

	// summaryPath is the summary endpoint.
	summaryPath = "/api/v2/webhook-events/summary"

	// summaryByDatePath is the summary-by-date endpoint.
	summaryByDatePath = "/api/v2/webhook-events/summary-by-date"

	// eventID identifies the event the single-event endpoint operates on.
	eventID = "we-1"
)

// eventFixture is a spec-shaped webhook event with every documented field
// populated, including the nullable ones and the raw payload.
const eventFixture = `{
	"id": "we-1",
	"organization_id": "org-1",
	"webhook_url": "https://example.com/hook",
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"timestamp_created_date": "2026-08-01",
	"success": true,
	"retry_count": 1,
	"will_retry": false,
	"lead_email": "lead@example.com",
	"status_code": 200,
	"response_time_ms": 42,
	"error_message": "none",
	"retry_group_id": "grp-1",
	"retry_successful": true,
	"timestamp_next_retry": "2026-08-01T11:00:00.000Z",
	"payload": {"event": "reply_received"}
}`

// eventFixtureNulls is the same event with every nullable field explicitly null,
// so an absent value stays distinguishable from a zero value.
const eventFixtureNulls = `{
	"id": "we-2",
	"organization_id": "org-1",
	"webhook_url": "https://example.com/hook",
	"timestamp_created": "2026-08-01T12:00:00.000Z",
	"timestamp_created_date": "2026-08-01",
	"success": false,
	"retry_count": 0,
	"will_retry": true,
	"lead_email": null,
	"status_code": null,
	"response_time_ms": null,
	"error_message": null,
	"retry_group_id": null,
	"retry_successful": null,
	"timestamp_next_retry": null
}`

// WebhookEventTestSuite exercises the Webhook Event API service against the mock
// router.
type WebhookEventTestSuite struct {
	instantlytest.Suite
}

// TestWebhookEventSuite runs the Webhook Event API suite.
func TestWebhookEventSuite(t *testing.T) {
	suite.Run(t, new(WebhookEventTestSuite))
}

// TestList verifies the options are sent and a page decodes, including
// nullable-vs-zero fields and the raw payload.
func (s *WebhookEventTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(listPath, req.URL.Path)
		s.Equal("50", req.URL.Query().Get("limit"))
		s.Equal("false", req.URL.Query().Get("success"))
		s.Equal("2026-08-01", req.URL.Query().Get("from"))

		_, _ = w.Write([]byte(instantlytest.Page([]string{eventFixture, eventFixtureNulls}, "cursor-2")))
	})

	page, err := s.svc().List(context.Background(),
		webhookevent.WithLimit(50),
		webhookevent.WithSuccess(false),
		webhookevent.WithFrom("2026-08-01"),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)

	populated := page.Items[0]
	s.Equal(eventID, populated.ID)
	s.True(populated.Success)
	s.InDelta(1, populated.RetryCount, 0)
	s.Require().NotNil(populated.StatusCode)
	s.InDelta(200, *populated.StatusCode, 0)
	s.Require().NotNil(populated.RetrySuccessful)
	s.True(*populated.RetrySuccessful)
	s.JSONEq(`{"event":"reply_received"}`, string(populated.Payload))

	// Nullable fields stay nil rather than collapsing to a zero value.
	bare := page.Items[1]
	s.False(bare.Success)
	s.True(bare.WillRetry)
	s.Nil(bare.LeadEmail)
	s.Nil(bare.StatusCode)
	s.Nil(bare.ResponseTimeMS)
	s.Nil(bare.RetrySuccessful)
	s.Nil(bare.TimestampNextRetry)
	s.Empty(bare.Payload)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string, even
// with a nil option.
func (s *WebhookEventTestSuite) TestListWithoutOptions() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery, "an unfiltered list must not send an empty query string")
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background(), nil)

	s.Require().NoError(err)
	s.Require().NotNil(page)
	s.Empty(page.Items)
}

// TestGet verifies a single event decodes.
func (s *WebhookEventTestSuite) TestGet() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(eventID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(eventFixture))
	})

	got, err := s.svc().Get(context.Background(), eventID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(eventID, got.ID)
	s.Require().NotNil(got.LeadEmail)
	s.Equal("lead@example.com", *got.LeadEmail)
}

// TestSummary verifies the window is sent and the aggregate summary decodes.
func (s *WebhookEventTestSuite) TestSummary() {
	s.Router.Get(summaryPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(summaryPath, req.URL.Path)
		s.Equal("2026-08-01", req.URL.Query().Get("from"))
		s.Equal("2026-08-31", req.URL.Query().Get("to"))

		_, _ = w.Write([]byte(
			`{"total_events":100,"successful_events":90,"failed_events":10,` +
				`"success_rate":0.9,"failure_rate":0.1}`,
		))
	})

	got, err := s.svc().Summary(context.Background(), "2026-08-01", "2026-08-31")

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.InDelta(100, got.TotalEvents, 0)
	s.InDelta(90, got.SuccessfulEvents, 0)
	s.InDelta(0.9, got.SuccessRate, 0)
}

// TestSummaryOmitsEmptyWindow verifies an unbounded summary sends no query
// string.
func (s *WebhookEventTestSuite) TestSummaryOmitsEmptyWindow() {
	s.Router.Get(summaryPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery, "an unbounded window must not send a query string")
		_, _ = w.Write([]byte(
			`{"total_events":0,"successful_events":0,"failed_events":0,"success_rate":0,"failure_rate":0}`,
		))
	})

	got, err := s.svc().Summary(context.Background(), "", "")

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Zero(got.TotalEvents)
}

// TestSummaryByDate verifies the wrapped items are unwrapped to a slice, and
// that only the supplied window bound is sent.
func (s *WebhookEventTestSuite) TestSummaryByDate() {
	s.Router.Get(summaryByDatePath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(summaryByDatePath, req.URL.Path)
		s.Equal("2026-08-31", req.URL.Query().Get("to"))
		s.Empty(req.URL.Query().Get("from"), "an unset window bound must not be sent")

		_, _ = w.Write([]byte(
			`{"items":[{"date":"2026-08-01","total_events":10,"successful_events":9,` +
				`"failed_events":1,"success_rate":0.9}]}`,
		))
	})

	got, err := s.svc().SummaryByDate(context.Background(), "", "2026-08-31")

	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal("2026-08-01", got[0].Date)
	s.InDelta(10, got[0].TotalEvents, 0)
	s.InDelta(0.9, got[0].SuccessRate, 0)
}

// TestPathParametersAreEscaped verifies a caller-supplied identifier cannot
// rewrite the request path.
func (s *WebhookEventTestSuite) TestPathParametersAreEscaped() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, eventFixture), nil
			},
		)},
	))

	_, err := webhookevent.New(client).Get(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/webhook-events/..%2Fadmin%3Fx=1", requestURI)
}

// TestListOptions verifies each documented list query parameter is rendered by
// exactly one option, under the key and value the API expects.
func (s *WebhookEventTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option webhookevent.ListOption
		key    string
		value  string
	}{
		{"limit", webhookevent.WithLimit(50), "limit", "50"},
		{"starting after", webhookevent.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"success", webhookevent.WithSuccess(true), "success", "true"},
		{"from", webhookevent.WithFrom("2026-08-01"), "from", "2026-08-01"},
		{"to", webhookevent.WithTo("2026-08-31"), "to", "2026-08-31"},
		{"search", webhookevent.WithSearch("lead@example.com"), "search", "lead@example.com"},
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
func (s *WebhookEventTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "list", Method: http.MethodGet, Path: listPath, Status: http.StatusTooManyRequests,
			Call: func() error { _, err := svc.List(ctx); return err },
		},
		{
			Name: "get", Method: http.MethodGet, Path: idPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Get(ctx, "missing"); return err },
		},
		{
			Name: "summary", Method: http.MethodGet, Path: summaryPath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.Summary(ctx, "", ""); return err },
		},
		{
			Name: "summaryByDate", Method: http.MethodGet, Path: summaryByDatePath, Status: http.StatusForbidden,
			Call: func() error { _, err := svc.SummaryByDate(ctx, "", ""); return err },
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor parses a valid
// timestamp and reports an error for an unparseable one.
func (s *WebhookEventTestSuite) TestParsedTimestampCreated() {
	got, err := (&webhookevent.Event{TimestampCreated: "2026-08-01T10:00:00.000Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&webhookevent.Event{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// svc builds a Webhook Event service pointed at the suite's mock client.
func (s *WebhookEventTestSuite) svc() *webhookevent.Service {
	return webhookevent.New(s.Client)
}
