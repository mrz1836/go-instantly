package webhook_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/webhook"
)

// Router patterns and identifiers the webhook endpoints are exercised with. The
// patterns carry the full request path, including the /api/v2 prefix.
const (
	// listPath is the list/collection endpoint.
	listPath = "/api/v2/webhooks"

	// idPattern is the router pattern for the single-webhook endpoints.
	idPattern = "/api/v2/webhooks/:id"

	// eventTypesPath is the event-types endpoint.
	eventTypesPath = "/api/v2/webhooks/event-types"

	// testPattern is the router pattern for the test-delivery endpoint.
	testPattern = "/api/v2/webhooks/:id/test"

	// resumePattern is the router pattern for the resume endpoint.
	resumePattern = "/api/v2/webhooks/:id/resume"

	// webhookID identifies the webhook the single-webhook endpoints operate on.
	webhookID = "wh-1"
)

// webhookFixture is a spec-shaped webhook with every documented field populated,
// including the nullable ones.
const webhookFixture = `{
	"id": "wh-1",
	"organization": "org-1",
	"target_hook_url": "https://example.com/hook",
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"name": "My Webhook",
	"event_type": "reply_received",
	"campaign": "camp-1",
	"custom_interest_value": 2,
	"status": 1,
	"timestamp_error": "2026-08-02T10:00:00.000Z",
	"headers": {"X-Token": "abc"}
}`

// webhookFixtureNulls is the same webhook with every nullable field explicitly
// null, so an absent value stays distinguishable from a zero value.
const webhookFixtureNulls = `{
	"id": "wh-2",
	"organization": "org-1",
	"target_hook_url": "https://example.com/bare",
	"timestamp_created": "2026-08-01T11:00:00.000Z",
	"name": null,
	"event_type": null,
	"campaign": null,
	"custom_interest_value": null,
	"status": null,
	"timestamp_error": null
}`

// WebhookTestSuite exercises the Webhook API service against the mock router.
type WebhookTestSuite struct {
	instantlytest.Suite
}

// TestWebhookSuite runs the Webhook API suite.
func TestWebhookSuite(t *testing.T) {
	suite.Run(t, new(WebhookTestSuite))
}

// TestCreate verifies the create body reaches the API and the webhook decodes.
func (s *WebhookTestSuite) TestCreate() {
	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodPost, req.Method)
		s.Equal(listPath, req.URL.Path)

		var received webhook.CreateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("https://example.com/hook", received.TargetHookURL)
		s.Equal(webhook.EventReplyReceived, received.EventType)
		s.JSONEq(`{"X-Token":"abc"}`, string(received.Headers))

		_, _ = w.Write([]byte(webhookFixture))
	})

	got, err := s.svc().Create(context.Background(), webhook.CreateRequest{
		TargetHookURL: "https://example.com/hook",
		EventType:     webhook.EventReplyReceived,
		Name:          instantly.Ptr("My Webhook"),
		Headers:       json.RawMessage(`{"X-Token":"abc"}`),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(webhookID, got.ID)
	s.Require().NotNil(got.EventType)
	s.Equal(webhook.EventReplyReceived, *got.EventType)
}

// TestCreateOmitsUnsetFields verifies the optional fields are left out of the
// body when they are not set.
func (s *WebhookTestSuite) TestCreateOmitsUnsetFields() {
	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.NotContains(received, "event_type")
		s.NotContains(received, "campaign")
		s.NotContains(received, "headers")

		_, _ = w.Write([]byte(webhookFixture))
	})

	got, err := s.svc().Create(context.Background(), webhook.CreateRequest{
		TargetHookURL: "https://example.com/hook",
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
}

// TestList verifies a page decodes, including the enum values and the
// nullable-vs-zero fields.
func (s *WebhookTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("50", req.URL.Query().Get("limit"))
		s.Equal("reply_received", req.URL.Query().Get("event_type"))

		_, _ = w.Write([]byte(instantlytest.Page([]string{webhookFixture, webhookFixtureNulls}, "cursor-2")))
	})

	page, err := s.svc().List(context.Background(),
		webhook.WithLimit(50),
		webhook.WithEventType(webhook.EventReplyReceived),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)

	populated := page.Items[0]
	s.Equal(webhookID, populated.ID)
	s.Require().NotNil(populated.Status)
	s.Equal(webhook.StatusActive, *populated.Status)
	s.Require().NotNil(populated.CustomInterestValue)
	s.InDelta(2, *populated.CustomInterestValue, 0)
	s.JSONEq(`{"X-Token":"abc"}`, string(populated.Headers))

	// Nullable fields stay nil rather than collapsing to a zero value.
	bare := page.Items[1]
	s.Nil(bare.Name)
	s.Nil(bare.EventType)
	s.Nil(bare.Campaign)
	s.Nil(bare.Status)
	s.Nil(bare.TimestampError)
	s.Empty(bare.Headers)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string, even
// with a nil option.
func (s *WebhookTestSuite) TestListWithoutOptions() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery, "an unfiltered list must not send an empty query string")
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background(), nil)

	s.Require().NoError(err)
	s.Require().NotNil(page)
	s.Empty(page.Items)
}

// TestGet verifies a single webhook decodes.
func (s *WebhookTestSuite) TestGet() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(webhookID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(webhookFixture))
	})

	got, err := s.svc().Get(context.Background(), webhookID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("https://example.com/hook", got.TargetHookURL)
	s.Require().NotNil(got.TimestampError)
	s.Equal("2026-08-02T10:00:00.000Z", *got.TimestampError)
}

// TestUpdate verifies the patch body is sent and the updated webhook decodes.
func (s *WebhookTestSuite) TestUpdate() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(webhookID, instantlytest.PathParam(req, "id"))

		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("Renamed", received["name"])
		s.Equal("email_opened", received["event_type"])

		_, _ = w.Write([]byte(webhookFixture))
	})

	got, err := s.svc().Update(context.Background(), webhookID, webhook.UpdateRequest{
		Name:      instantly.Ptr("Renamed"),
		EventType: webhook.EventEmailOpened,
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(webhookID, got.ID)
}

// TestUpdateOmitsUnsetFields verifies an empty patch sends no field at all.
func (s *WebhookTestSuite) TestUpdateOmitsUnsetFields() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Empty(received, "an unset patch field must not be sent")

		_, _ = w.Write([]byte(webhookFixture))
	})

	got, err := s.svc().Update(context.Background(), webhookID, webhook.UpdateRequest{})

	s.Require().NoError(err)
	s.Require().NotNil(got)
}

// TestDelete verifies the deleted webhook is returned to the caller.
func (s *WebhookTestSuite) TestDelete() {
	s.Router.Delete(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(webhookID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(webhookFixture))
	})

	got, err := s.svc().Delete(context.Background(), webhookID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(webhookID, got.ID)
}

// TestEventTypes verifies the wrapped event types are unwrapped to a slice, and
// that the exact route is not shadowed by the :id route.
func (s *WebhookTestSuite) TestEventTypes() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, _ *http.Request) {
		// Registered first, yet the exact event-types route must win.
		w.WriteHeader(http.StatusInternalServerError)
	})
	s.Router.Get(eventTypesPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(eventTypesPath, req.URL.Path)
		_, _ = w.Write([]byte(
			`{"event_types":[{"id":"reply_received","label":"Reply Received","type":"reply_received"},` +
				`{"id":"email_sent","label":"Email Sent","type":"email_sent"}]}`,
		))
	})

	got, err := s.svc().EventTypes(context.Background())

	s.Require().NoError(err)
	s.Require().Len(got, 2)
	s.Equal("reply_received", got[0].ID)
	s.Equal("Reply Received", got[0].Label)
	s.Equal("email_sent", got[1].Type)
}

// TestTest verifies the test-delivery outcome decodes, with its nullable fields.
func (s *WebhookTestSuite) TestTest() {
	s.Router.Post(testPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(webhookID, instantlytest.PathParam(req, "id"))

		body, err := instantlytest.ReadAll(req)
		s.NoError(err)
		s.Empty(body, "a test delivery sends no request body")

		_, _ = w.Write([]byte(`{"success":true,"status_code":200,"response_time_ms":42,"message":"ok"}`))
	})

	got, err := s.svc().Test(context.Background(), webhookID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.True(got.Success)
	s.Require().NotNil(got.StatusCode)
	s.InDelta(200, *got.StatusCode, 0)
	s.Require().NotNil(got.ResponseTimeMS)
	s.InDelta(42, *got.ResponseTimeMS, 0)
}

// TestResume verifies the resume endpoint returns the reactivated webhook.
func (s *WebhookTestSuite) TestResume() {
	s.Router.Post(resumePattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(webhookID, instantlytest.PathParam(req, "id"))

		body, err := instantlytest.ReadAll(req)
		s.NoError(err)
		s.Empty(body, "resuming a webhook sends no request body")

		_, _ = w.Write([]byte(webhookFixture))
	})

	got, err := s.svc().Resume(context.Background(), webhookID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(webhookID, got.ID)
}

// TestPathParametersAreEscaped verifies a caller-supplied identifier cannot
// rewrite the request path.
func (s *WebhookTestSuite) TestPathParametersAreEscaped() {
	tests := []struct {
		name     string
		call     func(svc *webhook.Service) error
		expected string
	}{
		{
			name: "get",
			call: func(svc *webhook.Service) error {
				_, err := svc.Get(context.Background(), "../admin?x=1")
				return err
			},
			expected: "/api/v2/webhooks/..%2Fadmin%3Fx=1",
		},
		{
			name: "test",
			call: func(svc *webhook.Service) error {
				_, err := svc.Test(context.Background(), "../admin?x=1")
				return err
			},
			expected: "/api/v2/webhooks/..%2Fadmin%3Fx=1/test",
		},
		{
			name: "resume",
			call: func(svc *webhook.Service) error {
				_, err := svc.Resume(context.Background(), "../admin?x=1")
				return err
			},
			expected: "/api/v2/webhooks/..%2Fadmin%3Fx=1/resume",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			var requestURI string

			client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
				&http.Client{Transport: instantlytest.RoundTripFunc(
					func(req *http.Request) (*http.Response, error) {
						requestURI = req.URL.EscapedPath()
						return instantlytest.JSONResponse(http.StatusOK, webhookFixture), nil
					},
				)},
			))

			s.Require().NoError(test.call(webhook.New(client)))
			s.Equal(test.expected, requestURI)
		})
	}
}

// TestListOptions verifies each documented list query parameter is rendered by
// exactly one option, under the key and value the API expects.
func (s *WebhookTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option webhook.ListOption
		key    string
		value  string
	}{
		{"limit", webhook.WithLimit(50), "limit", "50"},
		{"starting after", webhook.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"campaign", webhook.WithCampaign("camp-1"), "campaign", "camp-1"},
		{"event type", webhook.WithEventType(webhook.EventEmailBounced), "event_type", "email_bounced"},
	}

	s.Require().Len(tests, 4, "every documented list query parameter needs an option")

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
func (s *WebhookTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "create", Method: http.MethodPost, Path: listPath, Status: http.StatusPaymentRequired,
			Call: func() error { _, err := svc.Create(ctx, webhook.CreateRequest{}); return err },
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
			Call: func() error { _, err := svc.Update(ctx, "missing", webhook.UpdateRequest{}); return err },
		},
		{
			Name: "delete", Method: http.MethodDelete, Path: idPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Delete(ctx, "missing"); return err },
		},
		{
			Name: "eventTypes", Method: http.MethodGet, Path: eventTypesPath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.EventTypes(ctx); return err },
		},
		{
			Name: "test", Method: http.MethodPost, Path: testPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Test(ctx, "missing"); return err },
		},
		{
			Name: "resume", Method: http.MethodPost, Path: resumePattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Resume(ctx, "missing"); return err },
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor parses a valid
// timestamp and reports an error for an unparseable one.
func (s *WebhookTestSuite) TestParsedTimestampCreated() {
	got, err := (&webhook.Webhook{TimestampCreated: "2026-08-01T10:00:00.000Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&webhook.Webhook{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// svc builds a Webhook service pointed at the suite's mock client.
func (s *WebhookTestSuite) svc() *webhook.Service {
	return webhook.New(s.Client)
}
