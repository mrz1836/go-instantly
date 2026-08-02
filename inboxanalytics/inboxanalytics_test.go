package inboxanalytics_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/inboxanalytics"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// Router patterns and identifiers the inbox-placement-analytics endpoints are
// exercised with. The patterns carry the full request path, including the
// /api/v2 prefix.
const (
	// listPath is the list/collection endpoint.
	listPath = "/api/v2/inbox-placement-analytics"

	// idPattern is the router pattern for the single-event endpoint.
	idPattern = "/api/v2/inbox-placement-analytics/:id"

	// statsByTestIDPath is the stats-by-test-id endpoint.
	statsByTestIDPath = "/api/v2/inbox-placement-analytics/stats-by-test-id"

	// insightsPath is the deliverability-insights endpoint.
	insightsPath = "/api/v2/inbox-placement-analytics/deliverability-insights"

	// statsByDatePath is the stats-by-date endpoint.
	statsByDatePath = "/api/v2/inbox-placement-analytics/stats-by-date"

	// testID identifies the test the events belong to.
	testID = "test-uuid-1"

	// eventID identifies the single placement event.
	eventID = "event-uuid-1"
)

// analyticsFixture is a received placement event with every documented field
// populated, including the recipient-side fields and the raw reports.
const analyticsFixture = `{
	"id": "event-uuid-1",
	"organization_id": "org-uuid-1",
	"test_id": "test-uuid-1",
	"timestamp_created": "2026-08-01T19:32:41.209Z",
	"timestamp_created_date": "2026-08-01",
	"record_type": 2,
	"recipient_email": "seed@instantly.ai",
	"recipient_esp": 1,
	"recipient_geo": 1,
	"recipient_type": 1,
	"sender_email": "sender@example.com",
	"sender_esp": 2,
	"is_spam": false,
	"has_category": false,
	"dkim_pass": true,
	"dmarc_pass": true,
	"spf_pass": true,
	"smtp_ip_blacklist_report": {"listed": []},
	"authentication_failure_results": {"dkim": "pass"}
}`

// analyticsFixtureNulls is a sent placement event with every recipient-side field
// explicitly null, so an absent value stays distinguishable from a zero value.
const analyticsFixtureNulls = `{
	"id": "event-uuid-2",
	"organization_id": "org-uuid-1",
	"test_id": "test-uuid-1",
	"timestamp_created": "2026-08-01T19:33:41.209Z",
	"timestamp_created_date": "2026-08-01",
	"record_type": 1,
	"recipient_email": null,
	"recipient_esp": null,
	"recipient_geo": null,
	"recipient_type": null,
	"sender_email": null,
	"sender_esp": null,
	"is_spam": null,
	"has_category": null,
	"dkim_pass": null,
	"dmarc_pass": null,
	"spf_pass": null
}`

// InboxAnalyticsTestSuite exercises the Inbox Placement Analytics API service
// against the mock router.
type InboxAnalyticsTestSuite struct {
	instantlytest.Suite
}

// TestInboxAnalyticsSuite runs the Inbox Placement Analytics API suite.
func TestInboxAnalyticsSuite(t *testing.T) {
	suite.Run(t, new(InboxAnalyticsTestSuite))
}

// TestList verifies the required test_id filter and options are sent, and that a
// page decodes including nullable-vs-zero fields and the enum values.
func (s *InboxAnalyticsTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodGet, req.Method)
		s.Equal(listPath, req.URL.Path)
		s.Equal(testID, req.URL.Query().Get("test_id"))
		s.Equal("50", req.URL.Query().Get("limit"))
		s.Equal("2026-08-01", req.URL.Query().Get("date_from"))

		_, _ = w.Write([]byte(instantlytest.Page([]string{analyticsFixture, analyticsFixtureNulls}, "cursor-2")))
	})

	page, err := s.svc().List(context.Background(), testID,
		inboxanalytics.WithLimit(50),
		inboxanalytics.WithDateFrom("2026-08-01"),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)

	received := page.Items[0]
	s.Equal(eventID, received.ID)
	s.Require().NotNil(received.RecordType)
	s.Equal(inboxanalytics.RecordTypeReceived, *received.RecordType)
	s.Require().NotNil(received.RecipientESP)
	s.Equal(inboxanalytics.ESPGoogle, *received.RecipientESP)
	s.Require().NotNil(received.SenderESP)
	s.Equal(inboxanalytics.ESPMicrosoft, *received.SenderESP)
	s.Require().NotNil(received.SPFPass)
	s.True(*received.SPFPass)
	s.JSONEq(`{"listed":[]}`, string(received.SMTPIPBlacklistReport))

	// The recipient-side fields of a sent event stay nil rather than collapsing
	// to a zero value.
	sent := page.Items[1]
	s.Require().NotNil(sent.RecordType)
	s.Equal(inboxanalytics.RecordTypeSent, *sent.RecordType)
	s.Nil(sent.RecipientEmail)
	s.Nil(sent.RecipientESP)
	s.Nil(sent.IsSpam)
	s.Nil(sent.DKIMPass)
}

// TestListSendsOnlyTestID verifies an options-free list sends the required
// test_id and nothing else.
func (s *InboxAnalyticsTestSuite) TestListSendsOnlyTestID() {
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

// TestGet verifies a single placement event decodes.
func (s *InboxAnalyticsTestSuite) TestGet() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(eventID, instantlytest.PathParam(req, "id"))

		_, _ = w.Write([]byte(analyticsFixture))
	})

	got, err := s.svc().Get(context.Background(), eventID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(eventID, got.ID)
	s.Equal(testID, got.TestID)
	s.Require().NotNil(got.RecipientEmail)
	s.Equal("seed@instantly.ai", *got.RecipientEmail)
	s.JSONEq(`{"dkim":"pass"}`, string(got.AuthenticationFailureResults))
}

// TestStatsByTestID verifies the enum-slice body is sent as integer arrays and
// the bare-array response decodes.
func (s *InboxAnalyticsTestSuite) TestStatsByTestID() {
	s.Router.Post(statsByTestIDPath, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal([]any{testID}, received["test_ids"])
		s.Equal([]any{1.0, 2.0}, received["recipient_esp"])
		s.Equal([]any{1.0}, received["recipient_geo"])

		_, _ = w.Write([]byte(
			`[{"test_id":"test-uuid-1","count":10,"spam_count":2,"spam_percent":20,` +
				`"inbox_count":7,"inbox_percent":70,"category_count":1,"category_percent":10}]`,
		))
	})

	got, err := s.svc().StatsByTestID(context.Background(), inboxanalytics.StatsByTestIDRequest{
		TestIDs:      []string{testID},
		RecipientESP: []inboxanalytics.ESP{inboxanalytics.ESPGoogle, inboxanalytics.ESPMicrosoft},
		RecipientGeo: []inboxanalytics.Geo{inboxanalytics.GeoUS},
	})

	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal(testID, got[0].TestID)
	s.InDelta(10, got[0].Count, 0)
	s.InDelta(70, got[0].InboxPercent, 0)
}

// TestDeliverabilityInsights verifies the body is sent and the nullable-field
// bare-array response decodes.
func (s *InboxAnalyticsTestSuite) TestDeliverabilityInsights() {
	s.Router.Post(insightsPath, func(w http.ResponseWriter, req *http.Request) {
		var received inboxanalytics.DeliverabilityInsightsRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(testID, received.TestID)
		if s.NotNil(received.ShowPrevious) {
			s.True(*received.ShowPrevious)
		}

		_, _ = w.Write([]byte(
			`[{"test_id":"test-uuid-1","recipient_esp":1,"from":"2026-08-01","to":"2026-08-31",` +
				`"inbox_percentage":70,"spam_percentage":20,"category_percentage":10,` +
				`"prev_inbox_percentage":null,"previous_from":null}]`,
		))
	})

	got, err := s.svc().DeliverabilityInsights(context.Background(), inboxanalytics.DeliverabilityInsightsRequest{
		TestID:       testID,
		ShowPrevious: instantly.Ptr(true),
		RecipientESP: []inboxanalytics.ESP{inboxanalytics.ESPLibero},
	})

	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal(testID, got[0].TestID)
	s.Require().NotNil(got[0].RecipientESP)
	s.Equal(inboxanalytics.ESPGoogle, *got[0].RecipientESP)
	s.Require().NotNil(got[0].InboxPercentage)
	s.InDelta(70, *got[0].InboxPercentage, 0)

	// The nullable comparison-window fields stay nil.
	s.Nil(got[0].PrevInboxPercentage)
	s.Nil(got[0].PreviousFrom)
}

// TestStatsByDate verifies the body is sent and the bare-array response decodes.
func (s *InboxAnalyticsTestSuite) TestStatsByDate() {
	s.Router.Post(statsByDatePath, func(w http.ResponseWriter, req *http.Request) {
		var received inboxanalytics.StatsByDateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(testID, received.TestID)
		s.Equal([]inboxanalytics.RecipientType{inboxanalytics.RecipientProfessional}, received.RecipientType)

		_, _ = w.Write([]byte(
			`[{"timestamp_created_date":"2026-08-01","sent_count":10,"received_count":9,` +
				`"spam_count":2,"inbox_count":6,"category_count":1}]`,
		))
	})

	got, err := s.svc().StatsByDate(context.Background(), inboxanalytics.StatsByDateRequest{
		TestID:        testID,
		RecipientType: []inboxanalytics.RecipientType{inboxanalytics.RecipientProfessional},
	})

	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal("2026-08-01", got[0].TimestampCreatedDate)
	s.InDelta(10, got[0].SentCount, 0)
	s.InDelta(9, got[0].ReceivedCount, 0)
}

// TestPathParametersAreEscaped verifies a caller-supplied identifier cannot
// rewrite the request path.
func (s *InboxAnalyticsTestSuite) TestPathParametersAreEscaped() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, analyticsFixture), nil
			},
		)},
	))

	_, err := inboxanalytics.New(client).Get(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/inbox-placement-analytics/..%2Fadmin%3Fx=1", requestURI)
}

// TestListOptions verifies each documented list query parameter is rendered by
// exactly one option, under the key and value the API expects.
func (s *InboxAnalyticsTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option inboxanalytics.ListOption
		key    string
		value  string
	}{
		{"limit", inboxanalytics.WithLimit(50), "limit", "50"},
		{"starting after", inboxanalytics.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"date from", inboxanalytics.WithDateFrom("2026-08-01"), "date_from", "2026-08-01"},
		{"date to", inboxanalytics.WithDateTo("2026-08-31"), "date_to", "2026-08-31"},
		{"recipient geo", inboxanalytics.WithRecipientGeo("1,2"), "recipient_geo", "1,2"},
		{"recipient type", inboxanalytics.WithRecipientType("1"), "recipient_type", "1"},
		{"recipient esp", inboxanalytics.WithRecipientESP("1,2,8"), "recipient_esp", "1,2,8"},
		{"sender email", inboxanalytics.WithSenderEmail("sender@x.com"), "sender_email", "sender@x.com"},
	}

	s.Require().Len(tests, 8, "every documented list query parameter needs an option")

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
func (s *InboxAnalyticsTestSuite) TestFailures() {
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
		{
			Name: "statsByTestID", Method: http.MethodPost, Path: statsByTestIDPath, Status: http.StatusUnauthorized,
			Call: func() error {
				_, err := svc.StatsByTestID(ctx, inboxanalytics.StatsByTestIDRequest{})
				return err
			},
		},
		{
			Name: "insights", Method: http.MethodPost, Path: insightsPath, Status: http.StatusPaymentRequired,
			Call: func() error {
				_, err := svc.DeliverabilityInsights(ctx, inboxanalytics.DeliverabilityInsightsRequest{})
				return err
			},
		},
		{
			Name: "statsByDate", Method: http.MethodPost, Path: statsByDatePath, Status: http.StatusNotFound,
			Call: func() error {
				_, err := svc.StatsByDate(ctx, inboxanalytics.StatsByDateRequest{})
				return err
			},
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor parses a valid
// timestamp and reports an error for an unparseable one.
func (s *InboxAnalyticsTestSuite) TestParsedTimestampCreated() {
	got, err := (&inboxanalytics.Analytics{TimestampCreated: "2026-08-01T19:32:41.209Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&inboxanalytics.Analytics{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// svc builds an Inbox Placement Analytics service pointed at the suite's mock
// client.
func (s *InboxAnalyticsTestSuite) svc() *inboxanalytics.Service {
	return inboxanalytics.New(s.Client)
}
