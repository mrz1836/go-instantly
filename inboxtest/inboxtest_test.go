package inboxtest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/inboxtest"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// Router patterns and identifiers the inbox-placement-test endpoints are
// exercised with. The patterns carry the full request path, including the
// /api/v2 prefix.
const (
	// listPath is the list/collection endpoint.
	listPath = "/api/v2/inbox-placement-tests"

	// idPattern is the router pattern for the single-test endpoints.
	idPattern = "/api/v2/inbox-placement-tests/:id"

	// espOptionsPath is the email-service-provider-options endpoint.
	espOptionsPath = "/api/v2/inbox-placement-tests/email-service-provider-options"

	// testID identifies the test the single-test endpoints operate on.
	testID = "test-uuid-1"
)

// testFixture is a spec-shaped inbox placement test with every documented field
// populated, including the nullable ones and the nested payloads.
const testFixture = `{
	"id": "test-uuid-1",
	"organization_id": "org-uuid-1",
	"name": "My Inbox Placement Test",
	"type": 1,
	"sending_method": 1,
	"email_subject": "My Email Subject",
	"email_body": "Hi, this is my email body",
	"emails": ["john@doe.com"],
	"recipients": ["johndoe@instantly.ai"],
	"timestamp_created": "2026-08-01T19:32:22.234Z",
	"campaign_id": "campaign-uuid-1",
	"delivery_mode": 1,
	"description": "This is a test description",
	"not_sending_status": "daily_limits_hit",
	"status": 1,
	"tags": ["tag-uuid-1"],
	"test_code": "ptid_abc",
	"text_only": true,
	"timestamp_next_run": "2026-08-02T19:32:22.234Z",
	"recipients_labels": [
		{"esp": "Google", "region": "North America", "sub_region": "US", "type": "Professional"}
	],
	"schedule": {"days": {"2": true, "3": true}, "timezone": "America/Chihuahua", "timing": {"from": "02:30"}},
	"automations": [{"when": {"condition": "placement_goes_below", "condition_value": 80}, "then": {"pause": true}}],
	"metadata": {"campaign": {"id": "campaign-id", "name": "Campaign Name"}}
}`

// testFixtureNulls is the same test with every nullable field explicitly null,
// so an absent value stays distinguishable from a zero value.
const testFixtureNulls = `{
	"id": "test-uuid-2",
	"organization_id": "org-uuid-1",
	"name": "Bare Test",
	"type": 2,
	"sending_method": 2,
	"email_subject": "Subject",
	"email_body": "Body",
	"emails": ["a@b.c"],
	"recipients": ["seed@instantly.ai"],
	"timestamp_created": "2026-08-01T20:00:00.000Z",
	"campaign_id": null,
	"delivery_mode": null,
	"description": null,
	"not_sending_status": null,
	"status": null,
	"tags": null,
	"test_code": null,
	"text_only": null,
	"timestamp_next_run": null
}`

// InboxTestSuite exercises the Inbox Placement Test API service against the mock
// router.
type InboxTestSuite struct {
	instantlytest.Suite
}

// TestInboxTestSuite runs the Inbox Placement Test API suite.
func TestInboxTestSuite(t *testing.T) {
	suite.Run(t, new(InboxTestSuite))
}

// TestCreate verifies the create body reaches the API and the test decodes.
func (s *InboxTestSuite) TestCreate() {
	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodPost, req.Method)
		s.Equal(listPath, req.URL.Path)

		var received inboxtest.CreateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("My Inbox Placement Test", received.Name)
		s.Equal(inboxtest.TypeOneTime, received.Type)
		s.Equal(inboxtest.SendingFromInstantly, received.SendingMethod)
		s.Equal([]string{"john@doe.com"}, received.Emails)
		if s.NotNil(received.DeliveryMode) {
			s.Equal(inboxtest.DeliveryOneByOne, *received.DeliveryMode)
		}
		s.JSONEq(`{"timezone":"Etc/GMT+12"}`, string(received.Schedule))

		_, _ = w.Write([]byte(testFixture))
	})

	got, err := s.svc().Create(context.Background(), inboxtest.CreateRequest{
		Name:          "My Inbox Placement Test",
		Type:          inboxtest.TypeOneTime,
		SendingMethod: inboxtest.SendingFromInstantly,
		EmailSubject:  "My Email Subject",
		EmailBody:     "Hi, this is my email body",
		Emails:        []string{"john@doe.com"},
		DeliveryMode:  instantly.Ptr(inboxtest.DeliveryOneByOne),
		Schedule:      json.RawMessage(`{"timezone":"Etc/GMT+12"}`),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(testID, got.ID)
	s.Equal(inboxtest.TypeOneTime, got.Type)
}

// TestCreateOmitsUnsetFields verifies the optional fields are left out of the
// body when they are not set.
func (s *InboxTestSuite) TestCreateOmitsUnsetFields() {
	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.NotContains(received, "description")
		s.NotContains(received, "delivery_mode")
		s.NotContains(received, "schedule")
		s.NotContains(received, "run_immediately")

		_, _ = w.Write([]byte(testFixture))
	})

	got, err := s.svc().Create(context.Background(), inboxtest.CreateRequest{
		Name:          "My Inbox Placement Test",
		Type:          inboxtest.TypeOneTime,
		SendingMethod: inboxtest.SendingFromInstantly,
		EmailSubject:  "My Email Subject",
		EmailBody:     "Hi, this is my email body",
		Emails:        []string{"john@doe.com"},
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
}

// TestList verifies a page decodes, including nullable-vs-zero fields, the enum
// values, and the nested payloads.
func (s *InboxTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(http.MethodGet, req.Method)
		s.Equal(listPath, req.URL.Path)
		s.Equal("50", req.URL.Query().Get("limit"))
		s.Equal("1", req.URL.Query().Get("status"))

		_, _ = w.Write([]byte(instantlytest.Page([]string{testFixture, testFixtureNulls}, "cursor-2")))
	})

	page, err := s.svc().List(context.Background(),
		inboxtest.WithLimit(50),
		inboxtest.WithStatus(inboxtest.StatusActive),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)

	populated := page.Items[0]
	s.Equal(testID, populated.ID)
	s.Equal(inboxtest.TypeOneTime, populated.Type)
	s.Equal(inboxtest.SendingFromInstantly, populated.SendingMethod)
	s.Require().NotNil(populated.DeliveryMode)
	s.Equal(inboxtest.DeliveryOneByOne, *populated.DeliveryMode)
	s.Require().NotNil(populated.Status)
	s.Equal(inboxtest.StatusActive, *populated.Status)
	s.Require().Len(populated.RecipientsLabels, 1)
	s.Equal("Google", populated.RecipientsLabels[0].ESP)
	s.Equal("Professional", populated.RecipientsLabels[0].Type)
	s.JSONEq(`{"campaign":{"id":"campaign-id","name":"Campaign Name"}}`, string(populated.Metadata))

	// Nullable fields stay nil rather than collapsing to a zero value.
	bare := page.Items[1]
	s.Equal(inboxtest.TypeAutomated, bare.Type)
	s.Nil(bare.CampaignID)
	s.Nil(bare.DeliveryMode)
	s.Nil(bare.Status)
	s.Nil(bare.TextOnly)
	s.Empty(bare.Tags)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string.
func (s *InboxTestSuite) TestListWithoutOptions() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery, "an unfiltered list must not send an empty query string")
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background(), nil)

	s.Require().NoError(err)
	s.Require().NotNil(page)
	s.Empty(page.Items)
}

// TestGet verifies a single test decodes.
func (s *InboxTestSuite) TestGet() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))
		s.Empty(req.URL.RawQuery, "a plain Get must not send with_metadata")

		_, _ = w.Write([]byte(testFixture))
	})

	got, err := s.svc().Get(context.Background(), testID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(testID, got.ID)
	s.Require().NotNil(got.Description)
	s.Equal("This is a test description", *got.Description)
	s.JSONEq(`{"from":"02:30"}`, timingOf(got.Schedule))
}

// TestGetWithMetadata verifies the with_metadata option is sent and the metadata
// payload is preserved.
func (s *InboxTestSuite) TestGetWithMetadata() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))
		s.Equal("true", req.URL.Query().Get("with_metadata"))

		_, _ = w.Write([]byte(testFixture))
	})

	got, err := s.svc().Get(context.Background(), testID, inboxtest.WithMetadata())

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.JSONEq(`{"campaign":{"id":"campaign-id","name":"Campaign Name"}}`, string(got.Metadata))
}

// TestUpdate verifies the patch body is sent and the updated test decodes.
func (s *InboxTestSuite) TestUpdate() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))

		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("Renamed", received["name"])
		s.InDelta(2, received["status"], 0)

		_, _ = w.Write([]byte(testFixture))
	})

	got, err := s.svc().Update(context.Background(), testID, inboxtest.UpdateRequest{
		Name:   "Renamed",
		Status: instantly.Ptr(inboxtest.StatusPaused),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(testID, got.ID)
}

// TestUpdateOmitsUnsetFields verifies an empty patch sends no field at all, so an
// untouched field is never overwritten.
func (s *InboxTestSuite) TestUpdateOmitsUnsetFields() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Empty(received, "an unset patch field must not be sent")

		_, _ = w.Write([]byte(testFixture))
	})

	got, err := s.svc().Update(context.Background(), testID, inboxtest.UpdateRequest{})

	s.Require().NoError(err)
	s.Require().NotNil(got)
}

// TestDelete verifies the deleted test is returned to the caller.
func (s *InboxTestSuite) TestDelete() {
	s.Router.Delete(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(testID, instantlytest.PathParam(req, "id"))

		body, err := instantlytest.ReadAll(req)
		s.NoError(err)
		s.Empty(body, "deleting a test sends no request body")

		_, _ = w.Write([]byte(testFixture))
	})

	got, err := s.svc().Delete(context.Background(), testID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(testID, got.ID)
}

// TestESPOptions verifies the ESP-options endpoint decodes as a bare array, and
// that its exact route is not shadowed by the :id route.
func (s *InboxTestSuite) TestESPOptions() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, _ *http.Request) {
		// Registered first, yet the exact ESP-options route must win.
		w.WriteHeader(http.StatusInternalServerError)
	})
	s.Router.Get(espOptionsPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(espOptionsPath, req.URL.Path)

		_, _ = w.Write([]byte(
			`[{"esp":"Google","region":"North America","sub_region":"US","type":"Professional"},` +
				`{"esp":"Microsoft","region":"Europe","sub_region":"UK","type":"Personal"}]`,
		))
	})

	got, err := s.svc().ESPOptions(context.Background())

	s.Require().NoError(err)
	s.Require().Len(got, 2)
	s.Equal("Google", got[0].ESP)
	s.Equal("US", got[0].SubRegion)
	s.Equal("Microsoft", got[1].ESP)
}

// TestPathParametersAreEscaped verifies a caller-supplied identifier cannot
// rewrite the request path.
func (s *InboxTestSuite) TestPathParametersAreEscaped() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, testFixture), nil
			},
		)},
	))

	_, err := inboxtest.New(client).Get(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/inbox-placement-tests/..%2Fadmin%3Fx=1", requestURI)
}

// TestListOptions verifies each documented list query parameter is rendered by
// exactly one option, under the key and value the API expects.
func (s *InboxTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option inboxtest.ListOption
		key    string
		value  string
	}{
		{"limit", inboxtest.WithLimit(50), "limit", "50"},
		{"starting after", inboxtest.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"search", inboxtest.WithSearch("my test"), "search", "my test"},
		{"status", inboxtest.WithStatus(inboxtest.StatusCompleted), "status", "3"},
		{"sort order", inboxtest.WithSortOrder(instantly.SortOrderAsc), "sort_order", "asc"},
	}

	s.Require().Len(tests, 5, "every documented list query parameter needs an option")

	for _, test := range tests {
		s.Run(test.name, func() {
			q := instantly.NewQuery()
			test.option(q)

			s.Require().Equal(1, q.Len(), "an option must render exactly one query parameter")
			s.Equal(test.value, q.Get(test.key))
		})
	}
}

// TestGetOptions verifies the get option renders exactly one query parameter.
func (s *InboxTestSuite) TestGetOptions() {
	tests := []struct {
		name   string
		option inboxtest.GetOption
		key    string
		value  string
	}{
		{"with metadata", inboxtest.WithMetadata(), "with_metadata", "true"},
	}

	s.Require().Len(tests, 1, "every documented get query parameter needs an option")

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
func (s *InboxTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "create", Method: http.MethodPost, Path: listPath, Status: http.StatusPaymentRequired,
			Call: func() error { _, err := svc.Create(ctx, inboxtest.CreateRequest{}); return err },
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
			Call: func() error { _, err := svc.Update(ctx, "missing", inboxtest.UpdateRequest{}); return err },
		},
		{
			Name: "delete", Method: http.MethodDelete, Path: idPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Delete(ctx, "missing"); return err },
		},
		{
			Name: "espOptions", Method: http.MethodGet, Path: espOptionsPath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.ESPOptions(ctx); return err },
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor parses a valid
// timestamp and reports an error for an unparseable one.
func (s *InboxTestSuite) TestParsedTimestampCreated() {
	got, err := (&inboxtest.Test{TimestampCreated: "2026-08-01T19:32:22.234Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&inboxtest.Test{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// svc builds an Inbox Placement Test service pointed at the suite's mock client.
func (s *InboxTestSuite) svc() *inboxtest.Service {
	return inboxtest.New(s.Client)
}

// timingOf extracts the timing object from a raw schedule payload so a test can
// assert on a nested field without modeling the whole schedule.
func timingOf(schedule json.RawMessage) string {
	var decoded struct {
		Timing json.RawMessage `json:"timing"`
	}
	_ = json.Unmarshal(schedule, &decoded)

	return string(decoded.Timing)
}
