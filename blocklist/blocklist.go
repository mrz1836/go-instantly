package blocklist

import (
	"context"
	"net/http"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Block List Entry API.
const basePath = "/api/v2/block-lists-entries"

// Service provides access to the Instantly.ai V2 Block List Entry API.
type Service struct {
	client *instantly.Client
}

// New builds a Block List Entry API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Entry is a single block list entry returned by the Instantly.ai V2 API.
type Entry struct {
	// ID is the unique identifier of the entry.
	ID string `json:"id"`

	// BLValue is the blocked value, either an email address or a domain.
	BLValue string `json:"bl_value"`

	// IsDomain reports whether the entry blocks a whole domain rather than a
	// single address.
	IsDomain bool `json:"is_domain"`

	// OrganizationID identifies the organization the entry belongs to.
	OrganizationID string `json:"organization_id"`

	// TimestampCreated is when the entry was created.
	TimestampCreated string `json:"timestamp_created"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded entry re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (e *Entry) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, e.TimestampCreated)
}

// ListResponse is a single page of block list entries.
//
// It aliases instantly.Page[Entry], the cursor-paginated envelope every resource
// shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[Entry]

// BulkCreateResult reports the outcome of a bulk-create request, splitting the
// created entries from the counts of valid and invalid inputs.
type BulkCreateResult struct {
	// Items are the entries that were created.
	Items []Entry `json:"items"`

	// ValidCount is how many of the supplied values were valid.
	ValidCount float64 `json:"valid_count"`

	// InvalidCount is how many of the supplied values were invalid.
	InvalidCount float64 `json:"invalid_count"`
}

// CreateRequest is the body of a create-block-list-entry request.
type CreateRequest struct {
	// BLValue is the email address or domain to block. Required.
	BLValue string `json:"bl_value"`
}

// UpdateRequest is the body of a patch-block-list-entry request. No field is
// required; an omitted field leaves the current value unchanged.
type UpdateRequest struct {
	// BLValue is the email address or domain to block.
	BLValue string `json:"bl_value,omitempty"`
}

// BulkCreateRequest is the body of a bulk-create-block-list-entries request.
type BulkCreateRequest struct {
	// BLValues are the email addresses or domains to block. Required.
	BLValues []string `json:"bl_values"`
}

// BulkDeleteRequest is the body of a bulk-delete-block-list-entries request.
type BulkDeleteRequest struct {
	// IDs are the entries to delete. Required.
	IDs []string `json:"ids"`
}

// Create adds a new block list entry and returns it.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Entry, error) {
	return instantly.PostResult[Entry](ctx, s.client, basePath, req)
}

// List returns a single page of block list entries filtered by the supplied
// options.
//
// Pagination is cursor based: pass the returned NextStartingAfter back with
// WithStartingAfter to fetch the following page.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}

// Get returns a single block list entry by its unique identifier.
func (s *Service) Get(ctx context.Context, id string) (*Entry, error) {
	return instantly.GetResult[Entry](ctx, s.client, instantly.JoinPath(basePath, id))
}

// Update patches a block list entry and returns its updated state.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Entry, error) {
	return instantly.PatchResult[Entry](ctx, s.client, instantly.JoinPath(basePath, id), req)
}

// Delete deletes a block list entry and returns the entry that was deleted.
func (s *Service) Delete(ctx context.Context, id string) (*Entry, error) {
	return instantly.DeleteResult[Entry](ctx, s.client, instantly.JoinPath(basePath, id))
}

// DeleteAll deletes every block list entry matching the filter and returns the
// entries that were deleted.
//
// domainsOnly restricts the deletion to domain entries, and search restricts it
// to entries matching a term; an empty search applies no term. With domainsOnly
// false and an empty search, every entry is deleted.
func (s *Service) DeleteAll(ctx context.Context, domainsOnly bool, search string) ([]Entry, error) {
	var out []Entry
	if err := s.client.Delete(ctx, filter(domainsOnly, search).Path(basePath), &out); err != nil {
		return nil, err
	}

	return out, nil
}

// BulkCreate adds several block list entries at once, reporting the created
// entries and the counts of valid and invalid inputs.
func (s *Service) BulkCreate(ctx context.Context, req BulkCreateRequest) (*BulkCreateResult, error) {
	return instantly.PostResult[BulkCreateResult](ctx, s.client, basePath+"/bulk-create", req)
}

// BulkDelete deletes several block list entries by id and returns the entries
// that were deleted.
func (s *Service) BulkDelete(ctx context.Context, req BulkDeleteRequest) ([]Entry, error) {
	var out []Entry
	if err := s.client.Post(ctx, basePath+"/bulk-delete", req, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// Download returns the whole block list as CSV bytes, filtered the same way as
// DeleteAll.
//
// The endpoint answers with text/csv rather than JSON, so the raw bytes are
// returned for the caller to write to a file or parse.
func (s *Service) Download(ctx context.Context, domainsOnly bool, search string) ([]byte, error) {
	return s.client.DoRaw(ctx, http.MethodGet, filter(domainsOnly, search).Path(basePath+"/download"), nil)
}

// filter builds the domains_only/search query the delete-all and download
// endpoints share, setting the search term only when one is supplied.
func filter(domainsOnly bool, search string) *instantly.Query {
	q := instantly.NewQuery().SetBool("domains_only", domainsOnly)
	if search != "" {
		q.SetString("search", search)
	}

	return q
}
