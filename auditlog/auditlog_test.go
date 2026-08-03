package auditlog_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/auditlog"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// listPath is the router pattern for the audit-logs list endpoint.
const listPath = "/api/v2/audit-logs"

// logFixture is a spec-shaped audit log with every documented field populated,
// including the nullable ones.
const logFixture = `{
	"id": "log-1",
	"timestamp": "2026-08-01T10:00:00.000Z",
	"organization_id": "org-1",
	"activity_type": 1,
	"ip_address": "127.0.0.1",
	"from_api": true,
	"affected_count": 3,
	"api_key_id": "key-1",
	"campaign_id": "camp-1",
	"list_id": "list-1",
	"subsequence_id": "sub-1",
	"webhook_id": "hook-1",
	"user_agent": "Mozilla/5.0",
	"user_id": "user-1",
	"user_name": "John Doe",
	"audit_metadata": {"detail": "value"}
}`

// logFixtureNulls is a spec-shaped audit log with only the required fields set
// and every nullable field explicitly null.
const logFixtureNulls = `{
	"id": "log-2",
	"timestamp": "2026-08-01T11:00:00.000Z",
	"organization_id": "org-1",
	"activity_type": 38,
	"ip_address": "10.0.0.1",
	"from_api": false,
	"affected_count": null,
	"api_key_id": null,
	"campaign_id": null,
	"list_id": null,
	"subsequence_id": null,
	"webhook_id": null,
	"user_agent": null,
	"user_id": null,
	"user_name": null
}`

// AuditLogTestSuite exercises the Audit Log API service against the mock router.
type AuditLogTestSuite struct {
	instantlytest.Suite
}

// TestAuditLogSuite runs the Audit Log API suite.
func TestAuditLogSuite(t *testing.T) {
	suite.Run(t, new(AuditLogTestSuite))
}

// TestList verifies a page decodes, including the enum and metadata, and the
// options are sent under the keys the API expects.
func (s *AuditLogTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(listPath, req.URL.Path)
		s.Equal("50", req.URL.Query().Get("limit"))
		s.Equal("1", req.URL.Query().Get("activity_type"))
		s.Equal("login", req.URL.Query().Get("search"))
		s.Equal("2026-08-01", req.URL.Query().Get("start_date"))
		s.Equal("2026-08-02", req.URL.Query().Get("end_date"))

		_, _ = w.Write([]byte(instantlytest.Page([]string{logFixture}, "cursor-2")))
	})

	page, err := s.svc().List(context.Background(),
		auditlog.WithLimit(50),
		auditlog.WithActivityType(auditlog.ActivityTypeUserLogin),
		auditlog.WithSearch("login"),
		auditlog.WithStartDate(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)),
		auditlog.WithEndDate(time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 1)
	s.Equal("cursor-2", page.NextStartingAfter)

	got := page.Items[0]
	s.Equal("log-1", got.ID)
	s.Equal(auditlog.ActivityTypeUserLogin, got.ActivityType)
	s.True(got.FromAPI)
	s.Require().NotNil(got.AffectedCount)
	s.InEpsilon(3.0, *got.AffectedCount, 1e-9)
	s.Require().NotNil(got.UserName)
	s.Equal("John Doe", *got.UserName)
	s.JSONEq(`{"detail":"value"}`, string(got.AuditMetadata))
}

// TestListNulls verifies a record with every nullable field null decodes to nil
// pointers rather than zero values.
func (s *AuditLogTestSuite) TestListNulls() {
	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(instantlytest.Page([]string{logFixtureNulls}, "")))
	})

	page, err := s.svc().List(context.Background())

	s.Require().NoError(err)
	s.Require().Len(page.Items, 1)

	got := page.Items[0]
	s.Equal("log-2", got.ID)
	s.Equal(auditlog.ActivityTypeAPIKeyDeleted, got.ActivityType)
	s.False(got.FromAPI)
	s.Nil(got.AffectedCount)
	s.Nil(got.APIKeyID)
	s.Nil(got.CampaignID)
	s.Nil(got.ListID)
	s.Nil(got.SubsequenceID)
	s.Nil(got.WebhookID)
	s.Nil(got.UserAgent)
	s.Nil(got.UserID)
	s.Nil(got.UserName)
	s.Nil(got.AuditMetadata)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string, even
// with a nil option.
func (s *AuditLogTestSuite) TestListWithoutOptions() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery, "an unfiltered list must not send an empty query string")
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background(), nil)

	s.Require().NoError(err)
	s.Require().NotNil(page)
	s.Empty(page.Items)
}

// TestListOptions verifies each documented list query parameter is rendered by
// exactly one option, under the key and value the API expects.
func (s *AuditLogTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option auditlog.ListOption
		key    string
		value  string
	}{
		{"limit", auditlog.WithLimit(50), "limit", "50"},
		{"starting after", auditlog.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"activity type", auditlog.WithActivityType(auditlog.ActivityTypeCampaignLaunch), "activity_type", "4"},
		{"search", auditlog.WithSearch("login"), "search", "login"},
		{"start date", auditlog.WithStartDate(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)), "start_date", "2026-08-01"},
		{"end date", auditlog.WithEndDate(time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)), "end_date", "2026-08-02"},
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

// TestFailures verifies the list endpoint surfaces the documented API error.
func (s *AuditLogTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "list", Method: http.MethodGet, Path: listPath, Status: http.StatusTooManyRequests,
			Call: func() error { _, err := svc.List(ctx); return err },
		},
	})
}

// TestParsedTimestamp verifies the RFC 3339 accessor parses a valid timestamp
// and reports an error for an unparseable one.
func (s *AuditLogTestSuite) TestParsedTimestamp() {
	got, err := (&auditlog.Log{Timestamp: "2026-08-01T10:00:00.000Z"}).ParsedTimestamp()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&auditlog.Log{Timestamp: "not-a-timestamp"}).ParsedTimestamp()
	s.Require().Error(err)
}

// TestActivityTypeValues verifies a sample of the named activity types map to
// their documented wire numbers, including across the numbering gap.
func (s *AuditLogTestSuite) TestActivityTypeValues() {
	s.Equal(int64(1), int64(auditlog.ActivityTypeUserLogin))
	s.Equal(int64(12), int64(auditlog.ActivityTypeSubsequenceUpdate))
	s.Equal(int64(18), int64(auditlog.ActivityTypeWebhookCreated))
	s.Equal(int64(35), int64(auditlog.ActivityTypeSuperSearchEnrichmentCreated))
	s.Equal(int64(38), int64(auditlog.ActivityTypeAPIKeyDeleted))
}

// svc builds an Audit Log service pointed at the suite's mock client.
func (s *AuditLogTestSuite) svc() *auditlog.Service {
	return auditlog.New(s.Client)
}
