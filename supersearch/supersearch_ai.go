package supersearch

import (
	"context"
	"encoding/json"

	"github.com/mrz1836/go-instantly"
)

// AIStatus is the processing status of an AI enrichment job.
type AIStatus int64

// The statuses an AI enrichment job can be in.
const (
	// AIStatusPending means the job is pending processing.
	AIStatusPending AIStatus = 1

	// AIStatusProcessing means the job is currently being processed.
	AIStatusProcessing AIStatus = 2

	// AIStatusCompleted means the job completed successfully.
	AIStatusCompleted AIStatus = 3

	// AIStatusFailed means the job failed.
	AIStatusFailed AIStatus = 4
)

// ModelVersion is the AI model an AI enrichment runs against.
//
// It is a named string so a call site is self-documenting and type-checked,
// while remaining forward compatible: a model the API adds later can still be
// passed as ModelVersion("new-model") without a package change.
type ModelVersion string

// The AI models the API documents.
const (
	ModelVersion35                   ModelVersion = "3.5"
	ModelVersion40                   ModelVersion = "4.0"
	ModelVersionGPT4o                ModelVersion = "gpt-4o"
	ModelVersionO3                   ModelVersion = "o3"
	ModelVersionGPT41                ModelVersion = "gpt-4.1"
	ModelVersionGPT41Mini            ModelVersion = "gpt-4.1-mini"
	ModelVersionGPT5Mini             ModelVersion = "gpt-5-mini"
	ModelVersionGPT5Nano             ModelVersion = "gpt-5-nano"
	ModelVersionGPT5                 ModelVersion = "gpt-5"
	ModelVersionGPT54                ModelVersion = "gpt-5.4"
	ModelVersionClaude45Sonnet       ModelVersion = "claude-4.5-sonnet"
	ModelVersionClaude46Sonnet       ModelVersion = "claude-4.6-sonnet"
	ModelVersionR1                   ModelVersion = "r1"
	ModelVersionGrok43               ModelVersion = "grok-4.3"
	ModelVersionGemini30Flash        ModelVersion = "gemini-3.0-flash"
	ModelVersionGemini35Flash        ModelVersion = "gemini-3.5-flash"
	ModelVersionSonar                ModelVersion = "sonar"
	ModelVersionSonarPro             ModelVersion = "sonar-pro"
	ModelVersionLightspeedResearch   ModelVersion = "instantly-ai-lightspeed-agent-for-web-research"
	ModelVersionLightspeedEmailWrite ModelVersion = "instantly-ai-lightspeed-agent-for-email-generation"
)

// AIEnrichment is an AI enrichment job.
//
// It is the shape shared by the create-AI response and the in-progress list, so
// fields a given endpoint does not return stay nil or zero.
type AIEnrichment struct {
	// ID is the unique identifier of the AI enrichment job.
	ID string `json:"id"`

	// OrganizationID identifies the organization the job belongs to.
	OrganizationID string `json:"organization_id,omitempty"`

	// ResourceID identifies the campaign or list the job targets.
	ResourceID string `json:"resource_id"`

	// ResourceType is whether the job targets a campaign or a list.
	ResourceType ResourceType `json:"resource_type"`

	// OutputColumn is the column the AI output is written to.
	OutputColumn string `json:"output_column"`

	// Status is the processing status of the job.
	Status AIStatus `json:"status"`

	// ModelVersion is the AI model the job runs against.
	ModelVersion ModelVersion `json:"model_version,omitempty"`

	// InputColumns are the columns fed into the AI prompt.
	InputColumns []string `json:"input_columns,omitempty"`

	// Limit is the maximum number of leads to enrich.
	Limit *float64 `json:"limit,omitempty"`

	// Overwrite reports whether existing values are overwritten.
	Overwrite *bool `json:"overwrite,omitempty"`

	// AutoUpdate reports whether the job updates automatically.
	AutoUpdate *bool `json:"auto_update,omitempty"`

	// TemplateID identifies the prompt template the job uses.
	TemplateID *string `json:"template_id,omitempty"`
}

// AIRequest is the body of a create-AI-enrichment request.
type AIRequest struct {
	// ResourceID identifies the campaign or list to enrich. Required.
	ResourceID string `json:"resource_id"`

	// OutputColumn is the column the AI output is written to. Required.
	OutputColumn string `json:"output_column"`

	// ResourceType is whether the enrichment targets a campaign or a list.
	// Required.
	ResourceType ResourceType `json:"resource_type"`

	// ModelVersion is the AI model to run against. Required.
	ModelVersion ModelVersion `json:"model_version"`

	// Prompt is the AI prompt to run.
	Prompt string `json:"prompt,omitempty"`

	// TemplateID identifies a prompt template to use in place of Prompt.
	TemplateID string `json:"template_id,omitempty"`

	// InputColumns are the columns fed into the AI prompt.
	InputColumns []string `json:"input_columns,omitempty"`

	// Limit is the maximum number of leads to enrich.
	Limit *float64 `json:"limit,omitempty"`

	// Status sets the initial job status.
	Status *AIStatus `json:"status,omitempty"`

	// AutoUpdate sets whether the job updates automatically.
	AutoUpdate *bool `json:"auto_update,omitempty"`

	// Overwrite overwrites existing values when true.
	Overwrite *bool `json:"overwrite,omitempty"`

	// SkipLeadsWithoutEmail skips leads that have no email when true.
	SkipLeadsWithoutEmail *bool `json:"skip_leads_without_email,omitempty"`

	// UseInstantlyAccount runs the enrichment on the Instantly account when true.
	UseInstantlyAccount *bool `json:"use_instantly_account,omitempty"`

	// Filters carries the raw enrichment filters, sent verbatim.
	Filters json.RawMessage `json:"filters,omitempty"`
}

// CreateAI creates an AI enrichment and returns the job.
func (s *Service) CreateAI(ctx context.Context, req AIRequest) (*AIEnrichment, error) {
	return instantly.PostResult[AIEnrichment](ctx, s.client, basePath+"/ai", req)
}

// AIInProgress returns the in-progress AI enrichment jobs for a resource.
func (s *Service) AIInProgress(ctx context.Context, resourceID string) ([]AIEnrichment, error) {
	path := instantly.JoinPath(basePath, "ai", resourceID, "in-progress")

	var out []AIEnrichment
	if err := s.client.Get(ctx, path, &out); err != nil {
		return nil, err
	}

	return out, nil
}
