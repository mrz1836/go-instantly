package campaign_test

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/campaign"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// TestActivatePause verifies the lifecycle actions POST to their sub-path with
// no body and return the campaign.
func (s *CampaignTestSuite) TestActivatePause() {
	tests := []struct {
		name    string
		pattern string
		call    func(svc *campaign.Service) (*campaign.Campaign, error)
	}{
		{"activate", activatePatt, func(svc *campaign.Service) (*campaign.Campaign, error) {
			return svc.Activate(context.Background(), testID)
		}},
		{"pause", pausePatt, func(svc *campaign.Service) (*campaign.Campaign, error) {
			return svc.Pause(context.Background(), testID)
		}},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.Router.Post(test.pattern, func(w http.ResponseWriter, req *http.Request) {
				s.Equal(testID, instantlytest.PathParam(req, "id"))

				body, err := instantlytest.ReadAll(req)
				s.NoError(err)
				s.Empty(body, "a lifecycle action sends no request body")

				_, _ = w.Write([]byte(campaignFixture))
			})

			got, err := test.call(s.svc())

			s.Require().NoError(err)
			s.Equal(testID, got.ID)
		})
	}
}

// TestDuplicate verifies the duplicate name is sent and the new campaign decodes.
func (s *CampaignTestSuite) TestDuplicate() {
	s.Router.Post(duplicatePatt, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))

		var received campaign.DuplicateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("Launch Copy", received.Name)

		_, _ = w.Write([]byte(campaignFixture))
	})

	got, err := s.svc().Duplicate(context.Background(), testID, campaign.DuplicateRequest{Name: "Launch Copy"})

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestShare verifies the share endpoint is called without a body and reports
// success as a nil error.
func (s *CampaignTestSuite) TestShare() {
	s.Router.Post(sharePatt, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))

		body, err := instantlytest.ReadAll(req)
		s.NoError(err)
		s.Empty(body, "sharing sends no request body")

		w.WriteHeader(http.StatusOK)
	})

	s.Require().NoError(s.svc().Share(context.Background(), testID))
}

// TestExport verifies the export returns the campaign's JSON representation.
func (s *CampaignTestSuite) TestExport() {
	s.Router.Post(exportPatt, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(campaignFixture))
	})

	got, err := s.svc().Export(context.Background(), testID)

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestCreateFromExport verifies the exported payload is sent verbatim and the new
// campaign decodes.
func (s *CampaignTestSuite) TestCreateFromExport() {
	s.Router.Post(fromExportPatt, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))

		body, err := instantlytest.ReadAll(req)
		s.NoError(err)
		s.JSONEq(`{"name":"Imported"}`, string(body))

		_, _ = w.Write([]byte(campaignFixture))
	})

	got, err := s.svc().CreateFromExport(context.Background(), testID, json.RawMessage(`{"name":"Imported"}`))

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestAddVariables verifies the variables array is sent and the campaign decodes.
func (s *CampaignTestSuite) TestAddVariables() {
	s.Router.Post(variablesPatt, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))

		var received campaign.AddVariablesRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.JSONEq(`[{"name":"industry"}]`, string(received.Variables))

		_, _ = w.Write([]byte(campaignFixture))
	})

	got, err := s.svc().AddVariables(context.Background(), testID, campaign.AddVariablesRequest{
		Variables: json.RawMessage(`[{"name":"industry"}]`),
	})

	s.Require().NoError(err)
	s.Equal(testID, got.ID)
}

// TestSendingStatus verifies the raw summary and diagnostics are preserved.
func (s *CampaignTestSuite) TestSendingStatus() {
	s.Router.Get(sendingPatt, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(`{"summary":{"sending":true},"diagnostics":[{"code":"OK"}]}`))
	})

	got, err := s.svc().SendingStatus(context.Background(), testID)

	s.Require().NoError(err)
	s.JSONEq(`{"sending":true}`, string(got.Summary))
	s.JSONEq(`[{"code":"OK"}]`, string(got.Diagnostics))
}

// TestCountLaunched verifies the launched count is unwrapped for the caller.
func (s *CampaignTestSuite) TestCountLaunched() {
	s.Router.Get(countPath, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"count":7}`))
	})

	count, err := s.svc().CountLaunched(context.Background())

	s.Require().NoError(err)
	s.Equal(int64(7), count)
}

// TestSearchByContact verifies the contact and sort options are sent and the
// items are unwrapped.
func (s *CampaignTestSuite) TestSearchByContact() {
	s.Router.Get(searchPath, func(w http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()
		s.Equal("lead@example.com", query.Get("search"))
		s.Equal("name", query.Get("sort_column"))
		s.Equal("asc", query.Get("sort_order"))

		_, _ = w.Write([]byte(`{"items":[` + campaignFixture + `]}`))
	})

	got, err := s.svc().SearchByContact(context.Background(), "lead@example.com",
		campaign.WithSortColumn("name"),
		campaign.WithSortOrder(instantly.SortOrderAsc),
	)

	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal(testID, got[0].ID)
}
