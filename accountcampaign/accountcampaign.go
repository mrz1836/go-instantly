// Package accountcampaign provides typed access to the Instantly.ai V2
// account-campaign-mappings endpoint: the campaigns a sending account belongs
// to.
//
//	svc := accountcampaign.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, "sender@example.com", accountcampaign.WithLimit(50))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package accountcampaign

import (
	"context"
	"iter"
	"net/url"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the account-campaign-mappings API.
const basePath = "/api/v2/account-campaign-mappings"

// Service provides access to the account-campaign-mappings endpoint.
type Service struct {
	client *instantly.Client
}

// New builds an account-campaign-mappings service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Mapping is a single campaign a sending account is associated with.
type Mapping struct {
	// CampaignID is the unique identifier of the campaign.
	CampaignID string `json:"campaign_id"`

	// CampaignName is the name of the campaign.
	CampaignName string `json:"campaign_name"`

	// Status is the status of the campaign.
	Status int64 `json:"status"`

	// TimestampCreated is when the mapping was created.
	TimestampCreated string `json:"timestamp_created"`
}

// ListResponse is a single page of account-campaign mappings.
type ListResponse struct {
	// Items are the mappings on this page.
	Items []Mapping `json:"items"`

	// NextStartingAfter is the cursor for the following page, and is empty on
	// the last page.
	NextStartingAfter string `json:"next_starting_after,omitempty"`
}

// ListOption customizes a List request.
type ListOption func(*instantly.Query)

// WithLimit sets the maximum number of mappings returned in a single page.
func WithLimit(limit int) ListOption {
	return func(q *instantly.Query) {
		q.SetInt("limit", limit)
	}
}

// WithStartingAfter sets the pagination cursor to resume from.
func WithStartingAfter(cursor string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("starting_after", cursor)
	}
}

// List returns a single page of campaigns associated with the given account
// email.
func (s *Service) List(ctx context.Context, email string, opts ...ListOption) (*ListResponse, error) {
	q := instantly.NewQuery()
	for _, opt := range opts {
		if opt != nil {
			opt(q)
		}
	}

	out := &ListResponse{}
	if err := s.client.Get(ctx, q.Path(basePath+"/"+url.PathEscape(email)), out); err != nil {
		return nil, err
	}

	return out, nil
}

// ListIter walks every page of campaigns associated with the given account
// email, yielding each with a nil error, or a nil *Mapping with the first error.
func (s *Service) ListIter(
	ctx context.Context, email string, opts ...ListOption,
) iter.Seq2[*Mapping, error] {
	return instantly.Iterate(ctx, func(ctx context.Context, cursor string) ([]Mapping, string, error) {
		pageOpts := opts
		if cursor != "" {
			pageOpts = append(append([]ListOption(nil), opts...), WithStartingAfter(cursor))
		}

		page, err := s.List(ctx, email, pageOpts...)
		if err != nil {
			return nil, "", err
		}

		return page.Items, page.NextStartingAfter, nil
	})
}
