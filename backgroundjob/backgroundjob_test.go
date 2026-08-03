package backgroundjob_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/backgroundjob"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// listPath is the router pattern for the background-jobs list endpoint.
const listPath = "/api/v2/background-jobs"

// idPath is the router pattern for the single-background-job endpoint.
const idPath = "/api/v2/background-jobs/:id"

// jobFixture is a spec-shaped background job with every documented field
// populated, including the nullable ones.
const jobFixture = `{
	"id": "job-1",
	"workspace_id": "ws-1",
	"type": "move-leads",
	"progress": 42,
	"status": "in-progress",
	"created_at": "2026-08-01T10:00:00.000Z",
	"updated_at": "2026-08-01T10:05:00.000Z",
	"entity_type": "list",
	"entity_id": "list-1",
	"user_id": "user-1",
	"data": {"success_count": 3}
}`

// jobFixtureNulls is a spec-shaped background job with the nullable entity_id
// and user_id explicitly null and the optional entity_type omitted.
const jobFixtureNulls = `{
	"id": "job-2",
	"workspace_id": "ws-1",
	"type": "export-leads",
	"progress": 100,
	"status": "success",
	"created_at": "2026-08-01T11:00:00.000Z",
	"updated_at": "2026-08-01T11:30:00.000Z",
	"entity_id": null,
	"user_id": null
}`

// BackgroundJobTestSuite exercises the Background Job API service against the
// mock router.
type BackgroundJobTestSuite struct {
	instantlytest.Suite
}

// TestBackgroundJobSuite runs the Background Job API suite.
func TestBackgroundJobSuite(t *testing.T) {
	suite.Run(t, new(BackgroundJobTestSuite))
}

// TestList verifies a page decodes, including the enums and free-form data, and
// the options are sent.
func (s *BackgroundJobTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(listPath, req.URL.Path)
		s.Equal("50", req.URL.Query().Get("limit"))
		s.Equal("move-leads", req.URL.Query().Get("type"))
		s.Equal("pending,in-progress", req.URL.Query().Get("status"))

		_, _ = w.Write([]byte(instantlytest.Page([]string{jobFixture}, "cursor-2")))
	})

	page, err := s.svc().List(context.Background(),
		backgroundjob.WithLimit(50),
		backgroundjob.WithType(backgroundjob.TypeMoveLeads),
		backgroundjob.WithStatus("pending,in-progress"),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 1)
	s.Equal("cursor-2", page.NextStartingAfter)

	got := page.Items[0]
	s.Equal("job-1", got.ID)
	s.Equal(backgroundjob.TypeMoveLeads, got.Type)
	s.Equal(backgroundjob.StatusInProgress, got.Status)
	s.Equal(backgroundjob.EntityTypeList, got.EntityType)
	s.InEpsilon(42.0, got.Progress, 1e-9)
	s.Require().NotNil(got.EntityID)
	s.Equal("list-1", *got.EntityID)
	s.JSONEq(`{"success_count":3}`, string(got.Data))
}

// TestListNulls verifies a job with the nullable fields null decodes to nil
// pointers and an empty optional enum.
func (s *BackgroundJobTestSuite) TestListNulls() {
	s.Router.Get(listPath, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(instantlytest.Page([]string{jobFixtureNulls}, "")))
	})

	page, err := s.svc().List(context.Background())

	s.Require().NoError(err)
	s.Require().Len(page.Items, 1)

	got := page.Items[0]
	s.Equal("job-2", got.ID)
	s.Equal(backgroundjob.StatusSuccess, got.Status)
	s.Empty(got.EntityType)
	s.Nil(got.EntityID)
	s.Nil(got.UserID)
	s.Nil(got.Data)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string, even
// with a nil option.
func (s *BackgroundJobTestSuite) TestListWithoutOptions() {
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
func (s *BackgroundJobTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option backgroundjob.ListOption
		key    string
		value  string
	}{
		{"limit", backgroundjob.WithLimit(50), "limit", "50"},
		{"starting after", backgroundjob.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"ids", backgroundjob.WithIDs("j1,j2"), "ids", "j1,j2"},
		{"included ids", backgroundjob.WithIncludedIDs("j3,j4"), "included_ids", "j3,j4"},
		{"excluded ids", backgroundjob.WithExcludedIDs("j5,j6"), "excluded_ids", "j5,j6"},
		{"type", backgroundjob.WithType(backgroundjob.TypeImportLeads), "type", "import-leads"},
		{"entity type", backgroundjob.WithEntityType(backgroundjob.EntityTypeCampaign), "entity_type", "campaign"},
		{"entity id", backgroundjob.WithEntityID("ent-1"), "entity_id", "ent-1"},
		{"status", backgroundjob.WithStatus("pending,paused"), "status", "pending,paused"},
		{"sort column", backgroundjob.WithSortColumn(backgroundjob.SortColumnUpdatedAt), "sort_column", "updated_at"},
		{"sort order", backgroundjob.WithSortOrder(instantly.SortOrderDesc), "sort_order", "desc"},
	}

	s.Require().Len(tests, 11, "every documented list query parameter needs an option")

	for _, test := range tests {
		s.Run(test.name, func() {
			q := instantly.NewQuery()
			test.option(q)

			s.Require().Equal(1, q.Len(), "an option must render exactly one query parameter")
			s.Equal(test.value, q.Get(test.key))
		})
	}
}

// TestGet verifies a single job decodes and the data_fields option is sent.
func (s *BackgroundJobTestSuite) TestGet() {
	s.Router.Get(idPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("job-1", instantlytest.PathParam(req, "id"))
		s.Equal("success_count,failed_count", req.URL.Query().Get("data_fields"))

		_, _ = w.Write([]byte(jobFixture))
	})

	job, err := s.svc().Get(context.Background(), "job-1",
		backgroundjob.WithDataFields("success_count,failed_count"),
	)

	s.Require().NoError(err)
	s.Equal("job-1", job.ID)
	s.Equal(backgroundjob.TypeMoveLeads, job.Type)
}

// TestGetWithoutOptions verifies a get with no options sends no query string.
func (s *BackgroundJobTestSuite) TestGetWithoutOptions() {
	s.Router.Get(idPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery, "a get without options must not send an empty query string")
		_, _ = w.Write([]byte(jobFixture))
	})

	job, err := s.svc().Get(context.Background(), "job-1", nil)

	s.Require().NoError(err)
	s.Equal("job-1", job.ID)
}

// TestGetOptions verifies each documented get query parameter is rendered by
// exactly one option, under the key and value the API expects.
func (s *BackgroundJobTestSuite) TestGetOptions() {
	tests := []struct {
		name   string
		option backgroundjob.GetOption
		key    string
		value  string
	}{
		{"data fields", backgroundjob.WithDataFields("success_count,failed_count"), "data_fields", "success_count,failed_count"},
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

// TestGetPathEscape verifies an id carrying reserved characters is escaped into
// the request path rather than altering it.
func (s *BackgroundJobTestSuite) TestGetPathEscape() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, jobFixture), nil
			},
		)},
	))

	_, err := backgroundjob.New(client).Get(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/background-jobs/..%2Fadmin%3Fx=1", requestURI)
}

// TestFailures verifies every endpoint surfaces the documented API error.
func (s *BackgroundJobTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "list", Method: http.MethodGet, Path: listPath, Status: http.StatusTooManyRequests,
			Call: func() error { _, err := svc.List(ctx); return err },
		},
		{
			Name: "get", Method: http.MethodGet, Path: idPath, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Get(ctx, "job-1"); return err },
		},
	})
}

// svc builds a Background Job service pointed at the suite's mock client.
func (s *BackgroundJobTestSuite) svc() *backgroundjob.Service {
	return backgroundjob.New(s.Client)
}
