package instantly

import (
	"net/url"
	"strconv"
)

// Query accumulates the query parameters a list request sends.
//
// Resource packages expose their own typed ListOption values that write into a
// Query, which keeps every query key and value compile-time checked while
// sharing a single accumulator and encoder across all 28 resources. Setting the
// same key twice keeps the last value.
type Query struct {
	values url.Values
}

// NewQuery returns an empty Query ready for options to write into.
func NewQuery() *Query {
	return &Query{values: url.Values{}}
}

// SetString sets a string query parameter, replacing any existing value.
func (q *Query) SetString(key, value string) *Query {
	q.values.Set(key, value)
	return q
}

// AddString appends a string value to a query parameter, keeping any values
// already set. It is how a repeated array parameter (key=a&key=b) is built.
func (q *Query) AddString(key, value string) *Query {
	q.values.Add(key, value)
	return q
}

// SetInt sets an integer query parameter.
func (q *Query) SetInt(key string, value int) *Query {
	q.values.Set(key, strconv.Itoa(value))
	return q
}

// SetBool sets a boolean query parameter, rendered as "true" or "false".
func (q *Query) SetBool(key string, value bool) *Query {
	q.values.Set(key, strconv.FormatBool(value))
	return q
}

// Get returns the first value set for key, or the empty string when key is
// unset. It is chiefly a testing convenience for asserting what an option wrote.
func (q *Query) Get(key string) string {
	return q.values.Get(key)
}

// Len reports how many distinct query parameters have been set. Tests use it as
// a completeness guard, asserting an option renders exactly one parameter.
func (q *Query) Len() int {
	return len(q.values)
}

// Encode renders the accumulated parameters as a URL-encoded query string.
func (q *Query) Encode() string {
	return q.values.Encode()
}

// Path appends the encoded parameters to base, returning the bare base when no
// parameter has been set so an unfiltered request never carries an empty "?".
func (q *Query) Path(base string) string {
	if q == nil {
		return base
	}

	return BuildPath(base, q.values)
}

// SetEnum sets a string-backed enum query parameter.
//
// It is a free function rather than a method because Go methods cannot take
// type parameters; resource options call it as instantly.SetEnum(q, key, value).
func SetEnum[E ~string](q *Query, key string, value E) *Query {
	q.values.Set(key, string(value))
	return q
}

// BuildPath appends encoded query parameters to a path, returning the bare path
// when there are none.
func BuildPath(base string, values url.Values) string {
	if len(values) == 0 {
		return base
	}

	return base + "?" + values.Encode()
}
