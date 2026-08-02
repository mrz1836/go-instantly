package workspacebilling

import (
	"context"
	"encoding/json"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Workspace Billing API.
const basePath = "/api/v2/workspace-billing"

// Service provides access to the Instantly.ai V2 Workspace Billing API.
type Service struct {
	client *instantly.Client
}

// New builds a Workspace Billing API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// PlanDetails is the current workspace's plan details.
//
// The addons and subscriptions payloads are provider-shaped and not documented
// as a fixed schema, so they are preserved verbatim as json.RawMessage.
type PlanDetails struct {
	// OrganizationID identifies the organization the plan belongs to.
	OrganizationID string `json:"organization_id,omitempty"`

	// OrganizationName is the display name of the organization.
	OrganizationName string `json:"organization_name,omitempty"`

	// Addons carries the raw addon detail, preserved verbatim.
	Addons json.RawMessage `json:"addons,omitempty"`

	// Subscriptions carries the raw subscription detail, preserved verbatim.
	Subscriptions json.RawMessage `json:"subscriptions,omitempty"`
}

// SubscriptionDetails is the current workspace's subscription details.
type SubscriptionDetails struct {
	// AllSubsCancelled reports whether every subscription has been cancelled.
	AllSubsCancelled bool `json:"all_subs_cancelled"`

	// Subscriptions are the individual subscriptions, each preserved verbatim
	// because the API does not document them as a fixed schema.
	Subscriptions []json.RawMessage `json:"subscriptions,omitempty"`
}

// PlanDetails returns the current workspace's plan details.
func (s *Service) PlanDetails(ctx context.Context) (*PlanDetails, error) {
	return instantly.GetResult[PlanDetails](ctx, s.client, basePath+"/plan-details")
}

// SubscriptionDetails returns the current workspace's subscription details.
func (s *Service) SubscriptionDetails(ctx context.Context) (*SubscriptionDetails, error) {
	return instantly.GetResult[SubscriptionDetails](ctx, s.client, basePath+"/subscription-details")
}
