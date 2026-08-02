package email

import (
	"time"

	"github.com/mrz1836/go-instantly"
)

// IStatus is the interest status an email can be filtered by.
//
// It mirrors the interest statuses recorded on a lead, delivered on an email as
// the i_status field.
type IStatus int64

// The interest statuses an email can carry.
const (
	// IStatusOutOfOffice means the lead replied out of office.
	IStatusOutOfOffice IStatus = 0

	// IStatusInterested means the lead is interested.
	IStatusInterested IStatus = 1

	// IStatusMeetingBooked means a meeting was booked with the lead.
	IStatusMeetingBooked IStatus = 2

	// IStatusMeetingCompleted means a meeting with the lead was completed.
	IStatusMeetingCompleted IStatus = 3

	// IStatusWon means the lead was won.
	IStatusWon IStatus = 4

	// IStatusNotInterested means the lead is not interested.
	IStatusNotInterested IStatus = -1

	// IStatusWrongPerson means the lead was the wrong person.
	IStatusWrongPerson IStatus = -2

	// IStatusLost means the lead was lost.
	IStatusLost IStatus = -3

	// IStatusNoShow means the lead was a no-show.
	IStatusNoShow IStatus = -4
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
// Options are typed per resource, so passing an option from another resource is
// a compile error. Only the options actually supplied are sent.
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
func WithIStatus(status IStatus) ListOption {
	return func(q *instantly.Query) {
		q.SetInt("i_status", int(status))
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
// given timestamp, sent as an RFC 3339 wire value.
func WithMinTimestampCreated(timestamp time.Time) ListOption {
	return func(q *instantly.Query) {
		q.SetString("min_timestamp_created", timestamp.Format(time.RFC3339))
	}
}

// WithMaxTimestampCreated restricts results to emails created at or before the
// given timestamp, sent as an RFC 3339 wire value.
func WithMaxTimestampCreated(timestamp time.Time) ListOption {
	return func(q *instantly.Query) {
		q.SetString("max_timestamp_created", timestamp.Format(time.RFC3339))
	}
}

// WithLatestOfThread restricts results to the most recent email of each thread
// when true.
func WithLatestOfThread(latestOfThread bool) ListOption {
	return func(q *instantly.Query) {
		q.SetBool("latest_of_thread", latestOfThread)
	}
}
