package instantly

import (
	"net/url"
	"strconv"
)

// EmailMode filters listed emails by the inbox category they belong to.
type EmailMode string

// The inbox categories a ListEmails request can filter on.
const (
	// EmailModeFocused restricts results to the focused inbox.
	EmailModeFocused EmailMode = "emode_focused"

	// EmailModeOthers restricts results to everything outside the focused inbox.
	EmailModeOthers EmailMode = "emode_others"

	// EmailModeAll returns emails from every inbox.
	EmailModeAll EmailMode = "emode_all"
)

// SortOrder is the direction results are sorted in.
type SortOrder string

// The directions results can be sorted in.
const (
	// SortOrderAsc sorts results oldest first.
	SortOrderAsc SortOrder = "asc"

	// SortOrderDesc sorts results newest first.
	SortOrderDesc SortOrder = "desc"
)

// EmailType filters listed emails by how they came to exist.
type EmailType string

// The ways an email can come to exist.
const (
	// EmailTypeReceived restricts results to emails received from a lead.
	EmailTypeReceived EmailType = "received"

	// EmailTypeSent restricts results to emails sent by a campaign.
	EmailTypeSent EmailType = "sent"

	// EmailTypeManual restricts results to emails sent manually.
	EmailTypeManual EmailType = "manual"
)

// EmailListOption customizes a ListEmails request.
//
// Options are typed per resource rather than shared across the package, so
// passing an option belonging to another resource is a compile error instead of
// a query parameter that is silently ignored. Only the options actually passed
// reach the API: an option that is never supplied sends no query parameter at
// all, rather than sending an empty one.
type EmailListOption func(*emailListQuery)

// emailListQuery accumulates the query parameters a ListEmails request sends.
type emailListQuery struct {
	values url.Values
}

// newEmailListQuery applies the supplied options and renders them as query
// parameters, returning nil when no option was supplied.
func newEmailListQuery(opts ...EmailListOption) url.Values {
	query := &emailListQuery{values: url.Values{}}

	for _, opt := range opts {
		if opt != nil {
			opt(query)
		}
	}

	if len(query.values) == 0 {
		return nil
	}

	return query.values
}

// WithEmailLimit sets the maximum number of emails returned in a single page.
func WithEmailLimit(limit int) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("limit", strconv.Itoa(limit))
	}
}

// WithEmailStartingAfter sets the pagination cursor to resume from, which is
// the NextStartingAfter value of a previous page.
//
// The cursor is an opaque value that may be an identifier or a timestamp
// depending on the request, so it is passed through verbatim.
func WithEmailStartingAfter(cursor string) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("starting_after", cursor)
	}
}

// WithEmailSearch restricts results to emails matching a search term.
func WithEmailSearch(term string) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("search", term)
	}
}

// WithEmailCampaignID restricts results to emails belonging to a campaign.
func WithEmailCampaignID(campaignID string) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("campaign_id", campaignID)
	}
}

// WithEmailListID restricts results to emails belonging to a lead list.
func WithEmailListID(listID string) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("list_id", listID)
	}
}

// WithEmailIStatus restricts results to emails with the given interest status.
func WithEmailIStatus(status int) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("i_status", strconv.Itoa(status))
	}
}

// WithEmailAccount restricts results to emails belonging to a sending account.
//
// The API spells this parameter eaccount; the option is named for how the value
// reads in Go rather than for its wire name.
func WithEmailAccount(account string) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("eaccount", account)
	}
}

// WithEmailIsUnread restricts results to unread emails when true, and to read
// emails when false.
func WithEmailIsUnread(isUnread bool) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("is_unread", strconv.FormatBool(isUnread))
	}
}

// WithEmailHasReminder restricts results to emails that carry a reminder when
// true, and to emails without one when false.
func WithEmailHasReminder(hasReminder bool) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("has_reminder", strconv.FormatBool(hasReminder))
	}
}

// WithEmailMode restricts results to a single inbox category.
func WithEmailMode(mode EmailMode) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("mode", string(mode))
	}
}

// WithEmailPreviewOnly returns a content preview in place of the full email
// body when true.
func WithEmailPreviewOnly(previewOnly bool) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("preview_only", strconv.FormatBool(previewOnly))
	}
}

// WithEmailSortOrder sets the direction results are sorted in.
func WithEmailSortOrder(order SortOrder) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("sort_order", string(order))
	}
}

// WithEmailScheduledOnly restricts results to scheduled emails when true.
func WithEmailScheduledOnly(scheduledOnly bool) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("scheduled_only", strconv.FormatBool(scheduledOnly))
	}
}

// WithEmailAssignedTo restricts results to emails assigned to a user.
func WithEmailAssignedTo(userID string) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("assigned_to", userID)
	}
}

// WithEmailLead restricts results to emails relating to a lead, identified by
// the lead's email address.
func WithEmailLead(lead string) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("lead", lead)
	}
}

// WithEmailCompanyDomain restricts results to emails relating to a company
// domain.
func WithEmailCompanyDomain(domain string) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("company_domain", domain)
	}
}

// WithEmailMarkedAsDone restricts results to emails marked as done when true,
// and to emails that are not when false.
func WithEmailMarkedAsDone(markedAsDone bool) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("marked_as_done", strconv.FormatBool(markedAsDone))
	}
}

// WithEmailType restricts results to emails that came to exist in a particular
// way.
func WithEmailType(emailType EmailType) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("email_type", string(emailType))
	}
}

// WithEmailMinTimestampCreated restricts results to emails created at or after
// the given timestamp.
func WithEmailMinTimestampCreated(timestamp string) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("min_timestamp_created", timestamp)
	}
}

// WithEmailMaxTimestampCreated restricts results to emails created at or before
// the given timestamp.
func WithEmailMaxTimestampCreated(timestamp string) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("max_timestamp_created", timestamp)
	}
}

// WithEmailLatestOfThread restricts results to the most recent email of each
// thread when true.
func WithEmailLatestOfThread(latestOfThread bool) EmailListOption {
	return func(query *emailListQuery) {
		query.values.Set("latest_of_thread", strconv.FormatBool(latestOfThread))
	}
}
