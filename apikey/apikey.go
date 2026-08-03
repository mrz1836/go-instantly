package apikey

import (
	"context"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the API Key API.
const basePath = "/api/v2/api-keys"

// Service provides access to the Instantly.ai V2 API Key API.
type Service struct {
	client *instantly.Client
}

// New builds an API Key API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// APIKey is a single API key returned by the Instantly.ai V2 API.
//
// The API declares every field required, so none is nullable. Key carries the
// secret token and is only returned in full when the key is first created.
type APIKey struct {
	// ID is the unique identifier of the API key.
	ID string `json:"id"`

	// Name is the human-readable name of the API key.
	Name string `json:"name"`

	// Scopes are the permissions the API key has been granted.
	Scopes []Scope `json:"scopes"`

	// Key is the secret token used to authenticate as this API key.
	Key string `json:"key"`

	// OrganizationID identifies the organization the API key belongs to.
	OrganizationID string `json:"organization_id"`

	// TimestampCreated is when the API key was created.
	TimestampCreated string `json:"timestamp_created"`

	// TimestampUpdated is when the API key was last updated.
	TimestampUpdated string `json:"timestamp_updated"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded key re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (k *APIKey) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, k.TimestampCreated)
}

// ListResponse is a single page of API keys.
//
// It aliases instantly.Page[APIKey], the cursor-paginated envelope every
// resource shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[APIKey]

// CreateRequest is the body of a create-API-key request. Both fields are
// required by the API.
type CreateRequest struct {
	// Name is the human-readable name of the API key.
	Name string `json:"name"`

	// Scopes are the permissions to grant the API key.
	Scopes []Scope `json:"scopes"`
}

// Create issues a new API key with the requested scopes and returns it.
//
// The returned Key is the only time the full secret token is exposed, so store
// it when the call succeeds.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*APIKey, error) {
	return instantly.PostResult[APIKey](ctx, s.client, basePath, req)
}

// List returns a single page of API keys filtered by the supplied options.
//
// Pagination is cursor based: pass the returned NextStartingAfter back with
// WithStartingAfter to fetch the following page.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}

// Delete deletes an API key and returns the key that was deleted.
func (s *Service) Delete(ctx context.Context, id string) (*APIKey, error) {
	return instantly.DeleteResult[APIKey](ctx, s.client, instantly.JoinPath(basePath, id))
}
