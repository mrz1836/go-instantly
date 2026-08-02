package campaign

import (
	"time"

	"github.com/mrz1836/go-instantly"
)

// ListOption customizes a List request.
//
// Options are typed per resource, so passing an option from another resource is
// a compile error. Only the options actually supplied are sent.
type ListOption func(*instantly.Query)

// WithLimit sets the maximum number of campaigns returned in a single page.
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

// WithSearch restricts results to campaigns matching a search term.
func WithSearch(term string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("search", term)
	}
}

// WithTagIDs restricts results to campaigns with any of the given tag ids. Each
// id is sent as a repeated tag_ids query parameter.
func WithTagIDs(tagIDs ...string) ListOption {
	return func(q *instantly.Query) {
		for _, id := range tagIDs {
			q.AddString("tag_ids", id)
		}
	}
}

// WithAISalesAgentID restricts results to campaigns created by an AI sales agent.
func WithAISalesAgentID(agentID string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("ai_sales_agent_id", agentID)
	}
}

// WithStatus restricts results to campaigns with the given status.
func WithStatus(status Status) ListOption {
	return func(q *instantly.Query) {
		q.SetInt("status", int(status))
	}
}

// WithExcludeStatus excludes campaigns with the given status from the results.
func WithExcludeStatus(status Status) ListOption {
	return func(q *instantly.Query) {
		q.SetInt("exclude_status", int(status))
	}
}

// SearchOption customizes a SearchByContact request.
type SearchOption func(*instantly.Query)

// WithSortColumn sets the column search results are sorted by.
func WithSortColumn(column string) SearchOption {
	return func(q *instantly.Query) {
		q.SetString("sort_column", column)
	}
}

// WithSortOrder sets the direction search results are sorted in.
func WithSortOrder(order instantly.SortOrder) SearchOption {
	return func(q *instantly.Query) {
		instantly.SetEnum(q, "sort_order", order)
	}
}

// AnalyticsOption customizes an analytics request. Different analytics endpoints
// accept different parameters; apply the ones documented for the report you call.
type AnalyticsOption func(*instantly.Query)

// WithID restricts analytics to a single campaign (analytics, overview).
func WithID(id string) AnalyticsOption {
	return func(q *instantly.Query) {
		q.SetString("id", id)
	}
}

// WithIDs restricts analytics to several campaigns (analytics, overview). Each
// id is sent as a repeated ids query parameter.
func WithIDs(ids ...string) AnalyticsOption {
	return func(q *instantly.Query) {
		for _, id := range ids {
			q.AddString("ids", id)
		}
	}
}

// WithCampaignID restricts analytics to a single campaign (daily, steps).
func WithCampaignID(id string) AnalyticsOption {
	return func(q *instantly.Query) {
		q.SetString("campaign_id", id)
	}
}

// WithStartDate restricts analytics to on or after the given date. The date is
// sent as a YYYY-MM-DD wire value; only its calendar date is used.
func WithStartDate(date time.Time) AnalyticsOption {
	return func(q *instantly.Query) {
		q.SetString("start_date", date.Format(time.DateOnly))
	}
}

// WithEndDate restricts analytics to on or before the given date. The date is
// sent as a YYYY-MM-DD wire value; only its calendar date is used.
func WithEndDate(date time.Time) AnalyticsOption {
	return func(q *instantly.Query) {
		q.SetString("end_date", date.Format(time.DateOnly))
	}
}

// WithCampaignStatus restricts analytics to campaigns with the given status
// (overview, daily).
func WithCampaignStatus(status Status) AnalyticsOption {
	return func(q *instantly.Query) {
		q.SetInt("campaign_status", int(status))
	}
}

// WithExcludeTotalLeadsCount omits the total leads count from the analytics
// report when true (analytics).
func WithExcludeTotalLeadsCount(exclude bool) AnalyticsOption {
	return func(q *instantly.Query) {
		q.SetBool("exclude_total_leads_count", exclude)
	}
}

// WithExpandCRMEvents expands CRM events in the analytics overview when true.
func WithExpandCRMEvents(expand bool) AnalyticsOption {
	return func(q *instantly.Query) {
		q.SetBool("expand_crm_events", expand)
	}
}

// WithIncludeOpportunitiesCount includes the opportunities count in the steps
// report when true.
func WithIncludeOpportunitiesCount(include bool) AnalyticsOption {
	return func(q *instantly.Query) {
		q.SetBool("include_opportunities_count", include)
	}
}
