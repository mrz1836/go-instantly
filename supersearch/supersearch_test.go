package supersearch_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/supersearch"
)

// Router patterns and identifiers the supersearch-enrichment endpoints are
// exercised with. The patterns carry the full request path, including the
// /api/v2 prefix.
const (
	// createPath is the create-enrichment endpoint. The API path carries a
	// trailing slash.
	createPath = "/api/v2/supersearch-enrichment/"

	// aiPath is the create-AI-enrichment endpoint.
	aiPath = "/api/v2/supersearch-enrichment/ai"

	// aiInProgressPattern is the router pattern for the in-progress AI endpoint.
	aiInProgressPattern = "/api/v2/supersearch-enrichment/ai/:resource_id/in-progress"

	// resourcePattern is the router pattern for the get-enrichment endpoint.
	resourcePattern = "/api/v2/supersearch-enrichment/:resource_id"

	// settingsPattern is the router pattern for the update-settings endpoint.
	settingsPattern = "/api/v2/supersearch-enrichment/:resource_id/settings"

	// enrichPath is the enrich-leads endpoint.
	enrichPath = "/api/v2/supersearch-enrichment/enrich-leads-from-supersearch"

	// runPath is the run-enrichment endpoint.
	runPath = "/api/v2/supersearch-enrichment/run"

	// countPath is the count-leads endpoint.
	countPath = "/api/v2/supersearch-enrichment/count-leads-from-supersearch"

	// previewPath is the preview-leads endpoint.
	previewPath = "/api/v2/supersearch-enrichment/preview-leads-from-supersearch"

	// facetPath is the signal-keywords-facet endpoint.
	facetPath = "/api/v2/supersearch-enrichment/signal-keywords-facet"

	// historyPattern is the router pattern for the enrichment-history endpoint.
	historyPattern = "/api/v2/supersearch-enrichment/history/:resource_id"

	// resourceID identifies the resource the endpoints operate on.
	resourceID = "res-1"
)

// Fixtures for the distinct response shapes the endpoints return. Nullable
// fields are exercised both populated and explicitly null so an absent value
// stays distinguishable from a zero value.
const (
	// createFixture is a create response, which omits the settings-only fields.
	createFixture = `{
		"id": "enr-1",
		"organization_id": "org-1",
		"resource_id": "res-1",
		"limit": 100,
		"enrichment_payload": {"columns": ["email"]}
	}`

	// enrichmentFixture is the full enrichment the settings endpoint returns.
	enrichmentFixture = `{
		"id": "enr-1",
		"organization_id": "org-1",
		"resource_id": "res-1",
		"resource_type": 2,
		"type": "work_email_enrichment",
		"limit": 100,
		"auto_update": true,
		"in_progress": false,
		"skip_rows_without_email": true,
		"enrichment_payload": {"columns": ["email"]}
	}`

	// runFixture is the run response, a subset of the enrichment shape.
	runFixture = `{"id": "enr-1", "resource_id": "res-1", "enrichment_payload": {"done": true}}`

	// resourceFixture is the get-enrichment status for a resource.
	resourceFixture = `{
		"resource_id": "res-1",
		"enrichment_payload": {"columns": ["email"]},
		"exists": true,
		"has_no_leads": false,
		"in_progress": true,
		"is_evergreen": false,
		"auto_update": true,
		"search_filters": {"query": "cto"}
	}`

	// resourceFixtureNulls has every nullable status field explicitly null.
	resourceFixtureNulls = `{
		"resource_id": "res-2",
		"enrichment_payload": {},
		"exists": null,
		"has_no_leads": null,
		"in_progress": null,
		"is_evergreen": null,
		"auto_update": null,
		"search_filters": null
	}`

	// aiFixture is the create-AI response, with template_id explicitly null.
	aiFixture = `{
		"id": "ai-1",
		"resource_id": "res-1",
		"resource_type": 1,
		"output_column": "ai_summary",
		"status": 1,
		"model_version": "gpt-4o",
		"overwrite": true,
		"auto_update": false,
		"input_columns": ["company"],
		"limit": 50,
		"template_id": null
	}`

	// aiInProgressFixture is one in-progress AI job, a subset of the AI shape.
	aiInProgressFixture = `{
		"id": "ai-2",
		"organization_id": "org-1",
		"output_column": "ai_summary",
		"resource_id": "res-1",
		"resource_type": 2,
		"status": 2
	}`

	// enrichLeadsFixture is the enrich-leads response, with a set background job
	// and a null live-list workflow.
	enrichLeadsFixture = `{
		"id": "enr-2",
		"organization_id": "org-1",
		"resource_id": "res-3",
		"resource_type": 2,
		"limit": 500,
		"list_name": "My List",
		"custom_flow": ["instantly"],
		"background_job_id": "job-1",
		"live_list_workflow_id": null,
		"search_filters": {"query": "cto"}
	}`

	// leadFixture is one previewed lead.
	leadFixture = `{
		"firstName": "Jane",
		"lastName": "Doe",
		"fullName": "Jane Doe",
		"jobTitle": "CTO",
		"companyId": "c-1",
		"companyName": "Acme",
		"companyLogo": "https://logo",
		"linkedIn": "https://li",
		"location": "NYC",
		"isOwned": false
	}`
)

// SuperSearchTestSuite exercises the SuperSearch Enrichment API service against
// the mock router.
type SuperSearchTestSuite struct {
	instantlytest.Suite
}

// TestSuperSearchSuite runs the SuperSearch Enrichment API suite.
func TestSuperSearchSuite(t *testing.T) {
	suite.Run(t, new(SuperSearchTestSuite))
}

// TestCreate verifies the create body reaches the API and the create response
// decodes, with the settings-only fields staying nil.
func (s *SuperSearchTestSuite) TestCreate() {
	s.Router.Post(createPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodPost, req.Method)
		s.Equal(createPath, req.URL.Path)

		var received supersearch.CreateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(resourceID, received.ResourceID)
		s.Equal(supersearch.EnrichmentWorkEmail, received.Type)
		s.Equal([]string{"instantly"}, received.CustomFlow)
		s.JSONEq(`[{"field":"title"}]`, string(received.Filters))

		_, _ = w.Write([]byte(createFixture))
	})

	got, err := s.svc().Create(context.Background(), supersearch.CreateRequest{
		ResourceID: resourceID,
		Type:       supersearch.EnrichmentWorkEmail,
		Limit:      instantly.Ptr(100.0),
		CustomFlow: []string{"instantly"},
		Filters:    json.RawMessage(`[{"field":"title"}]`),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("enr-1", got.ID)
	s.Require().NotNil(got.Limit)
	s.InDelta(100, *got.Limit, 0)

	// The settings-only fields are absent from a create response and must stay
	// nil rather than collapsing to a zero value.
	s.Nil(got.ResourceType)
	s.Nil(got.AutoUpdate)
}

// TestGet verifies the resource enrichment status decodes, including the raw
// search filters.
func (s *SuperSearchTestSuite) TestGet() {
	s.Router.Get(resourcePattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(resourceID, instantlytest.PathParam(req, "resource_id"))
		_, _ = w.Write([]byte(resourceFixture))
	})

	got, err := s.svc().Get(context.Background(), resourceID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(resourceID, got.ResourceID)
	s.Require().NotNil(got.Exists)
	s.True(*got.Exists)
	s.Require().NotNil(got.InProgress)
	s.True(*got.InProgress)
	s.JSONEq(`{"query":"cto"}`, string(got.SearchFilters))
}

// TestGetNullable verifies the nullable status fields stay nil rather than
// collapsing to a zero value.
func (s *SuperSearchTestSuite) TestGetNullable() {
	s.Router.Get(resourcePattern, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(resourceFixtureNulls))
	})

	got, err := s.svc().Get(context.Background(), "res-2")

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Nil(got.Exists)
	s.Nil(got.HasNoLeads)
	s.Nil(got.InProgress)
	s.Nil(got.IsEvergreen)
	s.Nil(got.AutoUpdate)
	// A JSON null is preserved verbatim by the raw message rather than dropped.
	s.JSONEq(`null`, string(got.SearchFilters))
}

// TestUpdateSettings verifies the patch body is sent and the full enrichment
// decodes, including the enum and nullable fields.
func (s *SuperSearchTestSuite) TestUpdateSettings() {
	s.Router.Patch(settingsPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(resourceID, instantlytest.PathParam(req, "resource_id"))

		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(true, received["auto_update"])
		s.NotContains(received, "is_evergreen", "an unset setting must not be sent")

		_, _ = w.Write([]byte(enrichmentFixture))
	})

	got, err := s.svc().UpdateSettings(context.Background(), resourceID, supersearch.SettingsRequest{
		AutoUpdate: instantly.Ptr(true),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Require().NotNil(got.ResourceType)
	s.Equal(supersearch.ResourceTypeList, *got.ResourceType)
	s.Equal(supersearch.EnrichmentWorkEmail, got.Type)
	s.Require().NotNil(got.AutoUpdate)
	s.True(*got.AutoUpdate)
}

// TestRun verifies the run body is sent and the run response decodes.
func (s *SuperSearchTestSuite) TestRun() {
	s.Router.Post(runPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(runPath, req.URL.Path)

		var received supersearch.RunRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(resourceID, received.ResourceID)
		if s.NotNil(received.Limit) {
			s.Equal(int64(25), *received.Limit)
		}

		_, _ = w.Write([]byte(runFixture))
	})

	got, err := s.svc().Run(context.Background(), supersearch.RunRequest{
		ResourceID: resourceID,
		Limit:      instantly.Ptr(int64(25)),
		Overwrite:  instantly.Ptr(true),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("enr-1", got.ID)
	s.JSONEq(`{"done":true}`, string(got.EnrichmentPayload))
}

// TestCreateAI verifies the AI body is sent and the AI job decodes, with the
// enum values and the null template.
func (s *SuperSearchTestSuite) TestCreateAI() {
	s.Router.Post(aiPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(aiPath, req.URL.Path)

		var received supersearch.AIRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal(resourceID, received.ResourceID)
		s.Equal(supersearch.ResourceTypeCampaign, received.ResourceType)
		s.Equal(supersearch.ModelVersionGPT4o, received.ModelVersion)
		s.Equal("ai_summary", received.OutputColumn)

		_, _ = w.Write([]byte(aiFixture))
	})

	got, err := s.svc().CreateAI(context.Background(), supersearch.AIRequest{
		ResourceID:   resourceID,
		OutputColumn: "ai_summary",
		ResourceType: supersearch.ResourceTypeCampaign,
		ModelVersion: supersearch.ModelVersionGPT4o,
		Prompt:       "Summarize the company",
		Status:       instantly.Ptr(supersearch.AIStatusPending),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("ai-1", got.ID)
	s.Equal(supersearch.ResourceTypeCampaign, got.ResourceType)
	s.Equal(supersearch.AIStatusPending, got.Status)
	s.Equal(supersearch.ModelVersionGPT4o, got.ModelVersion)
	s.Require().NotNil(got.Overwrite)
	s.True(*got.Overwrite)
	s.Nil(got.TemplateID, "a null template id must stay nil")
}

// TestAIInProgress verifies the in-progress AI jobs decode as a bare array, and
// that the exact route is not shadowed by the :resource_id route.
func (s *SuperSearchTestSuite) TestAIInProgress() {
	s.Router.Get(aiInProgressPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(resourceID, instantlytest.PathParam(req, "resource_id"))
		_, _ = w.Write([]byte(`[` + aiInProgressFixture + `]`))
	})

	got, err := s.svc().AIInProgress(context.Background(), resourceID)

	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal("ai-2", got[0].ID)
	s.Equal("org-1", got[0].OrganizationID)
	s.Equal(supersearch.AIStatusProcessing, got[0].Status)
	s.Equal(supersearch.ResourceTypeList, got[0].ResourceType)
}

// TestCountLeads verifies the search body is sent and the count decodes.
func (s *SuperSearchTestSuite) TestCountLeads() {
	s.Router.Post(countPath, func(w http.ResponseWriter, req *http.Request) {
		var received supersearch.SearchRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.JSONEq(`{"query":"cto"}`, string(received.SearchFilters))
		if s.NotNil(received.SkipOwnedLeads) {
			s.True(*received.SkipOwnedLeads)
		}

		_, _ = w.Write([]byte(`{"number_of_leads":1234}`))
	})

	got, err := s.svc().CountLeads(context.Background(), supersearch.SearchRequest{
		SearchFilters:  json.RawMessage(`{"query":"cto"}`),
		SkipOwnedLeads: instantly.Ptr(true),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.InDelta(1234, got.NumberOfLeads, 0)
}

// TestPreviewLeads verifies the search body is sent and the preview decodes,
// including the typed lead sample.
func (s *SuperSearchTestSuite) TestPreviewLeads() {
	s.Router.Post(previewPath, func(w http.ResponseWriter, req *http.Request) {
		var received supersearch.SearchRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.JSONEq(`{"query":"cto"}`, string(received.SearchFilters))

		_, _ = w.Write([]byte(
			`{"leads":[` + leadFixture + `],"number_of_leads":1234,"number_of_redacted_results":10}`,
		))
	})

	got, err := s.svc().PreviewLeads(context.Background(), supersearch.SearchRequest{
		SearchFilters:         json.RawMessage(`{"query":"cto"}`),
		ShowOneLeadPerCompany: instantly.Ptr(true),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.InDelta(1234, got.NumberOfLeads, 0)
	s.InDelta(10, got.NumberOfRedactedResults, 0)
	s.Require().Len(got.Leads, 1)
	s.Equal("Jane Doe", got.Leads[0].FullName)
	s.Equal("Acme", got.Leads[0].CompanyName)
	s.False(got.Leads[0].IsOwned)
}

// TestEnrichLeads verifies the enrich body is sent and the response decodes,
// including nullable-vs-zero fields.
func (s *SuperSearchTestSuite) TestEnrichLeads() {
	s.Router.Post(enrichPath, func(w http.ResponseWriter, req *http.Request) {
		var received supersearch.EnrichLeadsRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.JSONEq(`{"query":"cto"}`, string(received.SearchFilters))
		s.InDelta(500, received.Limit, 0)
		s.Equal("My List", received.ListName)

		_, _ = w.Write([]byte(enrichLeadsFixture))
	})

	got, err := s.svc().EnrichLeads(context.Background(), supersearch.EnrichLeadsRequest{
		SearchFilters:       json.RawMessage(`{"query":"cto"}`),
		Limit:               500,
		ListName:            "My List",
		WorkEmailEnrichment: instantly.Ptr(true),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("enr-2", got.ID)
	s.Require().NotNil(got.BackgroundJobID)
	s.Equal("job-1", *got.BackgroundJobID)
	s.Nil(got.LiveListWorkflowID, "a null live-list workflow must stay nil")
	s.JSONEq(`{"query":"cto"}`, string(got.SearchFilters))
}

// TestSignalKeywords verifies the facet body is sent and the wrapped keywords
// are unwrapped to a slice for the caller.
func (s *SuperSearchTestSuite) TestSignalKeywords() {
	s.Router.Post(facetPath, func(w http.ResponseWriter, req *http.Request) {
		var received supersearch.FacetRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("technologies", received.Category)
		s.Equal("keywords", received.Field)

		_, _ = w.Write([]byte(`{"keywords":[{"keyword":"ai","count":42},{"keyword":"ml","count":7}]}`))
	})

	got, err := s.svc().SignalKeywords(context.Background(), supersearch.FacetRequest{
		Category: "technologies",
		Field:    "keywords",
		Prefix:   "a",
	})

	s.Require().NoError(err)
	s.Require().Len(got, 2)
	s.Equal("ai", got[0].Keyword)
	s.InDelta(42, got[0].Count, 0)
}

// TestHistory verifies the enrichment history decodes as a bare array of raw
// entries, and that the exact route is not shadowed by the :resource_id route.
func (s *SuperSearchTestSuite) TestHistory() {
	s.Router.Get(historyPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(resourceID, instantlytest.PathParam(req, "resource_id"))
		_, _ = w.Write([]byte(`[{"event":"created"},{"event":"run"}]`))
	})

	got, err := s.svc().History(context.Background(), resourceID)

	s.Require().NoError(err)
	s.Require().Len(got, 2)
	s.JSONEq(`{"event":"created"}`, string(got[0]))
	s.JSONEq(`{"event":"run"}`, string(got[1]))
}

// TestPathParametersAreEscaped verifies a caller-supplied resource identifier
// cannot rewrite the request path.
func (s *SuperSearchTestSuite) TestPathParametersAreEscaped() {
	tests := []struct {
		name     string
		call     func(svc *supersearch.Service) error
		body     string
		expected string
	}{
		{
			name: "get",
			call: func(svc *supersearch.Service) error {
				_, err := svc.Get(context.Background(), "../admin?x=1")
				return err
			},
			body:     `{}`,
			expected: "/api/v2/supersearch-enrichment/..%2Fadmin%3Fx=1",
		},
		{
			name: "update settings",
			call: func(svc *supersearch.Service) error {
				_, err := svc.UpdateSettings(context.Background(), "../admin?x=1", supersearch.SettingsRequest{})
				return err
			},
			body:     `{}`,
			expected: "/api/v2/supersearch-enrichment/..%2Fadmin%3Fx=1/settings",
		},
		{
			name: "ai in progress",
			call: func(svc *supersearch.Service) error {
				_, err := svc.AIInProgress(context.Background(), "../admin?x=1")
				return err
			},
			body:     `[]`,
			expected: "/api/v2/supersearch-enrichment/ai/..%2Fadmin%3Fx=1/in-progress",
		},
		{
			name: "history",
			call: func(svc *supersearch.Service) error {
				_, err := svc.History(context.Background(), "../admin?x=1")
				return err
			},
			body:     `[]`,
			expected: "/api/v2/supersearch-enrichment/history/..%2Fadmin%3Fx=1",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			var requestURI string

			client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
				&http.Client{Transport: instantlytest.RoundTripFunc(
					func(req *http.Request) (*http.Response, error) {
						requestURI = req.URL.EscapedPath()
						return instantlytest.JSONResponse(http.StatusOK, test.body), nil
					},
				)},
			))

			s.Require().NoError(test.call(supersearch.New(client)))
			s.Equal(test.expected, requestURI)
		})
	}
}

// TestFailures verifies every endpoint surfaces the documented API error.
func (s *SuperSearchTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "create", Method: http.MethodPost, Path: createPath, Status: http.StatusPaymentRequired,
			Call: func() error { _, err := svc.Create(ctx, supersearch.CreateRequest{}); return err },
		},
		{
			Name: "get", Method: http.MethodGet, Path: resourcePattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Get(ctx, "missing"); return err },
		},
		{
			Name: "updateSettings", Method: http.MethodPatch, Path: settingsPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.UpdateSettings(ctx, "missing", supersearch.SettingsRequest{}); return err },
		},
		{
			Name: "run", Method: http.MethodPost, Path: runPath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.Run(ctx, supersearch.RunRequest{}); return err },
		},
		{
			Name: "createAI", Method: http.MethodPost, Path: aiPath, Status: http.StatusPaymentRequired,
			Call: func() error { _, err := svc.CreateAI(ctx, supersearch.AIRequest{}); return err },
		},
		{
			Name: "aiInProgress", Method: http.MethodGet, Path: aiInProgressPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.AIInProgress(ctx, "missing"); return err },
		},
		{
			Name: "countLeads", Method: http.MethodPost, Path: countPath, Status: http.StatusTooManyRequests,
			Call: func() error { _, err := svc.CountLeads(ctx, supersearch.SearchRequest{}); return err },
		},
		{
			Name: "previewLeads", Method: http.MethodPost, Path: previewPath, Status: http.StatusForbidden,
			Call: func() error { _, err := svc.PreviewLeads(ctx, supersearch.SearchRequest{}); return err },
		},
		{
			Name: "enrichLeads", Method: http.MethodPost, Path: enrichPath, Status: http.StatusPaymentRequired,
			Call: func() error { _, err := svc.EnrichLeads(ctx, supersearch.EnrichLeadsRequest{}); return err },
		},
		{
			Name: "signalKeywords", Method: http.MethodPost, Path: facetPath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.SignalKeywords(ctx, supersearch.FacetRequest{}); return err },
		},
		{
			Name: "history", Method: http.MethodGet, Path: historyPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.History(ctx, "missing"); return err },
		},
	})
}

// svc builds a SuperSearch Enrichment service pointed at the suite's mock client.
func (s *SuperSearchTestSuite) svc() *supersearch.Service {
	return supersearch.New(s.Client)
}
