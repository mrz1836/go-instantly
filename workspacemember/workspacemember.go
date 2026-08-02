package workspacemember

import (
	"context"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Workspace Member API.
const basePath = "/api/v2/workspace-members"

// Service provides access to the Instantly.ai V2 Workspace Member API.
type Service struct {
	client *instantly.Client
}

// New builds a Workspace Member API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Role is the access level a workspace member has.
type Role string

// The roles a workspace member can hold.
const (
	// RoleOwner has full access and workspace management.
	RoleOwner Role = "owner"

	// RoleAdmin has full access except workspace deletion.
	RoleAdmin Role = "admin"

	// RoleEditor can edit but not manage workspace settings.
	RoleEditor Role = "editor"

	// RoleView has read-only access.
	RoleView Role = "view"

	// RoleClient is the whitelabel agency-view role, not assignable through the
	// API.
	RoleClient Role = "client"
)

// Permission is a single granular permission a workspace member can hold.
type Permission string

// The permissions the API documents.
const (
	PermissionDashboardView            Permission = "dashboard.view"
	PermissionCampaignsView            Permission = "campaigns.view"
	PermissionCampaignsCreate          Permission = "campaigns.create"
	PermissionCampaignsEdit            Permission = "campaigns.edit"
	PermissionCampaignsDelete          Permission = "campaigns.delete"
	PermissionOrganizationManage       Permission = "organization.manage"
	PermissionOrganizationIntegrations Permission = "organization.integrations"
	PermissionOrganizationBilling      Permission = "organization.billing"
	PermissionOrganizationUsersManage  Permission = "organization.users.manage"
	PermissionLeadFinderView           Permission = "leadFinder.view"
	PermissionCustomLeadLabelsCreate   Permission = "customLeadLabels.create"
	PermissionCustomLeadLabelsEdit     Permission = "customLeadLabels.edit"
	PermissionCustomLeadLabelsDelete   Permission = "customLeadLabels.delete"
	PermissionUniboxAll                Permission = "unibox.all"
	PermissionAnalyticsView            Permission = "analytics.view"
	PermissionAgencyManage             Permission = "agency.manage"
	PermissionAccountsView             Permission = "accounts.view"
	PermissionAccountsManage           Permission = "accounts.manage"
	PermissionLeadManagementView       Permission = "leadManagement.view"
	PermissionLeadsMove                Permission = "leads.move"
	PermissionCRMView                  Permission = "crm.view"
	PermissionWebsiteVisitorsView      Permission = "websiteVisitors.view"
	PermissionBlocklistManage          Permission = "blocklist.manage"
	PermissionPreferencesManage        Permission = "preferences.manage"
	PermissionInboxPlacementView       Permission = "inboxPlacement.view"
	PermissionAIAgentsManage           Permission = "aiAgents.manage"
	PermissionGroupMembersInvite       Permission = "workspaceGroupMembers.invite"
	PermissionGroupMembersRemove       Permission = "workspaceGroupMembers.remove"
	PermissionGroupMembersLeave        Permission = "workspaceGroupMembers.leave"
)

// Name is the first and last name of a workspace member.
type Name struct {
	// First is the member's first name.
	First string `json:"first"`

	// Last is the member's last name.
	Last string `json:"last"`
}

// Member is a single workspace member returned by the Instantly.ai V2 API.
//
// Fields the API declares as nullable are pointers so an absent value stays
// distinguishable from a zero value.
type Member struct {
	// ID is the unique identifier of the membership.
	ID string `json:"id"`

	// Email is the address the member was invited with.
	Email string `json:"email"`

	// UserID identifies the user behind the membership.
	UserID string `json:"user_id"`

	// WorkspaceID identifies the workspace the member belongs to.
	WorkspaceID string `json:"workspace_id"`

	// Role is the access level the member holds.
	Role Role `json:"role"`

	// Accepted reports whether the member has accepted the invitation.
	Accepted bool `json:"accepted"`

	// TimestampCreated is when the membership was created.
	TimestampCreated string `json:"timestamp_created"`

	// Name is the first and last name of the member.
	Name Name `json:"name"`

	// Nickname is the display nickname of the member.
	Nickname *string `json:"nickname,omitempty"`

	// UserEmail is the member's account email, when it differs from Email.
	UserEmail *string `json:"user_email,omitempty"`

	// IssuerID identifies the user that issued the invitation.
	IssuerID *string `json:"issuer_id,omitempty"`

	// Permissions are the granular permissions the member holds.
	Permissions []Permission `json:"permissions,omitempty"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded member re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (m *Member) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, m.TimestampCreated)
}

// ListResponse is a single page of workspace members.
//
// It aliases instantly.Page[Member], the cursor-paginated envelope every
// resource shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[Member]

// CreateRequest is the body of an invite-workspace-member request.
type CreateRequest struct {
	// Email is the address to invite. Required.
	Email string `json:"email"`

	// Role is the access level to grant. Required.
	Role Role `json:"role"`

	// Nickname is the display nickname to set for the member.
	Nickname *string `json:"nickname,omitempty"`

	// UserEmail is the member's account email, when it differs from Email.
	UserEmail *string `json:"user_email,omitempty"`

	// Permissions are the granular permissions to grant.
	Permissions []Permission `json:"permissions,omitempty"`
}

// UpdateRequest is the body of a patch-workspace-member request. No field is
// required; an omitted field leaves the current value unchanged.
type UpdateRequest struct {
	// Role is the access level to set.
	Role Role `json:"role,omitempty"`

	// Nickname is the display nickname to set.
	Nickname *string `json:"nickname,omitempty"`
}

// Create invites a new workspace member and returns it.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Member, error) {
	return instantly.PostResult[Member](ctx, s.client, basePath, req)
}

// List returns a single page of workspace members filtered by the supplied
// options.
//
// Pagination is cursor based: pass the returned NextStartingAfter back with
// WithStartingAfter to fetch the following page.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}

// Get returns a single workspace member by its unique identifier.
func (s *Service) Get(ctx context.Context, id string) (*Member, error) {
	return instantly.GetResult[Member](ctx, s.client, instantly.JoinPath(basePath, id))
}

// Update patches a workspace member and returns its updated state.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Member, error) {
	return instantly.PatchResult[Member](ctx, s.client, instantly.JoinPath(basePath, id), req)
}

// Delete removes a workspace member and returns the member that was removed.
func (s *Service) Delete(ctx context.Context, id string) (*Member, error) {
	return instantly.DeleteResult[Member](ctx, s.client, instantly.JoinPath(basePath, id))
}
