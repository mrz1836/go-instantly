package instantly

import (
	"context"
	"iter"
	"net/http"
	"net/url"
)

// Page is one page of a cursor-paginated list, shared by every resource.
//
// Each resource package aliases its ListResponse to a Page of its own item type
// (type ListResponse = instantly.Page[Campaign]). Because the alias is an
// identical type rather than a distinct defined one, the generic helpers below —
// Paginate in particular — accept a resource's List method directly.
type Page[T any] struct {
	// Items are the resources on this page.
	Items []T `json:"items"`

	// NextStartingAfter is the cursor for the following page, and is empty on
	// the last page.
	NextStartingAfter string `json:"next_starting_after,omitempty"`
}

// CountResponse is the {"count":N} wrapper the counting endpoints return.
//
// It is the single decode target the launched-campaign, unread-email, and
// bulk-delete counts share, in place of a per-resource clone.
type CountResponse struct {
	// Count is the counted value the endpoint reported.
	Count int64 `json:"count"`
}

// ApplyOptions builds a Query from the supplied typed per-resource options.
//
// Every resource defines its own ListOption as func(*Query); the ~func(*Query)
// constraint accepts any of them, so this single helper replaces the identical
// apply loop each List method used to carry. A nil option is skipped rather than
// invoked.
func ApplyOptions[O ~func(*Query)](opts ...O) *Query {
	q := NewQuery()
	for _, opt := range opts {
		if opt != nil {
			opt(q)
		}
	}

	return q
}

// GetResult performs a GET and decodes the response into a fresh T.
//
// It collapses the out := &T{}; if err := client.Get(...); return out, nil block
// every read method used to repeat. On any error it returns a nil *T, so a
// failed decode never hands back a partly populated value.
func GetResult[T any](ctx context.Context, c *Client, path string) (*T, error) {
	return doResult[T](ctx, c, http.MethodGet, path, nil)
}

// PostResult performs a POST with payload and decodes the response into a fresh
// T. It returns a nil *T on any error.
func PostResult[T any](ctx context.Context, c *Client, path string, payload any) (*T, error) {
	return doResult[T](ctx, c, http.MethodPost, path, payload)
}

// PatchResult performs a PATCH with payload and decodes the response into a
// fresh T. It returns a nil *T on any error.
func PatchResult[T any](ctx context.Context, c *Client, path string, payload any) (*T, error) {
	return doResult[T](ctx, c, http.MethodPatch, path, payload)
}

// DeleteResult performs a DELETE and decodes the response into a fresh T. It
// returns a nil *T on any error.
func DeleteResult[T any](ctx context.Context, c *Client, path string) (*T, error) {
	return doResult[T](ctx, c, http.MethodDelete, path, nil)
}

// doResult is the shared core of the typed result helpers: it decodes a single
// resource in one call and returns a nil *T whenever the request or decode
// fails, so a caller never sees a partly populated value alongside an error.
func doResult[T any](ctx context.Context, c *Client, method, path string, payload any) (*T, error) {
	out := new(T)
	if err := c.Do(ctx, method, path, payload, out); err != nil {
		return nil, err
	}

	return out, nil
}

// Paginate turns a resource's List method into a range-over-func sequence,
// carrying the caller's options onto every page and overriding the pagination
// cursor with each page's next_starting_after.
//
// It collapses the near-identical ListIter closures the GET-with-options
// resources used to carry: pass the caller's options, the option constructor
// that sets the cursor (WithStartingAfter), and the List method itself. The
// caller's option slice is never mutated. All the termination guards live in
// Iterate, which this delegates to.
func Paginate[T any, O ~func(*Query)](
	ctx context.Context,
	opts []O,
	withCursor func(string) O,
	list func(ctx context.Context, opts ...O) (*Page[T], error),
) iter.Seq2[*T, error] {
	return Iterate(ctx, func(ctx context.Context, cursor string) ([]T, string, error) {
		pageOpts := opts
		if cursor != "" {
			// The cursor is appended last so it overrides any starting cursor the
			// caller supplied. A fresh slice keeps the caller's options unmutated.
			pageOpts = append(append([]O(nil), opts...), withCursor(cursor))
		}

		page, err := list(ctx, pageOpts...)
		if err != nil {
			return nil, "", err
		}

		return page.Items, page.NextStartingAfter, nil
	})
}

// JoinPath escapes each segment and joins it onto base with a slash.
//
// It replaces the hand-built base + "/" + url.PathEscape(id) joins, escaping
// every segment so an id carrying a slash, space, or other reserved character
// cannot alter the request path. With no segments it returns base unchanged.
func JoinPath(base string, segments ...string) string {
	for _, segment := range segments {
		base += "/" + url.PathEscape(segment)
	}

	return base
}
