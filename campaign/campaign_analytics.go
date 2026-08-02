package campaign

import (
	"context"

	"github.com/mrz1836/go-instantly"
)

// Analytics is the aggregate analytics for a single campaign.
type Analytics struct {
	// CampaignID is the campaign the analytics are for.
	CampaignID string `json:"campaign_id"`

	// CampaignName is the name of the campaign.
	CampaignName string `json:"campaign_name"`

	// CampaignStatus is the status of the campaign.
	CampaignStatus Status `json:"campaign_status"`

	// CampaignIsEvergreen reports whether the campaign is evergreen.
	CampaignIsEvergreen bool `json:"campaign_is_evergreen"`

	// LeadsCount is the number of leads in the campaign.
	LeadsCount int64 `json:"leads_count"`

	// ContactedCount is the number of leads contacted.
	ContactedCount int64 `json:"contacted_count"`

	// NewLeadsContactedCount is the number of new leads contacted.
	NewLeadsContactedCount int64 `json:"new_leads_contacted_count"`

	// EmailsSentCount is the number of emails sent.
	EmailsSentCount int64 `json:"emails_sent_count"`

	// OpenCount is the number of opens.
	OpenCount int64 `json:"open_count"`

	// OpenCountUnique is the number of unique opens.
	OpenCountUnique int64 `json:"open_count_unique"`

	// OpenCountUniqueByStep is the number of unique opens by step.
	OpenCountUniqueByStep int64 `json:"open_count_unique_by_step"`

	// LinkClickCount is the number of link clicks.
	LinkClickCount int64 `json:"link_click_count"`

	// LinkClickCountUnique is the number of unique link clicks.
	LinkClickCountUnique int64 `json:"link_click_count_unique"`

	// LinkClickCountUniqueByStep is the number of unique link clicks by step.
	LinkClickCountUniqueByStep int64 `json:"link_click_count_unique_by_step"`

	// ReplyCount is the number of replies.
	ReplyCount int64 `json:"reply_count"`

	// ReplyCountUnique is the number of unique replies.
	ReplyCountUnique int64 `json:"reply_count_unique"`

	// ReplyCountUniqueByStep is the number of unique replies by step.
	ReplyCountUniqueByStep int64 `json:"reply_count_unique_by_step"`

	// ReplyCountAutomatic is the number of automatic replies.
	ReplyCountAutomatic int64 `json:"reply_count_automatic"`

	// ReplyCountAutomaticUnique is the number of unique automatic replies.
	ReplyCountAutomaticUnique int64 `json:"reply_count_automatic_unique"`

	// ReplyCountAutomaticUniqueByStep is the number of unique automatic replies by step.
	ReplyCountAutomaticUniqueByStep int64 `json:"reply_count_automatic_unique_by_step"`

	// BouncedCount is the number of bounces.
	BouncedCount int64 `json:"bounced_count"`

	// UnsubscribedCount is the number of unsubscribes.
	UnsubscribedCount int64 `json:"unsubscribed_count"`

	// CompletedCount is the number of leads that completed the campaign.
	CompletedCount int64 `json:"completed_count"`

	// TotalOpportunities is the number of opportunities.
	TotalOpportunities int64 `json:"total_opportunities"`

	// TotalOpportunityValue is the total value of opportunities.
	TotalOpportunityValue float64 `json:"total_opportunity_value"`
}

// AnalyticsOverview is the aggregate analytics across one or more campaigns,
// including CRM outcome totals.
type AnalyticsOverview struct {
	// ContactedCount is the number of leads contacted.
	ContactedCount int64 `json:"contacted_count"`

	// NewLeadsContactedCount is the number of new leads contacted.
	NewLeadsContactedCount int64 `json:"new_leads_contacted_count"`

	// EmailsSentCount is the number of emails sent.
	EmailsSentCount int64 `json:"emails_sent_count"`

	// OpenCount is the number of opens.
	OpenCount int64 `json:"open_count"`

	// OpenCountUnique is the number of unique opens.
	OpenCountUnique int64 `json:"open_count_unique"`

	// OpenCountUniqueByStep is the number of unique opens by step.
	OpenCountUniqueByStep int64 `json:"open_count_unique_by_step"`

	// LinkClickCount is the number of link clicks.
	LinkClickCount int64 `json:"link_click_count"`

	// LinkClickCountUnique is the number of unique link clicks.
	LinkClickCountUnique int64 `json:"link_click_count_unique"`

	// LinkClickCountUniqueByStep is the number of unique link clicks by step.
	LinkClickCountUniqueByStep int64 `json:"link_click_count_unique_by_step"`

	// ReplyCount is the number of replies.
	ReplyCount int64 `json:"reply_count"`

	// ReplyCountUnique is the number of unique replies.
	ReplyCountUnique int64 `json:"reply_count_unique"`

	// ReplyCountUniqueByStep is the number of unique replies by step.
	ReplyCountUniqueByStep int64 `json:"reply_count_unique_by_step"`

	// ReplyCountAutomatic is the number of automatic replies.
	ReplyCountAutomatic int64 `json:"reply_count_automatic"`

	// ReplyCountAutomaticUnique is the number of unique automatic replies.
	ReplyCountAutomaticUnique int64 `json:"reply_count_automatic_unique"`

	// ReplyCountAutomaticUniqueByStep is the number of unique automatic replies by step.
	ReplyCountAutomaticUniqueByStep int64 `json:"reply_count_automatic_unique_by_step"`

	// BouncedCount is the number of bounces.
	BouncedCount int64 `json:"bounced_count"`

	// UnsubscribedCount is the number of unsubscribes.
	UnsubscribedCount int64 `json:"unsubscribed_count"`

	// CompletedCount is the number of leads that completed the campaign.
	CompletedCount int64 `json:"completed_count"`

	// TotalOpportunities is the number of opportunities.
	TotalOpportunities int64 `json:"total_opportunities"`

	// TotalOpportunityValue is the total value of opportunities.
	TotalOpportunityValue float64 `json:"total_opportunity_value"`

	// TotalInterested is the number of interested leads.
	TotalInterested int64 `json:"total_interested"`

	// TotalMeetingBooked is the number of meetings booked.
	TotalMeetingBooked int64 `json:"total_meeting_booked"`

	// TotalMeetingCompleted is the number of meetings completed.
	TotalMeetingCompleted int64 `json:"total_meeting_completed"`

	// TotalClosed is the number of closed deals.
	TotalClosed int64 `json:"total_closed"`
}

// DailyAnalytics is a single day of campaign analytics.
type DailyAnalytics struct {
	// Date is the day the analytics are for.
	Date string `json:"date"`

	// Sent is the number of emails sent.
	Sent int64 `json:"sent"`

	// Opened is the number of opens.
	Opened int64 `json:"opened"`

	// UniqueOpened is the number of unique opens.
	UniqueOpened int64 `json:"unique_opened"`

	// Replies is the number of replies.
	Replies int64 `json:"replies"`

	// UniqueReplies is the number of unique replies.
	UniqueReplies int64 `json:"unique_replies"`

	// RepliesAutomatic is the number of automatic replies.
	RepliesAutomatic int64 `json:"replies_automatic"`

	// UniqueRepliesAutomatic is the number of unique automatic replies.
	UniqueRepliesAutomatic int64 `json:"unique_replies_automatic"`

	// Clicks is the number of clicks.
	Clicks int64 `json:"clicks"`

	// UniqueClicks is the number of unique clicks.
	UniqueClicks int64 `json:"unique_clicks"`

	// Contacted is the number of leads contacted.
	Contacted int64 `json:"contacted"`

	// NewLeadsContacted is the number of new leads contacted.
	NewLeadsContacted int64 `json:"new_leads_contacted"`

	// Opportunities is the number of opportunities.
	Opportunities int64 `json:"opportunities"`

	// UniqueOpportunities is the number of unique opportunities.
	UniqueOpportunities int64 `json:"unique_opportunities"`
}

// StepAnalytics is the analytics for a single campaign step and variant.
type StepAnalytics struct {
	// Step is the campaign step the analytics are for.
	Step *string `json:"step,omitempty"`

	// Variant is the variant the analytics are for.
	Variant *string `json:"variant,omitempty"`

	// Sent is the number of emails sent.
	Sent int64 `json:"sent"`

	// Opened is the number of opens.
	Opened int64 `json:"opened"`

	// UniqueOpened is the number of unique opens.
	UniqueOpened int64 `json:"unique_opened"`

	// Replies is the number of replies.
	Replies int64 `json:"replies"`

	// UniqueReplies is the number of unique replies.
	UniqueReplies int64 `json:"unique_replies"`

	// RepliesAutomatic is the number of automatic replies.
	RepliesAutomatic int64 `json:"replies_automatic"`

	// Clicks is the number of clicks.
	Clicks int64 `json:"clicks"`

	// UniqueClicks is the number of unique clicks.
	UniqueClicks int64 `json:"unique_clicks"`

	// Opportunities is the number of opportunities.
	Opportunities int64 `json:"opportunities"`

	// UniqueOpportunities is the number of unique opportunities.
	UniqueOpportunities int64 `json:"unique_opportunities"`

	// MeetingsBooked is the number of meetings booked from this step.
	MeetingsBooked int64 `json:"meetings_booked"`

	// Won is the number of deals won from this step.
	Won int64 `json:"won"`
}

// Analytics returns the aggregate analytics for the campaigns matching the
// supplied options.
func (s *Service) Analytics(ctx context.Context, opts ...AnalyticsOption) ([]Analytics, error) {
	q := applyAnalytics(opts)

	var out []Analytics
	if err := s.client.Get(ctx, q.Path(basePath+"/analytics"), &out); err != nil {
		return nil, err
	}

	return out, nil
}

// AnalyticsOverview returns the aggregate analytics overview for the campaigns
// matching the supplied options.
func (s *Service) AnalyticsOverview(ctx context.Context, opts ...AnalyticsOption) (*AnalyticsOverview, error) {
	q := applyAnalytics(opts)

	out := &AnalyticsOverview{}
	if err := s.client.Get(ctx, q.Path(basePath+"/analytics/overview"), out); err != nil {
		return nil, err
	}

	return out, nil
}

// DailyAnalytics returns the daily analytics for the campaigns matching the
// supplied options.
func (s *Service) DailyAnalytics(ctx context.Context, opts ...AnalyticsOption) ([]DailyAnalytics, error) {
	q := applyAnalytics(opts)

	var out []DailyAnalytics
	if err := s.client.Get(ctx, q.Path(basePath+"/analytics/daily"), &out); err != nil {
		return nil, err
	}

	return out, nil
}

// StepsAnalytics returns the per-step analytics for the campaigns matching the
// supplied options.
func (s *Service) StepsAnalytics(ctx context.Context, opts ...AnalyticsOption) ([]StepAnalytics, error) {
	q := applyAnalytics(opts)

	var out []StepAnalytics
	if err := s.client.Get(ctx, q.Path(basePath+"/analytics/steps"), &out); err != nil {
		return nil, err
	}

	return out, nil
}

// applyAnalytics accumulates the analytics options into a query.
func applyAnalytics(opts []AnalyticsOption) *instantly.Query {
	q := instantly.NewQuery()
	for _, opt := range opts {
		if opt != nil {
			opt(q)
		}
	}

	return q
}
