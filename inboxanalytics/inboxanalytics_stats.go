package inboxanalytics

import "context"

// StatsByTestIDRequest is the body of a stats-by-test-id request.
//
// The filter slices are typed enums, distinct from the comma-joined string query
// parameters the List options use, because the API accepts them as JSON arrays
// here.
type StatsByTestIDRequest struct {
	// TestIDs are the tests to report stats for. Required.
	TestIDs []string `json:"test_ids"`

	// DateFrom restricts the stats to on or after this date.
	DateFrom string `json:"date_from,omitempty"`

	// DateTo restricts the stats to on or before this date.
	DateTo string `json:"date_to,omitempty"`

	// SenderEmail restricts the stats to a single sender address.
	SenderEmail string `json:"sender_email,omitempty"`

	// RecipientESP restricts the stats to the given recipient providers.
	RecipientESP []ESP `json:"recipient_esp,omitempty"`

	// RecipientGeo restricts the stats to the given recipient regions.
	RecipientGeo []Geo `json:"recipient_geo,omitempty"`

	// RecipientType restricts the stats to the given recipient types.
	RecipientType []RecipientType `json:"recipient_type,omitempty"`
}

// TestStats is the aggregate placement stats for a single test.
type TestStats struct {
	// TestID identifies the test the stats are for.
	TestID string `json:"test_id"`

	// Count is the total number of measured emails.
	Count float64 `json:"count"`

	// SpamCount is the number that landed in spam.
	SpamCount float64 `json:"spam_count"`

	// SpamPercent is the percentage that landed in spam.
	SpamPercent float64 `json:"spam_percent"`

	// InboxCount is the number that landed in the inbox.
	InboxCount float64 `json:"inbox_count"`

	// InboxPercent is the percentage that landed in the inbox.
	InboxPercent float64 `json:"inbox_percent"`

	// CategoryCount is the number that landed in a category tab.
	CategoryCount float64 `json:"category_count"`

	// CategoryPercent is the percentage that landed in a category tab.
	CategoryPercent float64 `json:"category_percent"`
}

// DeliverabilityInsightsRequest is the body of a deliverability-insights request.
type DeliverabilityInsightsRequest struct {
	// TestID identifies the test to report insights for. Required.
	TestID string `json:"test_id"`

	// DateFrom restricts the insights to on or after this date.
	DateFrom string `json:"date_from,omitempty"`

	// DateTo restricts the insights to on or before this date.
	DateTo string `json:"date_to,omitempty"`

	// PreviousDateFrom is the start of the comparison window.
	PreviousDateFrom string `json:"previous_date_from,omitempty"`

	// PreviousDateTo is the end of the comparison window.
	PreviousDateTo string `json:"previous_date_to,omitempty"`

	// ShowPrevious includes the comparison-window figures when true.
	ShowPrevious *bool `json:"show_previous,omitempty"`

	// RecipientESP restricts the insights to the given recipient providers.
	RecipientESP []ESP `json:"recipient_esp,omitempty"`

	// RecipientGeo restricts the insights to the given recipient regions.
	RecipientGeo []Geo `json:"recipient_geo,omitempty"`

	// RecipientType restricts the insights to the given recipient types.
	RecipientType []RecipientType `json:"recipient_type,omitempty"`
}

// DeliverabilityInsight is a single deliverability insight row.
//
// Every field is nullable, so all are pointers: an absent value stays
// distinguishable from a zero value.
type DeliverabilityInsight struct {
	// TestID identifies the test the insight is for.
	TestID string `json:"test_id"`

	// RecipientESP is the recipient's email service provider.
	RecipientESP *ESP `json:"recipient_esp,omitempty"`

	// SenderESP is the sender's email service provider.
	SenderESP *ESP `json:"sender_esp,omitempty"`

	// From is the start of the reporting window.
	From *string `json:"from,omitempty"`

	// To is the end of the reporting window.
	To *string `json:"to,omitempty"`

	// PreviousFrom is the start of the comparison window.
	PreviousFrom *string `json:"previous_from,omitempty"`

	// PreviousTo is the end of the comparison window.
	PreviousTo *string `json:"previous_to,omitempty"`

	// InboxPercentage is the percentage that landed in the inbox.
	InboxPercentage *float64 `json:"inbox_percentage,omitempty"`

	// SpamPercentage is the percentage that landed in spam.
	SpamPercentage *float64 `json:"spam_percentage,omitempty"`

	// CategoryPercentage is the percentage that landed in a category tab.
	CategoryPercentage *float64 `json:"category_percentage,omitempty"`

	// PrevInboxPercentage is the inbox percentage in the comparison window.
	PrevInboxPercentage *float64 `json:"prev_inbox_percentage,omitempty"`

	// PrevSpamPercentage is the spam percentage in the comparison window.
	PrevSpamPercentage *float64 `json:"prev_spam_percentage,omitempty"`

	// PrevCategoryPercentage is the category percentage in the comparison window.
	PrevCategoryPercentage *float64 `json:"prev_category_percentage,omitempty"`
}

// StatsByDateRequest is the body of a stats-by-date request.
type StatsByDateRequest struct {
	// TestID identifies the test to report stats for. Required.
	TestID string `json:"test_id"`

	// DateFrom restricts the stats to on or after this date.
	DateFrom string `json:"date_from,omitempty"`

	// DateTo restricts the stats to on or before this date.
	DateTo string `json:"date_to,omitempty"`

	// SenderEmail restricts the stats to a single sender address.
	SenderEmail string `json:"sender_email,omitempty"`

	// RecipientESP restricts the stats to the given recipient providers.
	RecipientESP []ESP `json:"recipient_esp,omitempty"`

	// RecipientGeo restricts the stats to the given recipient regions.
	RecipientGeo []Geo `json:"recipient_geo,omitempty"`

	// RecipientType restricts the stats to the given recipient types.
	RecipientType []RecipientType `json:"recipient_type,omitempty"`
}

// DateStats is the aggregate placement stats for a single calendar date.
type DateStats struct {
	// TimestampCreatedDate is the calendar date the stats are for.
	TimestampCreatedDate string `json:"timestamp_created_date"`

	// SentCount is the number of emails sent that day.
	SentCount float64 `json:"sent_count"`

	// ReceivedCount is the number of emails received that day.
	ReceivedCount float64 `json:"received_count"`

	// SpamCount is the number that landed in spam that day.
	SpamCount float64 `json:"spam_count"`

	// InboxCount is the number that landed in the inbox that day.
	InboxCount float64 `json:"inbox_count"`

	// CategoryCount is the number that landed in a category tab that day.
	CategoryCount float64 `json:"category_count"`
}

// StatsByTestID returns aggregate placement stats for one or more tests.
func (s *Service) StatsByTestID(ctx context.Context, req StatsByTestIDRequest) ([]TestStats, error) {
	var out []TestStats
	if err := s.client.Post(ctx, basePath+"/stats-by-test-id", req, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// DeliverabilityInsights returns deliverability insights for a test, optionally
// against a comparison window.
func (s *Service) DeliverabilityInsights(
	ctx context.Context, req DeliverabilityInsightsRequest,
) ([]DeliverabilityInsight, error) {
	var out []DeliverabilityInsight
	if err := s.client.Post(ctx, basePath+"/deliverability-insights", req, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// StatsByDate returns aggregate placement stats for a test, broken down by date.
func (s *Service) StatsByDate(ctx context.Context, req StatsByDateRequest) ([]DateStats, error) {
	var out []DateStats
	if err := s.client.Post(ctx, basePath+"/stats-by-date", req, &out); err != nil {
		return nil, err
	}

	return out, nil
}
