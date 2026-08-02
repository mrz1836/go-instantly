package campaign_test

import (
	"context"
	"net/http"
	"time"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/campaign"
)

// TestAnalytics verifies the per-campaign analytics slice decodes and options
// reach the API.
func (s *CampaignTestSuite) TestAnalytics() {
	s.Router.Get(analyticsPath, func(w http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()
		s.Equal(testID, query.Get("id"))
		s.Equal([]string{"c1", "c2"}, query["ids"])
		s.Equal("true", query.Get("exclude_total_leads_count"))

		_, _ = w.Write([]byte(
			`[{"campaign_id":"campaign-uuid-1","campaign_name":"Launch","campaign_status":1,` +
				`"campaign_is_evergreen":false,"leads_count":100,"contacted_count":80,` +
				`"emails_sent_count":200,"open_count":150,"reply_count":10,` +
				`"total_opportunity_value":1234.5}]`,
		))
	})

	got, err := s.svc().Analytics(context.Background(),
		campaign.WithID(testID),
		campaign.WithIDs("c1", "c2"),
		campaign.WithExcludeTotalLeadsCount(true),
	)

	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal(testID, got[0].CampaignID)
	s.Equal(campaign.StatusActive, got[0].CampaignStatus)
	s.Equal(int64(100), got[0].LeadsCount)
	s.InDelta(1234.5, got[0].TotalOpportunityValue, 0)
}

// TestAnalyticsOverview verifies the overview object decodes, including CRM
// totals.
func (s *CampaignTestSuite) TestAnalyticsOverview() {
	s.Router.Get(overviewPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("true", req.URL.Query().Get("expand_crm_events"))
		_, _ = w.Write([]byte(
			`{"contacted_count":80,"emails_sent_count":200,"open_count":150,"reply_count":10,` +
				`"total_opportunities":5,"total_opportunity_value":9999.99,"total_interested":4,` +
				`"total_meeting_booked":3,"total_meeting_completed":2,"total_closed":1}`,
		))
	})

	got, err := s.svc().AnalyticsOverview(context.Background(),
		campaign.WithExpandCRMEvents(true),
	)

	s.Require().NoError(err)
	s.Equal(int64(80), got.ContactedCount)
	s.Equal(int64(1), got.TotalClosed)
	s.InDelta(9999.99, got.TotalOpportunityValue, 0)
}

// TestDailyAnalytics verifies the daily slice decodes and options reach the API.
func (s *CampaignTestSuite) TestDailyAnalytics() {
	s.Router.Get(dailyPath, func(w http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()
		s.Equal(testID, query.Get("campaign_id"))
		s.Equal("2026-08-01", query.Get("start_date"))
		s.Equal("2026-08-31", query.Get("end_date"))
		s.Equal("1", query.Get("campaign_status"))

		_, _ = w.Write([]byte(
			`[{"date":"2026-08-01","sent":10,"opened":5,"unique_opened":4,"replies":2,` +
				`"unique_replies":2,"replies_automatic":0,"unique_replies_automatic":0,"clicks":1,` +
				`"unique_clicks":1,"contacted":8,"new_leads_contacted":8,"opportunities":1,` +
				`"unique_opportunities":1}]`,
		))
	})

	got, err := s.svc().DailyAnalytics(context.Background(),
		campaign.WithCampaignID(testID),
		campaign.WithStartDate(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)),
		campaign.WithEndDate(time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)),
		campaign.WithCampaignStatus(campaign.StatusActive),
	)

	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal("2026-08-01", got[0].Date)
	s.Equal(int64(10), got[0].Sent)
}

// TestStepsAnalytics verifies the per-step slice decodes, including nullable step
// and variant.
func (s *CampaignTestSuite) TestStepsAnalytics() {
	s.Router.Get(stepsPath, func(w http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()
		s.Equal(testID, query.Get("campaign_id"))
		s.Equal("true", query.Get("include_opportunities_count"))

		_, _ = w.Write([]byte(
			`[{"step":"1","variant":"A","sent":10,"opened":5,"unique_opened":4,"replies":2,` +
				`"unique_replies":2,"replies_automatic":0,"clicks":1,"unique_clicks":1,` +
				`"opportunities":1,"unique_opportunities":1,"meetings_booked":1,"won":0},` +
				`{"step":null,"variant":null,"sent":0,"opened":0,"unique_opened":0,"replies":0,` +
				`"unique_replies":0,"replies_automatic":0,"clicks":0,"unique_clicks":0,` +
				`"opportunities":0,"unique_opportunities":0,"meetings_booked":0,"won":0}]`,
		))
	})

	got, err := s.svc().StepsAnalytics(context.Background(),
		campaign.WithCampaignID(testID),
		campaign.WithIncludeOpportunitiesCount(true),
	)

	s.Require().NoError(err)
	s.Require().Len(got, 2)
	s.Require().NotNil(got[0].Step)
	s.Equal("1", *got[0].Step)
	s.Equal(int64(1), got[0].MeetingsBooked)
	s.Nil(got[1].Step, "a null step must stay nil")
	s.Nil(got[1].Variant)
}

// TestAnalyticsOptions verifies each analytics option renders correctly.
func (s *CampaignTestSuite) TestAnalyticsOptions() {
	tests := []struct {
		name   string
		option campaign.AnalyticsOption
		key    string
		value  string
	}{
		{"id", campaign.WithID("c1"), "id", "c1"},
		{"campaign id", campaign.WithCampaignID("c1"), "campaign_id", "c1"},
		{
			"start date",
			campaign.WithStartDate(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)),
			"start_date", "2026-08-01",
		},
		{
			"end date",
			campaign.WithEndDate(time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)),
			"end_date", "2026-08-31",
		},
		{"campaign status", campaign.WithCampaignStatus(campaign.StatusPaused), "campaign_status", "2"},
		{"exclude total leads", campaign.WithExcludeTotalLeadsCount(true), "exclude_total_leads_count", "true"},
		{"expand crm events", campaign.WithExpandCRMEvents(true), "expand_crm_events", "true"},
		{"include opportunities", campaign.WithIncludeOpportunitiesCount(true), "include_opportunities_count", "true"},
	}

	s.Require().Len(tests, 8)

	for _, test := range tests {
		s.Run(test.name, func() {
			q := instantly.NewQuery()
			test.option(q)

			s.Require().Equal(1, q.Len())
			s.Equal(test.value, q.Get(test.key))
		})
	}

	// WithIDs renders a repeated parameter.
	q := instantly.NewQuery()
	campaign.WithIDs("a", "b")(q)
	s.Equal("ids=a&ids=b", q.Encode())
}
