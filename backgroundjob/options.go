package backgroundjob

import "github.com/mrz1836/go-instantly"

// ListOption customizes a List request.
//
// Options are typed per resource, so passing an option from another resource is
// a compile error. Only the options actually supplied are sent.
type ListOption func(*instantly.Query)

// WithLimit sets the maximum number of jobs returned in a single page.
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

// WithIDs restricts results to the given job identifiers, sent as a
// comma-separated list.
func WithIDs(ids string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("ids", ids)
	}
}

// WithIncludedIDs restricts results to include the given job identifiers, sent
// as a comma-separated list.
func WithIncludedIDs(ids string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("included_ids", ids)
	}
}

// WithExcludedIDs excludes the given job identifiers from the results, sent as a
// comma-separated list.
func WithExcludedIDs(ids string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("excluded_ids", ids)
	}
}

// WithType restricts results to jobs of a single type.
func WithType(jobType Type) ListOption {
	return func(q *instantly.Query) {
		instantly.SetEnum(q, "type", jobType)
	}
}

// WithEntityType restricts results to jobs related to a single kind of entity.
func WithEntityType(entityType EntityType) ListOption {
	return func(q *instantly.Query) {
		instantly.SetEnum(q, "entity_type", entityType)
	}
}

// WithEntityID restricts results to jobs related to the given entities, sent as
// a comma-separated list.
func WithEntityID(entityID string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("entity_id", entityID)
	}
}

// WithStatus restricts results to jobs in the given statuses, sent as a
// comma-separated list. Because a single request can filter on several statuses
// at once, it takes a raw string rather than a single Status value.
func WithStatus(status string) ListOption {
	return func(q *instantly.Query) {
		q.SetString("status", status)
	}
}

// WithSortColumn sets the column results are sorted by.
func WithSortColumn(column SortColumn) ListOption {
	return func(q *instantly.Query) {
		instantly.SetEnum(q, "sort_column", column)
	}
}

// WithSortOrder sets the direction results are sorted in.
func WithSortOrder(order instantly.SortOrder) ListOption {
	return func(q *instantly.Query) {
		instantly.SetEnum(q, "sort_order", order)
	}
}

// GetOption customizes a Get request.
type GetOption func(*instantly.Query)

// WithDataFields selects which fields of the job's data payload are returned,
// sent as a comma-separated list.
func WithDataFields(fields string) GetOption {
	return func(q *instantly.Query) {
		q.SetString("data_fields", fields)
	}
}
