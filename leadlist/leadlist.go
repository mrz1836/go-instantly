// Package leadlist provides typed access to the Instantly.ai V2 Lead List API.
//
// It wraps the /api/v2/lead-lists endpoints: creating, listing, reading,
// patching, and deleting lead lists, plus reading a list's verification stats.
//
//	svc := leadlist.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, leadlist.WithLimit(50))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package leadlist

import (
	"context"
	"encoding/json"
	"iter"
	"net/url"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Lead List API.
const basePath = "/api/v2/lead-lists"

// Service provides access to the Instantly.ai V2 Lead List API.
type Service struct {
	client *instantly.Client
}

// New builds a Lead List API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// LeadList is a single lead list returned by the Instantly.ai V2 API.
type LeadList struct {
	// ID is the unique identifier of the list.
	ID string `json:"id"`

	// Name is the name of the list.
	Name string `json:"name"`

	// OrganizationID is the organization the list belongs to.
	OrganizationID string `json:"organization_id"`

	// TimestampCreated is when the list was created.
	TimestampCreated string `json:"timestamp_created"`

	// HasEnrichmentTask reports whether the list has an enrichment task.
	HasEnrichmentTask *bool `json:"has_enrichment_task,omitempty"`

	// OwnedBy is the owner ID of the list.
	OwnedBy *string `json:"owned_by,omitempty"`
}

// ListResponse is a single page of lead lists.
type ListResponse struct {
	// Items are the lead lists on this page.
	Items []LeadList `json:"items"`

	// NextStartingAfter is the cursor for the following page, and is empty on
	// the last page.
	NextStartingAfter string `json:"next_starting_after,omitempty"`
}

// VerificationStats are the verification statistics of a lead list.
type VerificationStats struct {
	// TotalLeads is the total number of leads in the list.
	TotalLeads int64 `json:"total_leads"`

	// Stats carries the per-status verification breakdown, preserved verbatim.
	Stats json.RawMessage `json:"stats,omitempty"`
}

// CreateRequest is the body of a create-lead-list request.
type CreateRequest struct {
	// Name is the name of the list. Required.
	Name string `json:"name"`

	// HasEnrichmentTask attaches an enrichment task to the list when set.
	HasEnrichmentTask *bool `json:"has_enrichment_task,omitempty"`

	// OwnedBy is the owner ID to assign the list to.
	OwnedBy *string `json:"owned_by,omitempty"`
}

// UpdateRequest is the body of a patch-lead-list request. No field is required;
// an omitted field leaves the current value unchanged.
type UpdateRequest struct {
	// Name is the name of the list.
	Name string `json:"name,omitempty"`

	// HasEnrichmentTask attaches an enrichment task to the list when set.
	HasEnrichmentTask *bool `json:"has_enrichment_task,omitempty"`

	// OwnedBy is the owner ID to assign the list to.
	OwnedBy *string `json:"owned_by,omitempty"`
}

// Create adds a new lead list and returns it.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*LeadList, error) {
	out := &LeadList{}
	if err := s.client.Post(ctx, basePath, req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// List returns a single page of lead lists filtered by the supplied options.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	q := instantly.NewQuery()
	for _, opt := range opts {
		if opt != nil {
			opt(q)
		}
	}

	out := &ListResponse{}
	if err := s.client.Get(ctx, q.Path(basePath), out); err != nil {
		return nil, err
	}

	return out, nil
}

// Get returns a single lead list by its unique identifier.
func (s *Service) Get(ctx context.Context, id string) (*LeadList, error) {
	out := &LeadList{}
	if err := s.client.Get(ctx, basePath+"/"+url.PathEscape(id), out); err != nil {
		return nil, err
	}

	return out, nil
}

// Update patches a lead list and returns its updated state.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*LeadList, error) {
	out := &LeadList{}
	if err := s.client.Patch(ctx, basePath+"/"+url.PathEscape(id), req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// Delete deletes a lead list and returns the list that was deleted.
func (s *Service) Delete(ctx context.Context, id string) (*LeadList, error) {
	out := &LeadList{}
	if err := s.client.Delete(ctx, basePath+"/"+url.PathEscape(id), out); err != nil {
		return nil, err
	}

	return out, nil
}

// VerificationStats returns the verification statistics of a lead list.
func (s *Service) VerificationStats(ctx context.Context, id string) (*VerificationStats, error) {
	out := &VerificationStats{}
	if err := s.client.Get(ctx, basePath+"/"+url.PathEscape(id)+"/verification-stats", out); err != nil {
		return nil, err
	}

	return out, nil
}

// ListIter walks every page of List, yielding each list with a nil error, or a
// nil *LeadList with the first error.
func (s *Service) ListIter(ctx context.Context, opts ...ListOption) iter.Seq2[*LeadList, error] {
	return instantly.Iterate(ctx, func(ctx context.Context, cursor string) ([]LeadList, string, error) {
		pageOpts := opts
		if cursor != "" {
			pageOpts = append(append([]ListOption(nil), opts...), WithStartingAfter(cursor))
		}

		page, err := s.List(ctx, pageOpts...)
		if err != nil {
			return nil, "", err
		}

		return page.Items, page.NextStartingAfter, nil
	})
}
