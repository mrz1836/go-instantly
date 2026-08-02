package customtag_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/customtag"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// Router patterns and identifiers the custom-tag endpoints are exercised with.
// The patterns carry the full request path, including the /api/v2 prefix.
const (
	// listPath is the list/collection endpoint.
	listPath = "/api/v2/custom-tags"

	// idPattern is the router pattern for the single-tag endpoints.
	idPattern = "/api/v2/custom-tags/:id"

	// togglePath is the toggle-resource endpoint.
	togglePath = "/api/v2/custom-tags/toggle-resource"

	// tagID identifies the tag the single-tag endpoints operate on.
	tagID = "tag-1"
)

// tagFixture is a spec-shaped custom tag with every documented field populated,
// including the nullable description.
const tagFixture = `{
	"id": "tag-1",
	"label": "VIP",
	"organization_id": "org-1",
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"timestamp_updated": "2026-08-01T11:00:00.000Z",
	"description": "High-value leads"
}`

// tagFixtureNulls is the same tag with the nullable description explicitly null,
// so an absent value stays distinguishable from a zero value.
const tagFixtureNulls = `{
	"id": "tag-2",
	"label": "Bare",
	"organization_id": "org-1",
	"timestamp_created": "2026-08-01T10:00:00.000Z",
	"timestamp_updated": "2026-08-01T11:00:00.000Z",
	"description": null
}`

// CustomTagTestSuite exercises the Custom Tag API service against the mock router.
type CustomTagTestSuite struct {
	instantlytest.Suite
}

// TestCustomTagSuite runs the Custom Tag API suite.
func TestCustomTagSuite(t *testing.T) {
	suite.Run(t, new(CustomTagTestSuite))
}

// TestCreate verifies the create body reaches the API and the tag decodes.
func (s *CustomTagTestSuite) TestCreate() {
	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		var received customtag.CreateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("VIP", received.Label)
		if s.NotNil(received.Description) {
			s.Equal("High-value leads", *received.Description)
		}

		_, _ = w.Write([]byte(tagFixture))
	})

	got, err := s.svc().Create(context.Background(), customtag.CreateRequest{
		Label:       "VIP",
		Description: instantly.Ptr("High-value leads"),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(tagID, got.ID)
	s.Require().NotNil(got.Description)
	s.Equal("High-value leads", *got.Description)
}

// TestList verifies a page decodes, including the nullable-vs-zero description.
func (s *CustomTagTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("50", req.URL.Query().Get("limit"))
		s.Equal("vip", req.URL.Query().Get("search"))
		s.Equal("res-1,res-2", req.URL.Query().Get("resource_ids"))

		_, _ = w.Write([]byte(instantlytest.Page([]string{tagFixture, tagFixtureNulls}, "cursor-2")))
	})

	page, err := s.svc().List(context.Background(),
		customtag.WithLimit(50),
		customtag.WithSearch("vip"),
		customtag.WithResourceIDs("res-1,res-2"),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 2)
	s.Equal("cursor-2", page.NextStartingAfter)

	populated := page.Items[0]
	s.Equal(tagID, populated.ID)
	s.Require().NotNil(populated.Description)
	s.Equal("High-value leads", *populated.Description)

	// The nullable description stays nil rather than collapsing to a zero value.
	s.Nil(page.Items[1].Description)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string, even
// with a nil option.
func (s *CustomTagTestSuite) TestListWithoutOptions() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery, "an unfiltered list must not send an empty query string")
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background(), nil)

	s.Require().NoError(err)
	s.Require().NotNil(page)
	s.Empty(page.Items)
}

// TestGet verifies a single tag decodes.
func (s *CustomTagTestSuite) TestGet() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(tagID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(tagFixture))
	})

	got, err := s.svc().Get(context.Background(), tagID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("VIP", got.Label)
}

// TestUpdate verifies the patch body is sent and the updated tag decodes.
func (s *CustomTagTestSuite) TestUpdate() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(tagID, instantlytest.PathParam(req, "id"))

		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("Renamed", received["label"])

		_, _ = w.Write([]byte(tagFixture))
	})

	got, err := s.svc().Update(context.Background(), tagID, customtag.UpdateRequest{
		Label: "Renamed",
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(tagID, got.ID)
}

// TestUpdateOmitsUnsetFields verifies an empty patch sends no field at all.
func (s *CustomTagTestSuite) TestUpdateOmitsUnsetFields() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Empty(received, "an unset patch field must not be sent")

		_, _ = w.Write([]byte(tagFixture))
	})

	got, err := s.svc().Update(context.Background(), tagID, customtag.UpdateRequest{})

	s.Require().NoError(err)
	s.Require().NotNil(got)
}

// TestDelete verifies the deleted tag is returned to the caller.
func (s *CustomTagTestSuite) TestDelete() {
	s.Router.Delete(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(tagID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(tagFixture))
	})

	got, err := s.svc().Delete(context.Background(), tagID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(tagID, got.ID)
}

// TestToggle verifies the toggle body is sent, including the enum and raw filter,
// and that the exact route is not shadowed by the :id route.
func (s *CustomTagTestSuite) TestToggle() {
	s.Router.Post(idPattern, func(w http.ResponseWriter, _ *http.Request) {
		// Registered first, yet the exact toggle-resource route must win.
		w.WriteHeader(http.StatusInternalServerError)
	})
	s.Router.Post(togglePath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(togglePath, req.URL.Path)

		var received customtag.ToggleRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal([]string{tagID}, received.TagIDs)
		s.Equal(customtag.ResourceTypeCampaign, received.ResourceType)
		s.True(received.Assign)
		s.JSONEq(`"ACC_FILTER_PAUSED"`, string(received.Filter))

		_, _ = w.Write([]byte(`{"success":true}`))
	})

	got, err := s.svc().Toggle(context.Background(), customtag.ToggleRequest{
		TagIDs:       []string{tagID},
		ResourceType: customtag.ResourceTypeCampaign,
		Assign:       true,
		ResourceIDs:  []string{"camp-1"},
		Filter:       json.RawMessage(`"ACC_FILTER_PAUSED"`),
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.True(got.Success)
}

// TestPathParametersAreEscaped verifies a caller-supplied identifier cannot
// rewrite the request path.
func (s *CustomTagTestSuite) TestPathParametersAreEscaped() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, tagFixture), nil
			},
		)},
	))

	_, err := customtag.New(client).Get(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/custom-tags/..%2Fadmin%3Fx=1", requestURI)
}

// TestListOptions verifies each documented list query parameter is rendered by
// exactly one option, under the key and value the API expects.
func (s *CustomTagTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option customtag.ListOption
		key    string
		value  string
	}{
		{"limit", customtag.WithLimit(50), "limit", "50"},
		{"starting after", customtag.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"search", customtag.WithSearch("vip"), "search", "vip"},
		{"resource ids", customtag.WithResourceIDs("res-1,res-2"), "resource_ids", "res-1,res-2"},
		{"tag ids", customtag.WithTagIDs("tag-1,tag-2"), "tag_ids", "tag-1,tag-2"},
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

// TestFailures verifies every endpoint surfaces the documented API error.
func (s *CustomTagTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "create", Method: http.MethodPost, Path: listPath, Status: http.StatusPaymentRequired,
			Call: func() error { _, err := svc.Create(ctx, customtag.CreateRequest{}); return err },
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
			Call: func() error { _, err := svc.Update(ctx, "missing", customtag.UpdateRequest{}); return err },
		},
		{
			Name: "delete", Method: http.MethodDelete, Path: idPattern, Status: http.StatusForbidden,
			Call: func() error { _, err := svc.Delete(ctx, "missing"); return err },
		},
		{
			Name: "toggle", Method: http.MethodPost, Path: togglePath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.Toggle(ctx, customtag.ToggleRequest{}); return err },
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor parses a valid
// timestamp and reports an error for an unparseable one.
func (s *CustomTagTestSuite) TestParsedTimestampCreated() {
	got, err := (&customtag.Tag{TimestampCreated: "2026-08-01T10:00:00.000Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&customtag.Tag{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// svc builds a Custom Tag service pointed at the suite's mock client.
func (s *CustomTagTestSuite) svc() *customtag.Service {
	return customtag.New(s.Client)
}
