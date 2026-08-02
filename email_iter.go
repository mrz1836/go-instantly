package instantly

import (
	"context"
	"iter"
)

// ListEmailsIter returns an iterator that walks every page of ListEmails,
// following next_starting_after until the API stops returning a cursor.
//
// It is strictly additive: ListEmails still returns exactly one page for
// callers who want to drive pagination themselves. Iteration stops at the first
// error, which is yielded with a nil *Email, and stops early when the context
// is canceled:
//
//	for email, err := range client.ListEmailsIter(ctx, instantly.WithEmailIsUnread(true)) {
//		if err != nil {
//			return err
//		}
//		fmt.Println(email.Subject)
//	}
//
// Every page is a separate request against an endpoint rate limited to 20
// requests per minute, so prefer narrowing the result set with options over
// walking it in full.
func (client *Client) ListEmailsIter(ctx context.Context, opts ...EmailListOption) iter.Seq2[*Email, error] {
	return func(yield func(*Email, error) bool) {
		// The caller's options are copied once with room for the cursor, so
		// paging never mutates the slice the caller owns.
		pageOpts := make([]EmailListOption, len(opts), len(opts)+1)
		copy(pageOpts, opts)

		for {
			cursor, more := client.yieldEmailPage(ctx, pageOpts, yield)
			if !more || cursor == "" {
				return
			}

			// The cursor is appended last so it overrides any starting cursor
			// the caller supplied, which would otherwise re-request the first
			// page forever.
			pageOpts = append(pageOpts[:len(opts)], WithEmailStartingAfter(cursor))
		}
	}
}

// yieldEmailPage fetches a single page of emails and hands each one to yield.
//
// It reports the cursor of the following page and whether iteration should
// continue: iteration stops once the consumer breaks, the context is canceled,
// the request fails, or the page comes back empty. Errors are yielded rather
// than returned so the consumer sees them in the sequence.
func (client *Client) yieldEmailPage(
	ctx context.Context, opts []EmailListOption, yield func(*Email, error) bool,
) (string, bool) {
	// Checked before every page so a cancellation between pages stops
	// iteration instead of issuing another request.
	if err := ctx.Err(); err != nil {
		yield(nil, err)
		return "", false
	}

	response, err := client.ListEmails(ctx, opts...)
	if err != nil {
		yield(nil, err)
		return "", false
	}

	// An empty page ends iteration even if the API keeps returning a cursor,
	// which bounds the total number of requests.
	if len(response.Items) == 0 {
		return "", false
	}

	for i := range response.Items {
		if !yield(&response.Items[i], nil) {
			return "", false
		}
	}

	return response.NextStartingAfter, true
}
