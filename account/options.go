package account

import (
	"github.com/mrz1836/go-instantly"
)

// Filter narrows a list request to a documented category of accounts.
type Filter string

// The account filters a List request can apply.
const (
	// FilterPaused returns paused accounts.
	FilterPaused Filter = "ACC_FILTER_PAUSED"

	// FilterError returns accounts with errors.
	FilterError Filter = "ACC_FILTER_ERROR"

	// FilterNoCTD returns accounts with no custom tracking domain.
	FilterNoCTD Filter = "ACC_FILTER_NO_CTD"

	// FilterPreWarmed returns pre-warmed accounts.
	FilterPreWarmed Filter = "ACC_FILTER_PW_ACCOUNTS"

	// FilterDFY returns done-for-you accounts.
	FilterDFY Filter = "ACC_FILTER_DFY"

	// FilterDFYSetupPending returns done-for-you accounts pending setup.
	FilterDFYSetupPending Filter = "ACC_FILTER_DFY_SETUP_PENDING"

	// FilterWarmupActive returns accounts whose warmup is active.
	FilterWarmupActive Filter = "ACC_FILTER_W_ACTIVE"

	// FilterWarmupPaused returns accounts whose warmup is paused.
	FilterWarmupPaused Filter = "ACC_FILTER_W_PAUSED"

	// FilterWarmupError returns accounts whose warmup has errors.
	FilterWarmupError Filter = "ACC_FILTER_W_ERROR"
)

// SortBy is the column a list request sorts by.
type SortBy string

// The columns a List request can sort by.
const (
	// SortByTimestampCreated sorts by creation time.
	SortByTimestampCreated SortBy = "timestamp_created"

	// SortByEmail sorts by email address.
	SortByEmail SortBy = "email"

	// SortByWarmupScore sorts by warmup score.
	SortByWarmupScore SortBy = "stat_warmup_score"

	// SortByStatus sorts by status.
	SortByStatus SortBy = "status"
)

// ListOption customizes a List request.
//
// Options are typed per resource, so passing an option from another resource is
// a compile error. Only the options actually supplied are sent.
type ListOption func(*instantly.Query)

// WithLimit sets the maximum number of accounts returned in a single page.
func WithLimit(limit int) ListOption {
	return func(q *instantly.Query) {
		q.SetInt("limit", limit)
	}
}

// WithStartingAfter sets the pagination cursor to resume from, which is the
// NextStartingAfter value of a previous page.
func WithStartingAfter(cursor string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("starting_after", cursor)
	}
}

// WithSearch restricts results to accounts matching a search term.
func WithSearch(term string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("search", term)
	}
}

// WithStatus restricts results to accounts with the given status.
func WithStatus(status Status) ListOption {
	return func(q *instantly.Query) {
		q.SetInt("status", int(status))
	}
}

// WithProviderCode restricts results to accounts on the given provider.
func WithProviderCode(code ProviderCode) ListOption {
	return func(q *instantly.Query) {
		q.SetInt("provider_code", int(code))
	}
}

// WithTagIDs restricts results to accounts that have any of the given tag ids,
// supplied as a comma-separated list.
func WithTagIDs(tagIDs string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("tag_ids", tagIDs)
	}
}

// WithTagIDsAll restricts results to accounts that have all of the given tag
// ids, supplied as a comma-separated list.
func WithTagIDsAll(tagIDs string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("tag_ids_all", tagIDs)
	}
}

// WithIncludeTags includes each account's tags in the response when true.
func WithIncludeTags(include bool) ListOption {
	return func(q *instantly.Query) {
		q.SetBool("include_tags", include)
	}
}

// WithFilter applies a documented account filter.
func WithFilter(filter Filter) ListOption {
	return func(q *instantly.Query) {
		instantly.SetEnum(q, "filter", filter)
	}
}

// WithSortBy sets the column results are sorted by.
func WithSortBy(sortBy SortBy) ListOption {
	return func(q *instantly.Query) {
		instantly.SetEnum(q, "sort_by", sortBy)
	}
}

// WithSortOrder sets the direction results are sorted in, which defaults to
// descending when a sort column is set.
func WithSortOrder(order instantly.SortOrder) ListOption {
	return func(q *instantly.Query) {
		instantly.SetEnum(q, "sort_order", order)
	}
}

// WithSkip sets how many accounts to skip for offset-based pagination.
func WithSkip(skip int) ListOption {
	return func(q *instantly.Query) {
		q.SetInt("skip", skip)
	}
}

// AnalyticsOption customizes a DailyAnalytics request.
type AnalyticsOption func(*instantly.Query)

// WithStartDate restricts daily analytics to on or after the given date.
func WithStartDate(date string) AnalyticsOption {
	return func(q *instantly.Query) {
		q.SetString("start_date", date)
	}
}

// WithEndDate restricts daily analytics to on or before the given date.
func WithEndDate(date string) AnalyticsOption {
	return func(q *instantly.Query) {
		q.SetString("end_date", date)
	}
}

// WithEmails restricts daily analytics to the given accounts. Each email is sent
// as a repeated emails query parameter.
func WithEmails(emails ...string) AnalyticsOption {
	return func(q *instantly.Query) {
		for _, email := range emails {
			q.AddString("emails", email)
		}
	}
}
