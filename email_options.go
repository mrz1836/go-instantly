package instantly

import "net/url"

// EmailListOption customizes a ListEmails request.
//
// Options are typed per resource rather than shared across the package, so
// passing an option belonging to another resource is a compile error instead of
// a query parameter that is silently ignored.
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
