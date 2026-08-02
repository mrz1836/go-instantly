package workspace

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Workspace API. Every endpoint operates on the
// single workspace the API key authenticates against.
const basePath = "/api/v2/workspaces/current"

// Service provides access to the Instantly.ai V2 Workspace API.
type Service struct {
	client *instantly.Client
}

// New builds a Workspace API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Workspace is the current workspace returned by the Instantly.ai V2 API.
//
// Fields the API declares as nullable are pointers so an absent value stays
// distinguishable from a zero value.
type Workspace struct {
	// ID is the unique identifier of the workspace.
	ID string `json:"id"`

	// Name is the display name of the workspace.
	Name string `json:"name"`

	// Owner is the user ID that owns the workspace.
	Owner string `json:"owner"`

	// TimestampCreated is when the workspace was created.
	TimestampCreated string `json:"timestamp_created"`

	// TimestampUpdated is when the workspace was last updated.
	TimestampUpdated string `json:"timestamp_updated"`

	// OrgLogoURL is the URL of the workspace's logo.
	OrgLogoURL *string `json:"org_logo_url,omitempty"`

	// OrgClientDomain is the workspace's whitelabel client domain.
	OrgClientDomain *string `json:"org_client_domain,omitempty"`

	// AddUnsubToBlock reports whether unsubscribes are added to the block list.
	AddUnsubToBlock *bool `json:"add_unsub_to_block,omitempty"`

	// DefaultOpportunityValue is the default opportunity value for the workspace.
	DefaultOpportunityValue *float64 `json:"default_opportunity_value,omitempty"`

	// ScheduledForRemovalAt is when the workspace is scheduled to be removed.
	ScheduledForRemovalAt *string `json:"scheduled_for_removal_at,omitempty"`

	// PlanID is the identifier of the workspace's main plan.
	PlanID *string `json:"plan_id,omitempty"`

	// PlanIDBundle is the identifier of the workspace's bundle plan.
	PlanIDBundle *string `json:"plan_id_bundle,omitempty"`

	// PlanIDCRM is the identifier of the workspace's CRM plan.
	PlanIDCRM *string `json:"plan_id_crm,omitempty"`

	// PlanIDInboxPlacement is the identifier of the workspace's inbox placement
	// plan.
	PlanIDInboxPlacement *string `json:"plan_id_inbox_placement,omitempty"`

	// PlanIDLeadfinder is the identifier of the workspace's lead finder plan.
	PlanIDLeadfinder *string `json:"plan_id_leadfinder,omitempty"`

	// PlanIDWebsiteVisitor is the identifier of the workspace's website visitor
	// plan.
	PlanIDWebsiteVisitor *string `json:"plan_id_website_visitor,omitempty"`

	// PlanIDVerification carries the raw verification plan detail, which the API
	// models as a nested object, so it is preserved verbatim.
	PlanIDVerification json.RawMessage `json:"plan_id_verification,omitempty"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded workspace re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (w *Workspace) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, w.TimestampCreated)
}

// DomainVerification is one DNS record needed to verify a whitelabel domain.
type DomainVerification struct {
	// Domain is the domain the record is for.
	Domain string `json:"domain"`

	// Type is the DNS record type.
	Type string `json:"type"`

	// Value is the value the record must hold.
	Value string `json:"value"`

	// Reason is why the record is required.
	Reason string `json:"reason"`
}

// DomainInfo is the verification status of the workspace's whitelabel domain.
type DomainInfo struct {
	// Name is the whitelabel domain name.
	Name string `json:"name"`

	// Verified reports whether the domain has been verified.
	Verified bool `json:"verified"`

	// Verification lists the DNS records needed to verify the domain.
	Verification []DomainVerification `json:"verification,omitempty"`
}

// UpdateRequest is the body of a patch-workspace request. No field is required;
// an omitted field leaves the current value unchanged.
type UpdateRequest struct {
	// Name is the display name of the workspace.
	Name string `json:"name,omitempty"`

	// OrgLogoURL is the URL of the workspace's logo.
	OrgLogoURL *string `json:"org_logo_url,omitempty"`
}

// SetDomainRequest is the body of a set-agency-domain request.
type SetDomainRequest struct {
	// Domain is the whitelabel agency domain to set. Required.
	Domain string `json:"domain"`
}

// ChangeOwnerRequest is the body of a change-owner request.
type ChangeOwnerRequest struct {
	// Email is the address of the new owner. Required.
	Email string `json:"email"`

	// Sec is the security confirmation token for the change. Required.
	Sec string `json:"sec"`
}

// Get returns the current workspace.
func (s *Service) Get(ctx context.Context) (*Workspace, error) {
	return instantly.GetResult[Workspace](ctx, s.client, basePath)
}

// Update patches the current workspace and returns its updated state.
func (s *Service) Update(ctx context.Context, req UpdateRequest) (*Workspace, error) {
	return instantly.PatchResult[Workspace](ctx, s.client, basePath, req)
}

// ScheduleRemoval schedules the current workspace for removal and returns it.
func (s *Service) ScheduleRemoval(ctx context.Context) (*Workspace, error) {
	return instantly.PostResult[Workspace](ctx, s.client, basePath+"/schedule-for-removal", nil)
}

// CancelRemoval cancels a scheduled removal of the current workspace and returns
// it.
func (s *Service) CancelRemoval(ctx context.Context) (*Workspace, error) {
	return instantly.DeleteResult[Workspace](ctx, s.client, basePath+"/schedule-for-removal")
}

// SetAgencyDomain sets the workspace's whitelabel agency domain and returns the
// workspace.
func (s *Service) SetAgencyDomain(ctx context.Context, req SetDomainRequest) (*Workspace, error) {
	return instantly.PostResult[Workspace](ctx, s.client, basePath+"/whitelabel-domain", req)
}

// DomainInfo returns the verification status of the workspace's whitelabel
// domain.
func (s *Service) DomainInfo(ctx context.Context) (*DomainInfo, error) {
	return instantly.GetResult[DomainInfo](ctx, s.client, basePath+"/whitelabel-domain")
}

// DeleteAgencyDomain removes the workspace's whitelabel agency domain and returns
// the workspace.
func (s *Service) DeleteAgencyDomain(ctx context.Context) (*Workspace, error) {
	return instantly.DeleteResult[Workspace](ctx, s.client, basePath+"/whitelabel-domain")
}

// ChangeOwner changes the owner of the current workspace and returns it.
func (s *Service) ChangeOwner(ctx context.Context, req ChangeOwnerRequest) (*Workspace, error) {
	return instantly.PostResult[Workspace](ctx, s.client, basePath+"/change-owner", req)
}
