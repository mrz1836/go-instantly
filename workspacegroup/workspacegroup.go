package workspacegroup

import (
	"context"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Workspace Group Member API.
const basePath = "/api/v2/workspace-group-members"

// Service provides access to the Instantly.ai V2 Workspace Group Member API.
type Service struct {
	client *instantly.Client
}

// New builds a Workspace Group Member API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Status is whether a sub workspace has accepted its group invitation.
type Status string

// The statuses a workspace group member can be in.
const (
	// StatusPending means the member has been invited but has not responded.
	StatusPending Status = "pending"

	// StatusAccepted means the member accepted the invitation.
	StatusAccepted Status = "accepted"

	// StatusRejected means the member rejected the invitation.
	StatusRejected Status = "rejected"
)

// Member is a single workspace group member returned by the Instantly.ai V2 API.
//
// Fields the API declares as nullable are pointers so an absent value stays
// distinguishable from a zero value.
type Member struct {
	// ID is the unique identifier of the group membership.
	ID string `json:"id"`

	// AdminWorkspaceID identifies the admin workspace of the group.
	AdminWorkspaceID string `json:"admin_workspace_id"`

	// SubWorkspaceID identifies the sub workspace that is a member of the group.
	SubWorkspaceID string `json:"sub_workspace_id"`

	// Status is whether the sub workspace has accepted its invitation.
	Status Status `json:"status"`

	// TimestampCreated is when the group membership was created.
	TimestampCreated string `json:"timestamp_created"`

	// TimestampUpdated is when the group membership was last updated.
	TimestampUpdated string `json:"timestamp_updated"`

	// AdminWorkspaceName is the display name of the admin workspace.
	AdminWorkspaceName *string `json:"admin_workspace_name,omitempty"`

	// SubWorkspaceName is the display name of the sub workspace.
	SubWorkspaceName *string `json:"sub_workspace_name,omitempty"`
}

// ParsedTimestampCreated parses TimestampCreated as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded member re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (m *Member) ParsedTimestampCreated() (time.Time, error) {
	return time.Parse(time.RFC3339, m.TimestampCreated)
}

// ListResponse is a single page of workspace group members.
//
// It aliases instantly.Page[Member], the cursor-paginated envelope every
// resource shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[Member]

// Admin is the admin workspace the current workspace belongs to, if any.
type Admin struct {
	// HasAdminWorkspace reports whether the current workspace has an admin
	// workspace.
	HasAdminWorkspace bool `json:"has_admin_workspace"`

	// WorkspaceName is the name of the admin workspace.
	WorkspaceName string `json:"workspace_name"`

	// WorkspaceGroupMemberID identifies the current workspace's group membership.
	WorkspaceGroupMemberID string `json:"workspace_group_member_id,omitempty"`
}

// CreateRequest is the body of an invite-workspace-group-member request.
type CreateRequest struct {
	// SubWorkspaceID identifies the sub workspace to invite into the group.
	// Required.
	SubWorkspaceID string `json:"sub_workspace_id"`
}

// Create invites a sub workspace into the group and returns the membership.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Member, error) {
	return instantly.PostResult[Member](ctx, s.client, basePath, req)
}

// List returns a single page of workspace group members.
//
// Pagination is cursor based: pass the returned NextStartingAfter back with
// WithStartingAfter to fetch the following page.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}

// Get returns a single workspace group member by its unique identifier.
func (s *Service) Get(ctx context.Context, id string) (*Member, error) {
	return instantly.GetResult[Member](ctx, s.client, instantly.JoinPath(basePath, id))
}

// Delete removes a workspace group member and returns the member that was
// removed.
func (s *Service) Delete(ctx context.Context, id string) (*Member, error) {
	return instantly.DeleteResult[Member](ctx, s.client, instantly.JoinPath(basePath, id))
}

// Admin returns the admin workspace the current workspace belongs to.
//
// It relies on the router matching this literal path ahead of the
// /workspace-group-members/{id} route.
func (s *Service) Admin(ctx context.Context) (*Admin, error) {
	return instantly.GetResult[Admin](ctx, s.client, basePath+"/admin")
}
