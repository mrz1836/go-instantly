// Package leadlabel provides typed access to the Instantly.ai V2 Lead Label API.
//
// It wraps the /api/v2/lead-labels endpoints: creating, listing, reading,
// patching, and deleting lead labels, plus testing which label an AI reply
// classifier would assign to a reply.
//
//	svc := leadlabel.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, leadlabel.WithLimit(50))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package leadlabel

import (
	"context"
	"encoding/json"
	"iter"
	"net/url"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Lead Label API.
const basePath = "/api/v2/lead-labels"

// Service provides access to the Instantly.ai V2 Lead Label API.
type Service struct {
	client *instantly.Client
}

// New builds a Lead Label API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// LeadLabel is a single lead label returned by the Instantly.ai V2 API.
type LeadLabel struct {
	// ID is the unique identifier of the label.
	ID string `json:"id"`

	// Label is the display label.
	Label string `json:"label"`

	// InterestStatusLabel is the interest-status label associated with the label.
	InterestStatusLabel string `json:"interest_status_label"`

	// InterestStatus is the numeric interest status the label maps to.
	InterestStatus int64 `json:"interest_status"`

	// CreatedBy is the user ID that created the label.
	CreatedBy string `json:"created_by"`

	// OrganizationID is the organization the label belongs to.
	OrganizationID string `json:"organization_id"`

	// TimestampCreated is when the label was created.
	TimestampCreated string `json:"timestamp_created"`

	// Description is the detailed description of the label.
	Description *string `json:"description,omitempty"`

	// UseWithAI reports whether the label is used by the AI reply classifier.
	UseWithAI *bool `json:"use_with_ai,omitempty"`
}

// ListResponse is a single page of lead labels.
type ListResponse struct {
	// Items are the lead labels on this page.
	Items []LeadLabel `json:"items"`

	// NextStartingAfter is the cursor for the following page, and is empty on
	// the last page.
	NextStartingAfter string `json:"next_starting_after,omitempty"`
}

// CreateRequest is the body of a create-lead-label request.
type CreateRequest struct {
	// Label is the display label. Required.
	Label string `json:"label"`

	// InterestStatusLabel is the interest-status label. Required.
	InterestStatusLabel string `json:"interest_status_label"`

	// Description is the detailed description of the label.
	Description *string `json:"description,omitempty"`

	// UseWithAI marks the label for use by the AI reply classifier.
	UseWithAI *bool `json:"use_with_ai,omitempty"`
}

// UpdateRequest is the body of a patch-lead-label request. No field is required;
// an omitted field leaves the current value unchanged.
type UpdateRequest struct {
	// Label is the display label.
	Label string `json:"label,omitempty"`

	// InterestStatusLabel is the interest-status label.
	InterestStatusLabel string `json:"interest_status_label,omitempty"`

	// Description is the detailed description of the label.
	Description *string `json:"description,omitempty"`

	// UseWithAI marks the label for use by the AI reply classifier.
	UseWithAI *bool `json:"use_with_ai,omitempty"`
}

// AIReplyLabelRequest is the body of a test-ai-reply-label request.
type AIReplyLabelRequest struct {
	// ReplyText is the reply text to classify. Required.
	ReplyText string `json:"reply_text"`
}

// AIReplyLabelResponse is the outcome of a test-ai-reply-label request.
type AIReplyLabelResponse struct {
	// Result carries the classification result, preserved verbatim.
	Result json.RawMessage `json:"result,omitempty"`

	// CustomLabelsConsidered carries the labels the classifier considered,
	// preserved verbatim.
	CustomLabelsConsidered json.RawMessage `json:"custom_labels_considered,omitempty"`
}

// Create adds a new lead label and returns it.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*LeadLabel, error) {
	out := &LeadLabel{}
	if err := s.client.Post(ctx, basePath, req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// List returns a single page of lead labels filtered by the supplied options.
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

// Get returns a single lead label by its unique identifier.
func (s *Service) Get(ctx context.Context, id string) (*LeadLabel, error) {
	out := &LeadLabel{}
	if err := s.client.Get(ctx, basePath+"/"+url.PathEscape(id), out); err != nil {
		return nil, err
	}

	return out, nil
}

// Update patches a lead label and returns its updated state.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*LeadLabel, error) {
	out := &LeadLabel{}
	if err := s.client.Patch(ctx, basePath+"/"+url.PathEscape(id), req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// Delete deletes a lead label and returns the label that was deleted.
func (s *Service) Delete(ctx context.Context, id string) (*LeadLabel, error) {
	out := &LeadLabel{}
	if err := s.client.Delete(ctx, basePath+"/"+url.PathEscape(id), out); err != nil {
		return nil, err
	}

	return out, nil
}

// TestAIReplyLabel reports which label the AI reply classifier would assign to
// the given reply text.
func (s *Service) TestAIReplyLabel(ctx context.Context, req AIReplyLabelRequest) (*AIReplyLabelResponse, error) {
	out := &AIReplyLabelResponse{}
	if err := s.client.Post(ctx, basePath+"/ai-reply-label", req, out); err != nil {
		return nil, err
	}

	return out, nil
}

// ListIter walks every page of List, yielding each label with a nil error, or a
// nil *LeadLabel with the first error.
func (s *Service) ListIter(ctx context.Context, opts ...ListOption) iter.Seq2[*LeadLabel, error] {
	return instantly.Iterate(ctx, func(ctx context.Context, cursor string) ([]LeadLabel, string, error) {
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
