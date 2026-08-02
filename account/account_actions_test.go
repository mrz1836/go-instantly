package account_test

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/account"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// TestPauseResumeMarkFixed verifies the single-account actions POST to their
// sub-path with no body and return the updated account.
func (s *AccountTestSuite) TestPauseResumeMarkFixed() {
	tests := []struct {
		name    string
		pattern string
		call    func(svc *account.Service) (*account.Account, error)
	}{
		{"pause", pausePattern, func(svc *account.Service) (*account.Account, error) {
			return svc.Pause(context.Background(), testEmail)
		}},
		{"resume", resumePattern, func(svc *account.Service) (*account.Account, error) {
			return svc.Resume(context.Background(), testEmail)
		}},
		{"mark fixed", markFixedPatt, func(svc *account.Service) (*account.Account, error) {
			return svc.MarkFixed(context.Background(), testEmail)
		}},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.Router.Post(test.pattern, func(w http.ResponseWriter, req *http.Request) {
				s.Equal(testEmail, instantlytest.PathParam(req, "email"))

				body, err := instantlytest.ReadAll(req)
				s.NoError(err)
				s.Empty(body, "a single-account action sends no request body")

				_, _ = w.Write([]byte(accountFixture))
			})

			got, err := test.call(s.svc())

			s.Require().NoError(err)
			s.Equal(testEmail, got.Email)
		})
	}
}

// TestPauseBulk verifies the bulk-pause body is sent and the split result decodes.
func (s *AccountTestSuite) TestPauseBulk() {
	s.Router.Post(bulkPausePath, func(w http.ResponseWriter, req *http.Request) {
		var received account.PauseBulkRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal([]string{"a@x.com", "b@x.com"}, received.Emails)

		_, _ = w.Write([]byte(`{"paused_emails":["a@x.com"],"failed_emails":["b@x.com"]}`))
	})

	got, err := s.svc().PauseBulk(context.Background(), account.PauseBulkRequest{
		Emails: []string{"a@x.com", "b@x.com"},
	})

	s.Require().NoError(err)
	s.Equal([]string{"a@x.com"}, got.PausedEmails)
	s.Equal([]string{"b@x.com"}, got.FailedEmails)
}

// TestMove verifies the move body is sent and the status decodes.
func (s *AccountTestSuite) TestMove() {
	s.Router.Post(movePath, func(w http.ResponseWriter, req *http.Request) {
		var received account.MoveRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("ws-1", received.SourceWorkspaceID)
		s.Equal("ws-2", received.DestinationWorkspaceID)

		_, _ = w.Write([]byte(`{"status":"success"}`))
	})

	got, err := s.svc().Move(context.Background(), account.MoveRequest{
		Emails:                 []string{testEmail},
		SourceWorkspaceID:      "ws-1",
		DestinationWorkspaceID: "ws-2",
	})

	s.Require().NoError(err)
	s.Equal("success", got.Status)
}

// TestEnableDisableWarmup verifies the warmup toggles POST to their sub-path and
// decode the background job.
func (s *AccountTestSuite) TestEnableDisableWarmup() {
	tests := []struct {
		name string
		path string
		call func(svc *account.Service) (*account.BackgroundJob, error)
	}{
		{"enable", enableWarmup, func(svc *account.Service) (*account.BackgroundJob, error) {
			return svc.EnableWarmup(context.Background(), account.WarmupToggleRequest{
				IncludeAllEmails: instantly.Ptr(true),
			})
		}},
		{"disable", disableWarmup, func(svc *account.Service) (*account.BackgroundJob, error) {
			return svc.DisableWarmup(context.Background(), account.WarmupToggleRequest{
				Emails: []string{testEmail},
			})
		}},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.Router.Post(test.path, func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(
					`{"id":"job-1","type":"warmup","status":"pending","progress":0,` +
						`"entity_type":"account","workspace_id":"ws-1",` +
						`"created_at":"2026-08-01T10:00:00.000Z","updated_at":"2026-08-01T10:00:00.000Z",` +
						`"entity_id":null,"user_id":null}`,
				))
			})

			got, err := test.call(s.svc())

			s.Require().NoError(err)
			s.Equal("job-1", got.ID)
			s.Equal("pending", got.Status)
			s.Nil(got.EntityID)
		})
	}
}

// TestWarmupAnalytics verifies the raw analytics payloads are preserved.
func (s *AccountTestSuite) TestWarmupAnalytics() {
	s.Router.Post(warmupAnalytics, func(w http.ResponseWriter, req *http.Request) {
		var received account.WarmupAnalyticsRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal([]string{testEmail}, received.Emails)

		_, _ = w.Write([]byte(`{"aggregate_data":{"sent":10},"email_date_data":{"x":1}}`))
	})

	got, err := s.svc().WarmupAnalytics(context.Background(), account.WarmupAnalyticsRequest{
		Emails: []string{testEmail},
	})

	s.Require().NoError(err)
	s.JSONEq(`{"sent":10}`, string(got.AggregateData))
	s.JSONEq(`{"x":1}`, string(got.EmailDateData))
}

// TestDailyAnalytics verifies the option query is sent and the slice decodes.
func (s *AccountTestSuite) TestDailyAnalytics() {
	s.Router.Get(dailyAnalytics, func(w http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()
		s.Equal("2026-08-01", query.Get("start_date"))
		s.Equal("2026-08-31", query.Get("end_date"))
		s.Equal([]string{"a@x.com", "b@x.com"}, query["emails"])

		_, _ = w.Write([]byte(
			`[{"date":"2026-08-01","email_account":"a@x.com","sent":10,"opened":5,` +
				`"unique_opened":4,"replies":2,"unique_replies":2,"replies_automatic":0,` +
				`"unique_replies_automatic":0,"clicks":1,"unique_clicks":1,"contacted":8,` +
				`"new_leads_contacted":8,"bounced":0}]`,
		))
	})

	got, err := s.svc().DailyAnalytics(context.Background(),
		account.WithStartDate(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)),
		account.WithEndDate(time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)),
		account.WithEmails("a@x.com", "b@x.com"),
	)

	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal("a@x.com", got[0].EmailAccount)
	s.Equal(int64(10), got[0].Sent)
}

// TestAnalyticsOptions verifies each daily-analytics option renders correctly.
func (s *AccountTestSuite) TestAnalyticsOptions() {
	q := instantly.NewQuery()
	account.WithStartDate(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))(q)
	account.WithEndDate(time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC))(q)
	account.WithEmails("a@x.com", "b@x.com")(q)

	s.Equal("2026-08-01", q.Get("start_date"))
	s.Equal("2026-08-31", q.Get("end_date"))
	s.Equal("emails=a%40x.com&emails=b%40x.com&end_date=2026-08-31&start_date=2026-08-01", q.Encode())
}

// TestCtdStatus verifies the host is sent and the status decodes.
func (s *AccountTestSuite) TestCtdStatus() {
	s.Router.Get(ctdStatusPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("track.example.com", req.URL.Query().Get("host"))
		_, _ = w.Write([]byte(`{"host":"track.example.com","success":true,"cname":true,"ssl":false}`))
	})

	got, err := s.svc().CtdStatus(context.Background(), "track.example.com")

	s.Require().NoError(err)
	s.True(got.Success)
	s.True(got.CNAME)
	s.False(got.SSL)
}

// TestVitals verifies the accounts are sent and the split result decodes.
func (s *AccountTestSuite) TestVitals() {
	s.Router.Post(vitalsPath, func(w http.ResponseWriter, req *http.Request) {
		var received account.VitalsRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal([]string{testEmail}, received.Accounts)

		_, _ = w.Write([]byte(
			`{"status":"done","success_list":[{"domain":"x.com","allPass":true,"dkim":true,` +
				`"dmarc":true,"mx":true,"spf":true}],"failure_list":[]}`,
		))
	})

	got, err := s.svc().TestVitals(context.Background(), account.VitalsRequest{
		Accounts: []string{testEmail},
	})

	s.Require().NoError(err)
	s.Equal("done", got.Status)
	s.Require().Len(got.SuccessList, 1)
	s.True(got.SuccessList[0].AllPass)
	s.Empty(got.FailureList)
}
