package leadlabel

import (
	"context"
	"encoding/json"
	"time"

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

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded label re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (l *LeadLabel) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, l.TimestampCreated)
}

// ListResponse is a single page of lead labels.
//
// It aliases instantly.Page[LeadLabel], the cursor-paginated envelope every
// resource shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[LeadLabel]

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
	return instantly.PostResult[LeadLabel](ctx, s.client, basePath, req)
}

// List returns a single page of lead labels filtered by the supplied options.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}

// Get returns a single lead label by its unique identifier.
func (s *Service) Get(ctx context.Context, id string) (*LeadLabel, error) {
	return instantly.GetResult[LeadLabel](ctx, s.client, instantly.JoinPath(basePath, id))
}

// Update patches a lead label and returns its updated state.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*LeadLabel, error) {
	return instantly.PatchResult[LeadLabel](ctx, s.client, instantly.JoinPath(basePath, id), req)
}

// Delete deletes a lead label and returns the label that was deleted.
func (s *Service) Delete(ctx context.Context, id string) (*LeadLabel, error) {
	return instantly.DeleteResult[LeadLabel](ctx, s.client, instantly.JoinPath(basePath, id))
}

// TestAIReplyLabel reports which label the AI reply classifier would assign to
// the given reply text.
func (s *Service) TestAIReplyLabel(ctx context.Context, req AIReplyLabelRequest) (*AIReplyLabelResponse, error) {
	return instantly.PostResult[AIReplyLabelResponse](ctx, s.client, basePath+"/ai-reply-label", req)
}
