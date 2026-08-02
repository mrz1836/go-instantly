package campaign_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/campaign"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// Router patterns and identifiers the campaign endpoints are exercised with.
const (
	listPath       = "/api/v2/campaigns"
	idPattern      = "/api/v2/campaigns/:id"
	activatePatt   = "/api/v2/campaigns/:id/activate"
	pausePatt      = "/api/v2/campaigns/:id/pause"
	duplicatePatt  = "/api/v2/campaigns/:id/duplicate"
	sharePatt      = "/api/v2/campaigns/:id/share"
	exportPatt     = "/api/v2/campaigns/:id/export"
	fromExportPatt = "/api/v2/campaigns/:id/from-export"
	variablesPatt  = "/api/v2/campaigns/:id/variables"
	sendingPatt    = "/api/v2/campaigns/:id/sending-status"
	countPath      = "/api/v2/campaigns/count-launched"
	searchPath     = "/api/v2/campaigns/search-by-contact"
	analyticsPath  = "/api/v2/campaigns/analytics"
	overviewPath   = "/api/v2/campaigns/analytics/overview"
	dailyPath      = "/api/v2/campaigns/analytics/daily"
	stepsPath      = "/api/v2/campaigns/analytics/steps"

	testID = "campaign-uuid-1"
)

// campaignFixture is a spec-shaped campaign with the required fields plus a
// representative set of populated optional fields.
const campaignFixture = `{
	"id": "campaign-uuid-1",
	"name": "Launch",
	"status": 1,
	"campaign_schedule": {
		"start_date": "2026-08-01",
		"schedules": [{"name": "Business hours", "timezone": "UTC", "timing": {"from":"09:00","to":"17:00"}, "days": {"1": true}}]
	},
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"timestamp_updated": "2026-08-01T11:00:00.000Z",
	"open_tracking": true,
	"daily_limit": 100,
	"email_list": ["sender@example.com"],
	"stop_on_reply": true,
	"sequences": [{"steps": []}],
	"custom_variables": {"foo": "bar"}
}`

// campaignFixtureNulls has every nullable field explicitly null.
const campaignFixtureNulls = `{
	"id": "campaign-uuid-2",
	"name": "Bare",
	"status": 0,
	"campaign_schedule": {"schedules": []},
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"timestamp_updated": "2026-08-01T11:00:00.000Z",
	"open_tracking": false,
	"daily_limit": null,
	"email_gap": null,
	"is_evergreen": null,
	"stop_on_reply": null,
	"organization": null,
	"owned_by": null,
	"pl_value": null
}`

// CampaignTestSuite exercises the Campaign API service against the mock router.
type CampaignTestSuite struct {
	instantlytest.Suite
}

// TestCampaignSuite runs the Campaign API suite.
func TestCampaignSuite(t *testing.T) {
	suite.Run(t, new(CampaignTestSuite))
}

// TestCreate verifies the create body reaches the API and the campaign decodes.
func (s *CampaignTestSuite) TestCreate() {
	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		var received campaign.CreateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("Launch", received.Name)
		if s.NotNil(received.CampaignSchedule.StartDate) {
			s.Equal("2026-08-01", *received.CampaignSchedule.StartDate)
		}
		if s.NotNil(received.OpenTracking) {
			s.True(*received.OpenTracking)
		}

		_, _ = w.Write([]byte(campaignFixture))
	})

	got, err := s.svc().Create(context.Background(), campaign.CreateRequest{
		Name: "Launch",
		CampaignSchedule: campaign.Schedule{
			StartDate: instantly.Ptr("2026-08-01"),
			Schedules: []campaign.ScheduleItem{{Name: "Business hours", Timezone: "UTC"}},
		},
		OpenTracking: instantly.Ptr(true),
		EmailList:    []string{"sender@example.com"},
	})

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
	s.Equal(campaign.StatusActive, got.Status)
}

// TestList verifies a page decodes, including nullable-vs-zero fields.
func (s *CampaignTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("50", req.URL.Query().Get("limit"))
		s.Equal("1", req.URL.Query().Get("status"))

		_, _ = w.Write([]byte(
			`{"items":[` + campaignFixture + `,` + campaignFixtureNulls + `],"next_starting_after":"cursor-2"}`,
		))
	})

	page, err := s.svc().List(context.Background(),
		campaign.WithLimit(50),
		campaign.WithStatus(campaign.StatusActive),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)

	populated := page.Items[0]
	s.Require().NotNil(populated.DailyLimit)
	s.InDelta(100, *populated.DailyLimit, 0)
	s.Equal([]string{"sender@example.com"}, populated.EmailList)
	s.JSONEq(`{"foo":"bar"}`, string(populated.CustomVariables))

	// Nullable fields stay nil rather than collapsing to a zero value.
	bare := page.Items[1]
	s.Nil(bare.DailyLimit)
	s.Nil(bare.IsEvergreen)
	s.Nil(bare.Organization)
	s.Nil(bare.PLValue)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string.
func (s *CampaignTestSuite) TestListWithoutOptions() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery)
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background(), nil)

	s.Require().NoError(err)
	s.Empty(page.Items)
}

// TestGet verifies a single campaign decodes, including the nested schedule.
func (s *CampaignTestSuite) TestGet() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(campaignFixture))
	})

	got, err := s.svc().Get(context.Background(), testID)

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
	s.Equal("Launch", got.Name)
	s.Require().Len(got.CampaignSchedule.Schedules, 1)
	s.Equal("Business hours", got.CampaignSchedule.Schedules[0].Name)
	s.JSONEq(`{"from":"09:00","to":"17:00"}`, string(got.CampaignSchedule.Schedules[0].Timing))
}

// TestUpdate verifies the patch body is sent and the campaign decodes.
func (s *CampaignTestSuite) TestUpdate() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))

		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("Renamed", received["name"])
		s.Equal(true, received["stop_on_reply"])

		_, _ = w.Write([]byte(campaignFixture))
	})

	got, err := s.svc().Update(context.Background(), testID, campaign.UpdateRequest{
		Name:        "Renamed",
		StopOnReply: instantly.Ptr(true),
	})

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestUpdateOmitsUnsetFields verifies an empty patch sends no field.
func (s *CampaignTestSuite) TestUpdateOmitsUnsetFields() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Empty(received, "an unset patch field must not be sent")

		_, _ = w.Write([]byte(campaignFixture))
	})

	got, err := s.svc().Update(context.Background(), testID, campaign.UpdateRequest{})

	s.Require().NoError(err)
	s.NotNil(got)
}

// TestDelete verifies the deleted campaign is returned to the caller.
func (s *CampaignTestSuite) TestDelete() {
	s.Router.Delete(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(campaignFixture))
	})

	got, err := s.svc().Delete(context.Background(), testID)

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestPathParametersAreEscaped verifies a caller-supplied id cannot rewrite the
// request path.
func (s *CampaignTestSuite) TestPathParametersAreEscaped() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, campaignFixture), nil
			},
		)},
	))

	_, err := campaign.New(client).Get(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/campaigns/..%2Fadmin%3Fx=1", requestURI)
}

// TestListOptions verifies each documented list query parameter is rendered by
// exactly one option, under the key and value the API expects.
func (s *CampaignTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option campaign.ListOption
		key    string
		value  string
	}{
		{"limit", campaign.WithLimit(50), "limit", "50"},
		{"starting after", campaign.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"search", campaign.WithSearch("launch"), "search", "launch"},
		{"tag ids", campaign.WithTagIDs("t1"), "tag_ids", "t1"},
		{"ai sales agent", campaign.WithAISalesAgentID("agent-1"), "ai_sales_agent_id", "agent-1"},
		{"status", campaign.WithStatus(campaign.StatusPaused), "status", "2"},
		{"exclude status", campaign.WithExcludeStatus(campaign.StatusCompleted), "exclude_status", "3"},
	}

	s.Require().Len(tests, 7, "every documented list query parameter needs an option")

	for _, test := range tests {
		s.Run(test.name, func() {
			q := instantly.NewQuery()
			test.option(q)

			s.Require().Equal(1, q.Len(), "an option must render exactly one query parameter")
			s.Equal(test.value, q.Get(test.key))
		})
	}

	// WithTagIDs renders each id as a repeated parameter.
	q := instantly.NewQuery()
	campaign.WithTagIDs("a", "b")(q)
	s.Equal("tag_ids=a&tag_ids=b", q.Encode())
}

// TestFailures verifies the CRUD and analytics endpoints surface the documented
// API error.
func (s *CampaignTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "create", Method: http.MethodPost, Path: listPath, Status: http.StatusPaymentRequired,
			Call: func() error { _, err := svc.Create(ctx, campaign.CreateRequest{}); return err },
		},
		{
			Name: "list", Method: http.MethodGet, Path: listPath, Status: http.StatusTooManyRequests,
			Call: func() error { _, err := svc.List(ctx); return err },
		},
		{
			Name: "get", Method: http.MethodGet, Path: idPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Get(ctx, "missing"); return err },
		},
		{
			Name: "update", Method: http.MethodPatch, Path: idPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Update(ctx, "missing", campaign.UpdateRequest{}); return err },
		},
		{
			Name: "delete", Method: http.MethodDelete, Path: idPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Delete(ctx, "missing"); return err },
		},
		{
			Name: "analytics", Method: http.MethodGet, Path: analyticsPath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.Analytics(ctx); return err },
		},
		{
			Name: "analyticsOverview", Method: http.MethodGet, Path: overviewPath, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.AnalyticsOverview(ctx); return err },
		},
		{
			Name: "dailyAnalytics", Method: http.MethodGet, Path: dailyPath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.DailyAnalytics(ctx); return err },
		},
		{
			Name: "stepsAnalytics", Method: http.MethodGet, Path: stepsPath, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.StepsAnalytics(ctx); return err },
		},
	})
}

// TestActionFailures verifies the lifecycle and action endpoints surface the
// documented API error.
func (s *CampaignTestSuite) TestActionFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "activate", Method: http.MethodPost, Path: activatePatt, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Activate(ctx, "missing"); return err },
		},
		{
			Name: "duplicate", Method: http.MethodPost, Path: duplicatePatt, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Duplicate(ctx, "missing", campaign.DuplicateRequest{}); return err },
		},
		{
			Name: "share", Method: http.MethodPost, Path: sharePatt, Status: http.StatusForbidden,
			Call: func() error { return svc.Share(ctx, "missing") },
		},
		{
			Name: "export", Method: http.MethodPost, Path: exportPatt, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Export(ctx, "missing"); return err },
		},
		{
			Name: "createFromExport", Method: http.MethodPost, Path: fromExportPatt, Status: http.StatusPaymentRequired,
			Call: func() error { _, err := svc.CreateFromExport(ctx, "missing", nil); return err },
		},
		{
			Name: "addVariables", Method: http.MethodPost, Path: variablesPatt, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.AddVariables(ctx, "missing", campaign.AddVariablesRequest{}); return err },
		},
		{
			Name: "sendingStatus", Method: http.MethodGet, Path: sendingPatt, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.SendingStatus(ctx, "missing"); return err },
		},
		{
			Name: "countLaunched", Method: http.MethodGet, Path: countPath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.CountLaunched(ctx); return err },
		},
		{
			Name: "searchByContact", Method: http.MethodGet, Path: searchPath, Status: http.StatusTooManyRequests,
			Call: func() error { _, err := svc.SearchByContact(ctx, "missing"); return err },
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor parses a valid
// timestamp and reports an error for an unparseable one.
func (s *CampaignTestSuite) TestParsedTimestampCreated() {
	got, err := (&campaign.Campaign{TimestampCreated: "2026-08-01T10:00:00.000Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&campaign.Campaign{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// svc builds a Campaign service pointed at the suite's mock client.
func (s *CampaignTestSuite) svc() *campaign.Service {
	return campaign.New(s.Client)
}
