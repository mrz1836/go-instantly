package instantly

import (
	"context"
	"iter"
)

// FetchPage fetches a single page of a cursor-paginated list.
//
// It returns the items on the page, the cursor for the following page (empty on
// the last page), and any error. The cursor passed in is empty on the first
// call and is the previous page's next cursor thereafter.
type FetchPage[T any] func(ctx context.Context, cursor string) (items []T, next string, err error)

// Iterate turns a paginated fetch into a range-over-func sequence, following the
// cursor until the API stops returning one.
//
// It centralizes every termination guard so each resource's ListIter is a thin
// closure over its own List call:
//
//   - the context is checked before every page, so a cancellation between pages
//     stops iteration instead of issuing another request;
//   - the first error is yielded once with a nil *T and ends iteration;
//   - an empty page ends iteration even if the API keeps returning a cursor,
//     which bounds the total number of requests;
//   - iteration stops once the next cursor comes back empty.
//
// Iteration also stops as soon as the consumer breaks out of the range loop, so
// no page is fetched that the consumer never reads.
func Iterate[T any](ctx context.Context, fetch FetchPage[T]) iter.Seq2[*T, error] {
	return func(yield func(*T, error) bool) {
		cursor := ""

		for {
			next, more := yieldPage(ctx, fetch, cursor, yield)
			if !more || next == "" {
				return
			}

			cursor = next
		}
	}
}

// yieldPage fetches a single page and hands each item to yield.
//
// It reports the next page's cursor and whether iteration should continue:
// iteration stops once the consumer breaks, the context is canceled, the fetch
// fails, or the page comes back empty. Errors are yielded rather than returned so
// the consumer sees them in the sequence.
func yieldPage[T any](
	ctx context.Context, fetch FetchPage[T], cursor string, yield func(*T, error) bool,
) (string, bool) {
	// Checked before every page so a cancellation between pages stops iteration
	// instead of issuing another request.
	if err := ctx.Err(); err != nil {
		yield(nil, err)
		return "", false
	}

	items, next, err := fetch(ctx, cursor)
	if err != nil {
		yield(nil, err)
		return "", false
	}

	// An empty page ends iteration even if the API keeps returning a cursor,
	// which bounds the total number of requests.
	if len(items) == 0 {
		return "", false
	}

	for i := range items {
		if !yield(&items[i], nil) {
			return "", false
		}
	}

	return next, true
}
