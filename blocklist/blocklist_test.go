package blocklist_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/blocklist"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// Router patterns and identifiers the block-list-entry endpoints are exercised
// with. The patterns carry the full request path, including the /api/v2 prefix.
const (
	// listPath is the list/collection endpoint (GET, POST, and delete-all DELETE).
	listPath = "/api/v2/block-lists-entries"

	// idPattern is the router pattern for the single-entry endpoints.
	idPattern = "/api/v2/block-lists-entries/:id"

	// bulkCreatePath is the bulk-create endpoint.
	bulkCreatePath = "/api/v2/block-lists-entries/bulk-create"

	// bulkDeletePath is the bulk-delete endpoint.
	bulkDeletePath = "/api/v2/block-lists-entries/bulk-delete"

	// downloadPath is the CSV download endpoint.
	downloadPath = "/api/v2/block-lists-entries/download"

	// entryID identifies the entry the single-entry endpoints operate on.
	entryID = "bl-1"
)

// entryFixture is a spec-shaped block list entry with every documented field
// populated. The API declares no nullable fields on an entry.
const entryFixture = `{
	"id": "bl-1",
	"bl_value": "spam.example.com",
	"is_domain": true,
	"organization_id": "org-1",
	"timestamp_created": "2026-08-01T10:00:00.000Z"
}`

// BlockListTestSuite exercises the Block List Entry API service against the mock
// router.
type BlockListTestSuite struct {
	instantlytest.Suite
}

// TestBlockListSuite runs the Block List Entry API suite.
func TestBlockListSuite(t *testing.T) {
	suite.Run(t, new(BlockListTestSuite))
}

// TestCreate verifies the create body reaches the API and the entry decodes.
func (s *BlockListTestSuite) TestCreate() {
	s.Router.Post(listPath, func(w http.ResponseWriter, req *http.Request) {
		var received blocklist.CreateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("spam.example.com", received.BLValue)

		_, _ = w.Write([]byte(entryFixture))
	})

	got, err := s.svc().Create(context.Background(), blocklist.CreateRequest{
		BLValue: "spam.example.com",
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(entryID, got.ID)
	s.True(got.IsDomain)
}

// TestList verifies a page decodes and the filter options are sent.
func (s *BlockListTestSuite) TestList() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("50", req.URL.Query().Get("limit"))
		s.Equal("true", req.URL.Query().Get("domains_only"))
		s.Equal("example.com", req.URL.Query().Get("search"))

		_, _ = w.Write([]byte(instantlytest.Page([]string{entryFixture}, "cursor-2")))
	})

	page, err := s.svc().List(context.Background(),
		blocklist.WithLimit(50),
		blocklist.WithDomainsOnly(true),
		blocklist.WithSearch("example.com"),
	)

	s.Require().NoError(err)
	s.Require().Len(page.Items, 1)
	s.Equal("cursor-2", page.NextStartingAfter)
	s.Equal("spam.example.com", page.Items[0].BLValue)
}

// TestListWithoutOptions verifies an unfiltered list sends no query string, even
// with a nil option.
func (s *BlockListTestSuite) TestListWithoutOptions() {
	s.Router.Get(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Empty(req.URL.RawQuery, "an unfiltered list must not send an empty query string")
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	page, err := s.svc().List(context.Background(), nil)

	s.Require().NoError(err)
	s.Require().NotNil(page)
	s.Empty(page.Items)
}

// TestGet verifies a single entry decodes.
func (s *BlockListTestSuite) TestGet() {
	s.Router.Get(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(entryID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(entryFixture))
	})

	got, err := s.svc().Get(context.Background(), entryID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal("spam.example.com", got.BLValue)
}

// TestUpdate verifies the patch body is sent and the updated entry decodes.
func (s *BlockListTestSuite) TestUpdate() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(entryID, instantlytest.PathParam(req, "id"))

		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal("new@example.com", received["bl_value"])

		_, _ = w.Write([]byte(entryFixture))
	})

	got, err := s.svc().Update(context.Background(), entryID, blocklist.UpdateRequest{
		BLValue: "new@example.com",
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(entryID, got.ID)
}

// TestUpdateOmitsUnsetFields verifies an empty patch sends no field at all.
func (s *BlockListTestSuite) TestUpdateOmitsUnsetFields() {
	s.Router.Patch(idPattern, func(w http.ResponseWriter, req *http.Request) {
		var received map[string]any
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Empty(received, "an unset patch field must not be sent")

		_, _ = w.Write([]byte(entryFixture))
	})

	got, err := s.svc().Update(context.Background(), entryID, blocklist.UpdateRequest{})

	s.Require().NoError(err)
	s.Require().NotNil(got)
}

// TestDelete verifies the deleted entry is returned to the caller.
func (s *BlockListTestSuite) TestDelete() {
	s.Router.Delete(idPattern, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(entryID, instantlytest.PathParam(req, "id"))
		_, _ = w.Write([]byte(entryFixture))
	})

	got, err := s.svc().Delete(context.Background(), entryID)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(entryID, got.ID)
}

// TestDeleteAll verifies the filter is sent, the delete-all route is not shadowed
// by the :id route, and the deleted entries decode as a bare array.
func (s *BlockListTestSuite) TestDeleteAll() {
	s.Router.Delete(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(listPath, req.URL.Path)
		s.Equal("true", req.URL.Query().Get("domains_only"))
		s.Equal("example.com", req.URL.Query().Get("search"))

		_, _ = w.Write([]byte(`[` + entryFixture + `]`))
	})

	got, err := s.svc().DeleteAll(context.Background(), true, "example.com")

	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal(entryID, got[0].ID)
}

// TestDeleteAllNoSearch verifies an empty search term is not sent, while the
// domains_only bool always is.
func (s *BlockListTestSuite) TestDeleteAllNoSearch() {
	s.Router.Delete(listPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal("false", req.URL.Query().Get("domains_only"))
		s.Empty(req.URL.Query().Get("search"), "an empty search term must not be sent")

		_, _ = w.Write([]byte(`[]`))
	})

	got, err := s.svc().DeleteAll(context.Background(), false, "")

	s.Require().NoError(err)
	s.Empty(got)
}

// TestBulkCreate verifies the bulk body is sent and the split result decodes.
func (s *BlockListTestSuite) TestBulkCreate() {
	s.Router.Post(bulkCreatePath, func(w http.ResponseWriter, req *http.Request) {
		var received blocklist.BulkCreateRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal([]string{"a@x.com", "bad"}, received.BLValues)

		_, _ = w.Write([]byte(`{"items":[` + entryFixture + `],"valid_count":1,"invalid_count":1}`))
	})

	got, err := s.svc().BulkCreate(context.Background(), blocklist.BulkCreateRequest{
		BLValues: []string{"a@x.com", "bad"},
	})

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Require().Len(got.Items, 1)
	s.InDelta(1, got.ValidCount, 0)
	s.InDelta(1, got.InvalidCount, 0)
}

// TestBulkDelete verifies the ids are sent and the deleted entries decode as a
// bare array.
func (s *BlockListTestSuite) TestBulkDelete() {
	s.Router.Post(bulkDeletePath, func(w http.ResponseWriter, req *http.Request) {
		var received blocklist.BulkDeleteRequest
		s.NoError(json.NewDecoder(req.Body).Decode(&received))
		s.Equal([]string{"bl-1", "bl-2"}, received.IDs)

		_, _ = w.Write([]byte(`[` + entryFixture + `]`))
	})

	got, err := s.svc().BulkDelete(context.Background(), blocklist.BulkDeleteRequest{
		IDs: []string{"bl-1", "bl-2"},
	})

	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Equal(entryID, got[0].ID)
}

// TestDownload verifies the CSV bytes are returned verbatim and the filter is
// sent.
func (s *BlockListTestSuite) TestDownload() {
	csv := "bl_value,is_domain\nspam.example.com,true\n"

	s.Router.Get(downloadPath, func(w http.ResponseWriter, req *http.Request) {
		s.Equal(downloadPath, req.URL.Path)
		s.Equal("true", req.URL.Query().Get("domains_only"))

		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte(csv))
	})

	got, err := s.svc().Download(context.Background(), true, "")

	s.Require().NoError(err)
	s.Equal(csv, string(got), "the raw CSV bytes must be returned unchanged")
}

// TestPathParametersAreEscaped verifies a caller-supplied identifier cannot
// rewrite the request path.
func (s *BlockListTestSuite) TestPathParametersAreEscaped() {
	var requestURI string

	client := instantly.NewClient(instantlytest.APIKey, instantly.WithHTTPClient(
		&http.Client{Transport: instantlytest.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				requestURI = req.URL.EscapedPath()
				return instantlytest.JSONResponse(http.StatusOK, entryFixture), nil
			},
		)},
	))

	_, err := blocklist.New(client).Get(context.Background(), "../admin?x=1")

	s.Require().NoError(err)
	s.Equal("/api/v2/block-lists-entries/..%2Fadmin%3Fx=1", requestURI)
}

// TestListOptions verifies each documented list query parameter is rendered by
// exactly one option, under the key and value the API expects.
func (s *BlockListTestSuite) TestListOptions() {
	tests := []struct {
		name   string
		option blocklist.ListOption
		key    string
		value  string
	}{
		{"limit", blocklist.WithLimit(50), "limit", "50"},
		{"starting after", blocklist.WithStartingAfter("cursor-2"), "starting_after", "cursor-2"},
		{"domains only", blocklist.WithDomainsOnly(true), "domains_only", "true"},
		{"search", blocklist.WithSearch("example.com"), "search", "example.com"},
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
func (s *BlockListTestSuite) TestFailures() {
	svc, ctx := s.svc(), context.Background()
	s.RunFailures([]instantlytest.FailureCase{
		{
			Name: "create", Method: http.MethodPost, Path: listPath, Status: http.StatusPaymentRequired,
			Call: func() error { _, err := svc.Create(ctx, blocklist.CreateRequest{}); return err },
		},
		{
			Name: "list", Method: http.MethodGet, Path: listPath, Status: http.StatusTooManyRequests,
			Call: func() error { _, err := svc.List(ctx); return err },
		},
		{
			Name: "deleteAll", Method: http.MethodDelete, Path: listPath, Status: http.StatusForbidden,
			Call: func() error { _, err := svc.DeleteAll(ctx, false, ""); return err },
		},
		{
			Name: "get", Method: http.MethodGet, Path: idPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Get(ctx, "missing"); return err },
		},
		{
			Name: "update", Method: http.MethodPatch, Path: idPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Update(ctx, "missing", blocklist.UpdateRequest{}); return err },
		},
		{
			Name: "delete", Method: http.MethodDelete, Path: idPattern, Status: http.StatusNotFound,
			Call: func() error { _, err := svc.Delete(ctx, "missing"); return err },
		},
		{
			Name: "bulkCreate", Method: http.MethodPost, Path: bulkCreatePath, Status: http.StatusPaymentRequired,
			Call: func() error { _, err := svc.BulkCreate(ctx, blocklist.BulkCreateRequest{}); return err },
		},
		{
			Name: "bulkDelete", Method: http.MethodPost, Path: bulkDeletePath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.BulkDelete(ctx, blocklist.BulkDeleteRequest{}); return err },
		},
		{
			Name: "download", Method: http.MethodGet, Path: downloadPath, Status: http.StatusUnauthorized,
			Call: func() error { _, err := svc.Download(ctx, false, ""); return err },
		},
	})
}

// TestParsedTimestampCreated verifies the RFC 3339 accessor parses a valid
// timestamp and reports an error for an unparseable one.
func (s *BlockListTestSuite) TestParsedTimestampCreated() {
	got, err := (&blocklist.Entry{TimestampCreated: "2026-08-01T10:00:00.000Z"}).ParsedTimestampCreated()
	s.Require().NoError(err)
	s.Equal(2026, got.Year())

	_, err = (&blocklist.Entry{TimestampCreated: "not-a-timestamp"}).ParsedTimestampCreated()
	s.Require().Error(err)
}

// svc builds a Block List Entry service pointed at the suite's mock client.
func (s *BlockListTestSuite) svc() *blocklist.Service {
	return blocklist.New(s.Client)
}
