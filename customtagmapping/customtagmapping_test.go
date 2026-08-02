package customtagmapping_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/customtagmapping"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// listPath is the router pattern for the custom-tag-mappings list endpoint.
const listPath = "/api/v2/custom-tag-mappings"

// mappingFixture is a spec-shaped custom tag mapping with every documented field
// populated. The API declares no nullable fields on a mapping.
const mappingFixture = `{
	"id": "ctm-1",
	"tag_id": "tag-1",
	"resource_id": "camp-1",
	"resource_type": 2,
	"organization_id": "org-1",
	"timestamp_created": "2026-08-01T10:00:00.000Z"
}`

// CustomTagMappingTestSuite exercises the Custom Tag Mapping API service against
// the mock router.
type CustomTagMappingTestSuite struct {
	instantlytest.Suite
}

// TestCustomTagMappingSuite runs the Custom Tag Mapping API suite.
func TestCustomTagMappingSuite(t *testing.T) {
	suite.Run(t, new(CustomTagMappingTestSuite))
}

// TestList verifies a page decodes, including the enum, and the options are sent.
func (s *CustomTagMappingTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(listPath, req.URL.Path)
		s.Equal("50", req.URL.Query().Get("limit"))
		s.Equal("camp-1,camp-2", req.URL.Query().Get("resource_ids"))

		_, _ = w.Write([]byte(instantlytest.Page([]string{mappingFixture}, "cursor-2")))
	})

	page, err := s.svc().List(context.Background(),
		customtagmapping.WithLimit(50),
		customtagmapping.WithResourceIDs("camp-1,camp-2"),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 1)
	s.Equal("cursor-2", page.NextStartingAfter)

	got := page.Items[0]
	s.Equal("ctm-1", got.ID)
	s.Equal("tag-1", got.TagID)
	s.Equal("camp-1", got.ResourceID)
	s.Equal(customtagmapping.ResourceTypeCampaign, got.ResourceType)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string, even
// with a nil option.
func (s *CustomTagMappingTestSuite) TestListWithoutOptions() {
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
func (s *CustomTagMappingTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option customtagmapping.ListOption
		key    string
		value  string
	}{
		{"limit", customtagmapping.WithLimit(50), "limit", "50"},
		{"starting after", customtagmapping.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"resource ids", customtagmapping.WithResourceIDs("camp-1,camp-2"), "resource_ids", "camp-1,camp-2"},
	}

	s.Require().Len(tests, 3, "every documented list query parameter needs an option")

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
func (s *CustomTagMappingTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "list", Method: http.MethodGet, Path: listPath, Status: http.StatusTooManyRequests,
			Call: func() error { _, err := svc.List(ctx); return err },
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor parses a valid
// timestamp and reports an error for an unparseable one.
func (s *CustomTagMappingTestSuite) TestParsedTimestampCreated() {
	got, err := (&customtagmapping.Mapping{TimestampCreated: "2026-08-01T10:00:00.000Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&customtagmapping.Mapping{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// svc builds a Custom Tag Mapping service pointed at the suite's mock client.
func (s *CustomTagMappingTestSuite) svc() *customtagmapping.Service {
	return customtagmapping.New(s.Client)
}
