package backgroundjob

import (
	"context"
	"encoding/json"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Background Job API.
const basePath = "/api/v2/background-jobs"

// Service provides access to the Instantly.ai V2 Background Job API.
type Service struct {
	client *instantly.Client
}

// New builds a Background Job API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Type is the kind of work a background job performs.
type Type string

// The kinds of work a background job can perform.
const (
	// TypeMoveLeads moves leads between lists or campaigns.
	TypeMoveLeads Type = "move-leads"

	// TypeImportLeads imports leads.
	TypeImportLeads Type = "import-leads"

	// TypeExportLeads exports leads.
	TypeExportLeads Type = "export-leads"

	// TypeUpdateWarmupAccounts updates warmup accounts.
	TypeUpdateWarmupAccounts Type = "update-warmup-accounts"

	// TypeRenameVariable renames a variable across leads.
	TypeRenameVariable Type = "rename-variable"

	// TypeBroadcastAIGenerate generates an AI broadcast.
	TypeBroadcastAIGenerate Type = "broadcast-ai-generate"

	// TypeBroadcastWebsiteScrape analyzes a website for a broadcast.
	TypeBroadcastWebsiteScrape Type = "broadcast-website-scrape"

	// TypeImportSubscribersFromCRM imports subscribers from a CRM.
	TypeImportSubscribersFromCRM Type = "import-subscribers-from-crm"

	// TypeResyncSubscriberCRMTags re-syncs subscriber CRM tags.
	TypeResyncSubscriberCRMTags Type = "resync-subscriber-crm-tags"
)

// Status is the state a background job is in.
type Status string

// The states a background job can be in.
const (
	// StatusPending means the job is waiting in the queue to be processed.
	StatusPending Status = "pending"

	// StatusInProgress means the job is being processed.
	StatusInProgress Status = "in-progress"

	// StatusSuccess means the job has been successfully processed.
	StatusSuccess Status = "success"

	// StatusFailed means the job has failed.
	StatusFailed Status = "failed"

	// StatusDraining means the job is replaying deferred live events.
	StatusDraining Status = "draining"

	// StatusPaused means the job is paused, for example waiting for quota or auth.
	StatusPaused Status = "paused"

	// StatusCancelled means the job was cancelled by the user.
	StatusCancelled Status = "cancelled"
)

// EntityType is the kind of entity a background job is related to.
type EntityType string

// The kinds of entity a background job can be related to.
const (
	// EntityTypeList is a lead list.
	EntityTypeList EntityType = "list"

	// EntityTypeCampaign is a campaign.
	EntityTypeCampaign EntityType = "campaign"

	// EntityTypeWorkspace is a workspace.
	EntityTypeWorkspace EntityType = "workspace"

	// EntityTypeBroadcast is an email-marketing broadcast.
	EntityTypeBroadcast EntityType = "broadcast"

	// EntityTypeSubscriberGroupSync is a subscriber-group sync.
	EntityTypeSubscriberGroupSync EntityType = "subscriber-group-sync"

	// EntityTypeSubscriberGroup is a subscriber group.
	EntityTypeSubscriberGroup EntityType = "subscriber-group"
)

// SortColumn is the column a background job list can be sorted by.
type SortColumn string

// The columns a background job list can be sorted by.
const (
	// SortColumnCreatedAt sorts by creation time.
	SortColumnCreatedAt SortColumn = "created_at"

	// SortColumnUpdatedAt sorts by last-update time.
	SortColumnUpdatedAt SortColumn = "updated_at"
)

// Job is a single background job returned by the Instantly.ai V2 API.
//
// Fields the API declares as nullable are pointers, so an absent value stays
// distinguishable from a zero value: a nil EntityID means the API reported
// nothing, which is not the same as an empty id. CreatedAt and UpdatedAt are
// kept as raw strings because the API does not document their exact format.
type Job struct {
	// ID is the unique identifier of the background job.
	ID string `json:"id"`

	// WorkspaceID identifies the workspace the job belongs to.
	WorkspaceID string `json:"workspace_id"`

	// Type is the kind of work the job performs.
	Type Type `json:"type"`

	// Progress is the completion percentage of the job, from 0 to 100.
	Progress float64 `json:"progress"`

	// Status is the state the job is in.
	Status Status `json:"status"`

	// CreatedAt is when the job was created.
	CreatedAt string `json:"created_at"`

	// UpdatedAt is when the job was last updated.
	UpdatedAt string `json:"updated_at"`

	// EntityType is the kind of entity the job is related to, when applicable.
	EntityType EntityType `json:"entity_type,omitempty"`

	// EntityID identifies the entity the job is related to, when applicable.
	EntityID *string `json:"entity_id,omitempty"`

	// UserID identifies the user who triggered the job, when known.
	UserID *string `json:"user_id,omitempty"`

	// Data carries any additional information about the job, which the API models
	// as a free-form object, so it is preserved verbatim.
	Data json.RawMessage `json:"data,omitempty"`
}

// ListResponse is a single page of background jobs.
//
// It aliases instantly.Page[Job], the cursor-paginated envelope every resource
// shares, so the generic pagination helpers accept List directly.
type ListResponse = instantly.Page[Job]

// List returns a single page of background jobs filtered by the supplied
// options.
//
// Pagination is cursor based: pass the returned NextStartingAfter back with
// WithStartingAfter to fetch the following page.
func (s *Service) List(ctx context.Context, opts ...ListOption) (*ListResponse, error) {
	return instantly.GetResult[ListResponse](ctx, s.client, instantly.ApplyOptions(opts...).Path(basePath))
}

// Get returns a single background job by its unique identifier.
//
// Pass WithDataFields to select which fields of the job's data payload are
// returned; without it the endpoint returns the payload as it sees fit.
func (s *Service) Get(ctx context.Context, id string, opts ...GetOption) (*Job, error) {
	path := instantly.ApplyOptions(opts...).Path(instantly.JoinPath(basePath, id))

	return instantly.GetResult[Job](ctx, s.client, path)
}
