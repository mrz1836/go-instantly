package email

import (
	"github.com/mrz1836/go-instantly"
)

// Mode filters listed emails by the inbox category they belong to.
type Mode string

// The inbox categories a List request can filter on.
const (
	// ModeFocused restricts results to the focused inbox.
	ModeFocused Mode = "emode_focused"

	// ModeOthers restricts results to everything outside the focused inbox.
	ModeOthers Mode = "emode_others"

	// ModeAll returns emails from every inbox.
	ModeAll Mode = "emode_all"
)

// Type filters listed emails by how they came to exist.
type Type string

// The ways an email can come to exist.
const (
	// TypeReceived restricts results to emails received from a lead.
	TypeReceived Type = "received"

	// TypeSent restricts results to emails sent by a campaign.
	TypeSent Type = "sent"

	// TypeManual restricts results to emails sent manually.
	TypeManual Type = "manual"
)

// ListOption customizes a List request.
//
// Options are typed per resource rather than shared across the SDK, so passing
// an option belonging to another resource is a compile error instead of a query
// parameter that is silently ignored. Only the options actually passed reach the
// API: an option that is never supplied sends no query parameter at all, rather
// than sending an empty one.
type ListOption func(*instantly.Query)

// WithLimit sets the maximum number of emails returned in a single page.
func WithLimit(limit int) ListOption {
	return func(q *instantly.Query) {
		q.SetInt("limit", limit)
	}
}

// WithStartingAfter sets the pagination cursor to resume from, which is the
// NextStartingAfter value of a previous page.
//
// The cursor is an opaque value that may be an identifier or a timestamp
// depending on the request, so it is passed through verbatim.
func WithStartingAfter(cursor string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("starting_after", cursor)
	}
}

// WithSearch restricts results to emails matching a search term.
func WithSearch(term string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("search", term)
	}
}

// WithCampaignID restricts results to emails belonging to a campaign.
func WithCampaignID(campaignID string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("campaign_id", campaignID)
	}
}

// WithListID restricts results to emails belonging to a lead list.
func WithListID(listID string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("list_id", listID)
	}
}

// WithIStatus restricts results to emails with the given interest status.
func WithIStatus(status int) ListOption {
	return func(q *instantly.Query) {
		q.SetInt("i_status", status)
	}
}

// WithAccount restricts results to emails belonging to a sending account.
//
// The API spells this parameter eaccount; the option is named for how the value
// reads in Go rather than for its wire name.
func WithAccount(account string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("eaccount", account)
	}
}

// WithIsUnread restricts results to unread emails when true, and to read emails
// when false.
func WithIsUnread(isUnread bool) ListOption {
	return func(q *instantly.Query) {
		q.SetBool("is_unread", isUnread)
	}
}

// WithHasReminder restricts results to emails that carry a reminder when true,
// and to emails without one when false.
func WithHasReminder(hasReminder bool) ListOption {
	return func(q *instantly.Query) {
		q.SetBool("has_reminder", hasReminder)
	}
}

// WithMode restricts results to a single inbox category.
func WithMode(mode Mode) ListOption {
	return func(q *instantly.Query) {
		instantly.SetEnum(q, "mode", mode)
	}
}

// WithPreviewOnly returns a content preview in place of the full email body when
// true.
func WithPreviewOnly(previewOnly bool) ListOption {
	return func(q *instantly.Query) {
		q.SetBool("preview_only", previewOnly)
	}
}

// WithSortOrder sets the direction results are sorted in.
func WithSortOrder(order instantly.SortOrder) ListOption {
	return func(q *instantly.Query) {
		instantly.SetEnum(q, "sort_order", order)
	}
}

// WithScheduledOnly restricts results to scheduled emails when true.
func WithScheduledOnly(scheduledOnly bool) ListOption {
	return func(q *instantly.Query) {
		q.SetBool("scheduled_only", scheduledOnly)
	}
}

// WithAssignedTo restricts results to emails assigned to a user.
func WithAssignedTo(userID string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("assigned_to", userID)
	}
}

// WithLead restricts results to emails relating to a lead, identified by the
// lead's email address.
func WithLead(lead string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("lead", lead)
	}
}

// WithCompanyDomain restricts results to emails relating to a company domain.
func WithCompanyDomain(domain string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("company_domain", domain)
	}
}

// WithMarkedAsDone restricts results to emails marked as done when true, and to
// emails that are not when false.
func WithMarkedAsDone(markedAsDone bool) ListOption {
	return func(q *instantly.Query) {
		q.SetBool("marked_as_done", markedAsDone)
	}
}

// WithType restricts results to emails that came to exist in a particular way.
func WithType(emailType Type) ListOption {
	return func(q *instantly.Query) {
		instantly.SetEnum(q, "email_type", emailType)
	}
}

// WithMinTimestampCreated restricts results to emails created at or after the
// given timestamp.
func WithMinTimestampCreated(timestamp string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("min_timestamp_created", timestamp)
	}
}

// WithMaxTimestampCreated restricts results to emails created at or before the
// given timestamp.
func WithMaxTimestampCreated(timestamp string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("max_timestamp_created", timestamp)
	}
}

// WithLatestOfThread restricts results to the most recent email of each thread
// when true.
func WithLatestOfThread(latestOfThread bool) ListOption {
	return func(q *instantly.Query) {
		q.SetBool("latest_of_thread", latestOfThread)
	}
}
