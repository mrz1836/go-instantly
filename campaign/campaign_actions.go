package campaign

import (
	"context"
	"encoding/json"

	"github.com/mrz1836/go-instantly"
)

// DuplicateRequest is the body of a duplicate-campaign request.
type DuplicateRequest struct {
	// Name is the name of the duplicated campaign. Optional; the API generates
	// one when omitted.
	Name string `json:"name,omitempty"`
}

// AddVariablesRequest is the body of an add-variables request.
type AddVariablesRequest struct {
	// Variables carries the variables to add, sent verbatim. It is a JSON array
	// of variable objects.
	Variables json.RawMessage `json:"variables"`
}

// SendingStatus is the sending status of a campaign.
//
// Both fields carry the raw API payload, which the API does not document as a
// fixed schema, so they are preserved verbatim.
type SendingStatus struct {
	// Summary is a summary of the campaign's sending status.
	Summary json.RawMessage `json:"summary,omitempty"`

	// Diagnostics are the detailed sending diagnostics.
	Diagnostics json.RawMessage `json:"diagnostics,omitempty"`
}

// searchResponse is the wrapper the search-by-contact endpoint returns.
type searchResponse struct {
	Items []Campaign `json:"items"`
}

// Activate activates (starts or resumes) a campaign and returns it.
func (s *Service) Activate(ctx context.Context, id string) (*Campaign, error) {
	return s.action(ctx, id, "activate")
}

// Pause stops (pauses) a campaign and returns it.
func (s *Service) Pause(ctx context.Context, id string) (*Campaign, error) {
	return s.action(ctx, id, "pause")
}

// Duplicate copies a campaign and returns the new campaign.
func (s *Service) Duplicate(ctx context.Context, id string, req DuplicateRequest) (*Campaign, error) {
	return instantly.PostResult[Campaign](ctx, s.client, instantly.JoinPath(basePath, id, "duplicate"), req)
}

// Share shares a campaign. The endpoint returns no content, so a nil return is
// the only signal of success.
func (s *Service) Share(ctx context.Context, id string) error {
	return s.client.Post(ctx, instantly.JoinPath(basePath, id, "share"), nil, nil)
}

// Export exports a campaign to its JSON representation and returns it.
func (s *Service) Export(ctx context.Context, id string) (*Campaign, error) {
	return instantly.PostResult[Campaign](ctx, s.client, instantly.JoinPath(basePath, id, "export"), nil)
}

// CreateFromExport creates a campaign from a shared or exported one, and returns
// the new campaign.
//
// The API does not document a fixed request schema for this endpoint, so the
// exported payload is sent verbatim; pass the JSON returned by Export.
func (s *Service) CreateFromExport(ctx context.Context, id string, body json.RawMessage) (*Campaign, error) {
	return instantly.PostResult[Campaign](ctx, s.client, instantly.JoinPath(basePath, id, "from-export"), body)
}

// AddVariables adds variables to a campaign and returns its updated state.
func (s *Service) AddVariables(ctx context.Context, id string, req AddVariablesRequest) (*Campaign, error) {
	return instantly.PostResult[Campaign](ctx, s.client, instantly.JoinPath(basePath, id, "variables"), req)
}

// SendingStatus returns the sending status of a campaign.
func (s *Service) SendingStatus(ctx context.Context, id string) (*SendingStatus, error) {
	return instantly.GetResult[SendingStatus](ctx, s.client, instantly.JoinPath(basePath, id, "sending-status"))
}

// CountLaunched returns the number of launched campaigns.
func (s *Service) CountLaunched(ctx context.Context) (int64, error) {
	out, err := instantly.GetResult[instantly.CountResponse](ctx, s.client, basePath+"/count-launched")
	if err != nil {
		return 0, err
	}

	return out.Count, nil
}

// SearchByContact returns the campaigns a lead, identified by email, belongs to.
func (s *Service) SearchByContact(
	ctx context.Context, contact string, opts ...SearchOption,
) ([]Campaign, error) {
	q := instantly.ApplyOptions(opts...).SetString("search", contact)

	out, err := instantly.GetResult[searchResponse](ctx, s.client, q.Path(basePath+"/search-by-contact"))
	if err != nil {
		return nil, err
	}

	return out.Items, nil
}

// action performs a POST lifecycle action that returns the campaign.
func (s *Service) action(ctx context.Context, id, verb string) (*Campaign, error) {
	return instantly.PostResult[Campaign](ctx, s.client, instantly.JoinPath(basePath, id, verb), nil)
}
