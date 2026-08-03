package apikey

// Scope is a permission an API key can be granted, spelled `<resource>:<action>`
// on the wire.
//
// The named constants below enumerate every scope the API documents, grouped by
// resource. Scope is a defined string type rather than a bare enum, so a scope
// the API adds in the future still works without a new release: pass it as
// Scope("new_resource:read").
type Scope string

// Scopes for every resource.
const (
	ScopeAllAll    Scope = "all:all"
	ScopeAllCreate Scope = "all:create"
	ScopeAllRead   Scope = "all:read"
	ScopeAllUpdate Scope = "all:update"
	ScopeAllDelete Scope = "all:delete"
)

// Scopes for AI agents.
const (
	ScopeAIAgentsAll    Scope = "ai_agents:all"
	ScopeAIAgentsCreate Scope = "ai_agents:create"
	ScopeAIAgentsRead   Scope = "ai_agents:read"
	ScopeAIAgentsUpdate Scope = "ai_agents:update"
	ScopeAIAgentsDelete Scope = "ai_agents:delete"
)

// Scopes for API keys.
const (
	ScopeAPIKeysAll    Scope = "api_keys:all"
	ScopeAPIKeysCreate Scope = "api_keys:create"
	ScopeAPIKeysRead   Scope = "api_keys:read"
	ScopeAPIKeysUpdate Scope = "api_keys:update"
	ScopeAPIKeysDelete Scope = "api_keys:delete"
)

// Scopes for audit logs.
const (
	ScopeAuditLogsAll    Scope = "audit_logs:all"
	ScopeAuditLogsCreate Scope = "audit_logs:create"
	ScopeAuditLogsRead   Scope = "audit_logs:read"
	ScopeAuditLogsUpdate Scope = "audit_logs:update"
	ScopeAuditLogsDelete Scope = "audit_logs:delete"
)

// Scopes for custom prompt templates.
const (
	ScopeCustomPromptTemplatesAll    Scope = "custom_prompt_templates:all"
	ScopeCustomPromptTemplatesCreate Scope = "custom_prompt_templates:create"
	ScopeCustomPromptTemplatesRead   Scope = "custom_prompt_templates:read"
	ScopeCustomPromptTemplatesUpdate Scope = "custom_prompt_templates:update"
	ScopeCustomPromptTemplatesDelete Scope = "custom_prompt_templates:delete"
)

// Scopes for account-campaign mappings.
const (
	ScopeAccountCampaignMappingsAll    Scope = "account_campaign_mappings:all"
	ScopeAccountCampaignMappingsCreate Scope = "account_campaign_mappings:create"
	ScopeAccountCampaignMappingsRead   Scope = "account_campaign_mappings:read"
	ScopeAccountCampaignMappingsUpdate Scope = "account_campaign_mappings:update"
	ScopeAccountCampaignMappingsDelete Scope = "account_campaign_mappings:delete"
)

// Scopes for campaigns.
const (
	ScopeCampaignsAll    Scope = "campaigns:all"
	ScopeCampaignsCreate Scope = "campaigns:create"
	ScopeCampaignsRead   Scope = "campaigns:read"
	ScopeCampaignsUpdate Scope = "campaigns:update"
	ScopeCampaignsDelete Scope = "campaigns:delete"
)

// Scopes for inbox placement tests.
const (
	ScopeInboxPlacementTestsAll    Scope = "inbox_placement_tests:all"
	ScopeInboxPlacementTestsCreate Scope = "inbox_placement_tests:create"
	ScopeInboxPlacementTestsRead   Scope = "inbox_placement_tests:read"
	ScopeInboxPlacementTestsUpdate Scope = "inbox_placement_tests:update"
	ScopeInboxPlacementTestsDelete Scope = "inbox_placement_tests:delete"
)

// Scopes for inbox placement analytics.
const (
	ScopeInboxPlacementAnalyticsAll    Scope = "inbox_placement_analytics:all"
	ScopeInboxPlacementAnalyticsCreate Scope = "inbox_placement_analytics:create"
	ScopeInboxPlacementAnalyticsRead   Scope = "inbox_placement_analytics:read"
	ScopeInboxPlacementAnalyticsUpdate Scope = "inbox_placement_analytics:update"
	ScopeInboxPlacementAnalyticsDelete Scope = "inbox_placement_analytics:delete"
)

// Scopes for inbox placement reports.
const (
	ScopeInboxPlacementReportsAll    Scope = "inbox_placement_reports:all"
	ScopeInboxPlacementReportsCreate Scope = "inbox_placement_reports:create"
	ScopeInboxPlacementReportsRead   Scope = "inbox_placement_reports:read"
	ScopeInboxPlacementReportsUpdate Scope = "inbox_placement_reports:update"
	ScopeInboxPlacementReportsDelete Scope = "inbox_placement_reports:delete"
)

// Scopes for lead lists.
const (
	ScopeLeadListsAll    Scope = "lead_lists:all"
	ScopeLeadListsCreate Scope = "lead_lists:create"
	ScopeLeadListsRead   Scope = "lead_lists:read"
	ScopeLeadListsUpdate Scope = "lead_lists:update"
	ScopeLeadListsDelete Scope = "lead_lists:delete"
)

// Scopes for leads.
const (
	ScopeLeadsAll    Scope = "leads:all"
	ScopeLeadsCreate Scope = "leads:create"
	ScopeLeadsRead   Scope = "leads:read"
	ScopeLeadsUpdate Scope = "leads:update"
	ScopeLeadsDelete Scope = "leads:delete"
)

// Scopes for background jobs.
const (
	ScopeBackgroundJobsAll    Scope = "background-jobs:all"
	ScopeBackgroundJobsCreate Scope = "background-jobs:create"
	ScopeBackgroundJobsRead   Scope = "background-jobs:read"
	ScopeBackgroundJobsUpdate Scope = "background-jobs:update"
	ScopeBackgroundJobsDelete Scope = "background-jobs:delete"
)

// Scopes for custom tags.
const (
	ScopeCustomTagsAll    Scope = "custom_tags:all"
	ScopeCustomTagsCreate Scope = "custom_tags:create"
	ScopeCustomTagsRead   Scope = "custom_tags:read"
	ScopeCustomTagsUpdate Scope = "custom_tags:update"
	ScopeCustomTagsDelete Scope = "custom_tags:delete"
)

// Scopes for custom tag mappings.
const (
	ScopeCustomTagMappingsAll    Scope = "custom_tag_mappings:all"
	ScopeCustomTagMappingsCreate Scope = "custom_tag_mappings:create"
	ScopeCustomTagMappingsRead   Scope = "custom_tag_mappings:read"
	ScopeCustomTagMappingsUpdate Scope = "custom_tag_mappings:update"
	ScopeCustomTagMappingsDelete Scope = "custom_tag_mappings:delete"
)

// Scopes for CRM actions.
const (
	ScopeCRMActionsAll    Scope = "crm_actions:all"
	ScopeCRMActionsCreate Scope = "crm_actions:create"
	ScopeCRMActionsRead   Scope = "crm_actions:read"
	ScopeCRMActionsUpdate Scope = "crm_actions:update"
	ScopeCRMActionsDelete Scope = "crm_actions:delete"
)

// Scopes for sending accounts.
const (
	ScopeAccountsAll    Scope = "accounts:all"
	ScopeAccountsCreate Scope = "accounts:create"
	ScopeAccountsRead   Scope = "accounts:read"
	ScopeAccountsUpdate Scope = "accounts:update"
	ScopeAccountsDelete Scope = "accounts:delete"
)

// Scopes for block list entries.
const (
	ScopeBlockListEntriesAll    Scope = "block_list_entries:all"
	ScopeBlockListEntriesCreate Scope = "block_list_entries:create"
	ScopeBlockListEntriesRead   Scope = "block_list_entries:read"
	ScopeBlockListEntriesUpdate Scope = "block_list_entries:update"
	ScopeBlockListEntriesDelete Scope = "block_list_entries:delete"
)

// Scopes for lead labels.
const (
	ScopeLeadLabelsAll    Scope = "lead-labels:all"
	ScopeLeadLabelsCreate Scope = "lead-labels:create"
	ScopeLeadLabelsRead   Scope = "lead-labels:read"
	ScopeLeadLabelsUpdate Scope = "lead-labels:update"
	ScopeLeadLabelsDelete Scope = "lead-labels:delete"
)

// Scopes for email verifications.
const (
	ScopeEmailVerificationsAll    Scope = "email_verifications:all"
	ScopeEmailVerificationsCreate Scope = "email_verifications:create"
	ScopeEmailVerificationsRead   Scope = "email_verifications:read"
)

// Scopes for emails.
const (
	ScopeEmailsAll    Scope = "emails:all"
	ScopeEmailsCreate Scope = "emails:create"
	ScopeEmailsRead   Scope = "emails:read"
	ScopeEmailsUpdate Scope = "emails:update"
	ScopeEmailsDelete Scope = "emails:delete"
)

// Scopes for email templates.
const (
	ScopeEmailTemplatesAll    Scope = "email_templates:all"
	ScopeEmailTemplatesCreate Scope = "email_templates:create"
	ScopeEmailTemplatesRead   Scope = "email_templates:read"
	ScopeEmailTemplatesUpdate Scope = "email_templates:update"
	ScopeEmailTemplatesDelete Scope = "email_templates:delete"
)

// Scopes for workspaces.
const (
	ScopeWorkspacesAll    Scope = "workspaces:all"
	ScopeWorkspacesCreate Scope = "workspaces:create"
	ScopeWorkspacesRead   Scope = "workspaces:read"
	ScopeWorkspacesUpdate Scope = "workspaces:update"
	ScopeWorkspacesDelete Scope = "workspaces:delete"
)

// Scopes for workspace billing.
const (
	ScopeWorkspaceBillingAll    Scope = "workspace_billing:all"
	ScopeWorkspaceBillingCreate Scope = "workspace_billing:create"
	ScopeWorkspaceBillingRead   Scope = "workspace_billing:read"
	ScopeWorkspaceBillingUpdate Scope = "workspace_billing:update"
	ScopeWorkspaceBillingDelete Scope = "workspace_billing:delete"
)

// Scopes for workspace group members.
const (
	ScopeWorkspaceGroupMembersAll    Scope = "workspace_group_members:all"
	ScopeWorkspaceGroupMembersCreate Scope = "workspace_group_members:create"
	ScopeWorkspaceGroupMembersRead   Scope = "workspace_group_members:read"
	ScopeWorkspaceGroupMembersUpdate Scope = "workspace_group_members:update"
	ScopeWorkspaceGroupMembersDelete Scope = "workspace_group_members:delete"
)

// Scopes for workspace members.
const (
	ScopeWorkspaceMembersAll    Scope = "workspace_members:all"
	ScopeWorkspaceMembersCreate Scope = "workspace_members:create"
	ScopeWorkspaceMembersRead   Scope = "workspace_members:read"
	ScopeWorkspaceMembersUpdate Scope = "workspace_members:update"
	ScopeWorkspaceMembersDelete Scope = "workspace_members:delete"
)

// Scopes for campaign subsequences.
const (
	ScopeSubsequencesAll    Scope = "subsequences:all"
	ScopeSubsequencesCreate Scope = "subsequences:create"
	ScopeSubsequencesRead   Scope = "subsequences:read"
	ScopeSubsequencesUpdate Scope = "subsequences:update"
	ScopeSubsequencesDelete Scope = "subsequences:delete"
)

// Scopes for AI SDR.
const (
	ScopeAISDRAll    Scope = "ai_sdr:all"
	ScopeAISDRCreate Scope = "ai_sdr:create"
	ScopeAISDRRead   Scope = "ai_sdr:read"
	ScopeAISDRUpdate Scope = "ai_sdr:update"
	ScopeAISDRDelete Scope = "ai_sdr:delete"
)

// Scopes for AI SDR replies.
const (
	ScopeAISDRRepliesAll    Scope = "ai_sdr_replies:all"
	ScopeAISDRRepliesCreate Scope = "ai_sdr_replies:create"
	ScopeAISDRRepliesRead   Scope = "ai_sdr_replies:read"
	ScopeAISDRRepliesUpdate Scope = "ai_sdr_replies:update"
	ScopeAISDRRepliesDelete Scope = "ai_sdr_replies:delete"
)

// Scopes for AI inbox manager analytics.
const (
	ScopeAIInboxManagerAnalyticsAll    Scope = "ai_inbox_manager_analytics:all"
	ScopeAIInboxManagerAnalyticsCreate Scope = "ai_inbox_manager_analytics:create"
	ScopeAIInboxManagerAnalyticsRead   Scope = "ai_inbox_manager_analytics:read"
	ScopeAIInboxManagerAnalyticsUpdate Scope = "ai_inbox_manager_analytics:update"
	ScopeAIInboxManagerAnalyticsDelete Scope = "ai_inbox_manager_analytics:delete"
)

// Scopes for sales flows.
const (
	ScopeSalesFlowsAll    Scope = "sales_flows:all"
	ScopeSalesFlowsCreate Scope = "sales_flows:create"
	ScopeSalesFlowsRead   Scope = "sales_flows:read"
	ScopeSalesFlowsUpdate Scope = "sales_flows:update"
	ScopeSalesFlowsDelete Scope = "sales_flows:delete"
)

// Scopes for webhooks.
const (
	ScopeWebhooksAll    Scope = "webhooks:all"
	ScopeWebhooksCreate Scope = "webhooks:create"
	ScopeWebhooksRead   Scope = "webhooks:read"
	ScopeWebhooksUpdate Scope = "webhooks:update"
	ScopeWebhooksDelete Scope = "webhooks:delete"
)

// Scopes for webhook events.
const (
	ScopeWebhookEventsAll    Scope = "webhook_events:all"
	ScopeWebhookEventsCreate Scope = "webhook_events:create"
	ScopeWebhookEventsRead   Scope = "webhook_events:read"
	ScopeWebhookEventsUpdate Scope = "webhook_events:update"
	ScopeWebhookEventsDelete Scope = "webhook_events:delete"
)

// Scopes for security tokens.
const (
	ScopeSecurityTokensAll    Scope = "security_tokens:all"
	ScopeSecurityTokensCreate Scope = "security_tokens:create"
	ScopeSecurityTokensRead   Scope = "security_tokens:read"
	ScopeSecurityTokensUpdate Scope = "security_tokens:update"
	ScopeSecurityTokensDelete Scope = "security_tokens:delete"
)

// Scopes for DFY email account orders.
const (
	ScopeDFYEmailAccountOrdersAll    Scope = "dfy_email_account_orders:all"
	ScopeDFYEmailAccountOrdersCreate Scope = "dfy_email_account_orders:create"
	ScopeDFYEmailAccountOrdersRead   Scope = "dfy_email_account_orders:read"
	ScopeDFYEmailAccountOrdersUpdate Scope = "dfy_email_account_orders:update"
	ScopeDFYEmailAccountOrdersDelete Scope = "dfy_email_account_orders:delete"
)

// Scopes for authentication.
const (
	ScopeAuthAll    Scope = "auth:all"
	ScopeAuthCreate Scope = "auth:create"
	ScopeAuthRead   Scope = "auth:read"
	ScopeAuthUpdate Scope = "auth:update"
	ScopeAuthDelete Scope = "auth:delete"
)
