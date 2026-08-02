package email

import (
	"context"
	"iter"

	"github.com/mrz1836/go-instantly"
)

// ListIter returns an iterator that walks every page of List, following
// next_starting_after until the API stops returning a cursor.
//
// It is strictly additive: List still returns exactly one page for callers who
// want to drive pagination themselves. Iteration stops at the first error, which
// is yielded with a nil *Email, and stops early when the context is canceled:
//
//	for email, err := range svc.ListIter(ctx, email.WithIsUnread(true)) {
//		if err != nil {
//			return err
//		}
//		fmt.Println(email.Subject)
//	}
//
// Every page is a separate request against an endpoint rate limited to 20
// requests per minute, so prefer narrowing the result set with options over
// walking it in full.
func (s *Service) ListIter(ctx context.Context, opts ...ListOption) iter.Seq2[*Email, error] {
	return instantly.Iterate(ctx, func(ctx context.Context, cursor string) ([]Email, string, error) {
		pageOpts := opts
		if cursor != "" {
			// The cursor is appended last so it overrides any starting cursor
			// the caller supplied, which would otherwise re-request the first
			// page forever. A fresh slice is built so the caller's options are
			// never mutated.
			pageOpts = append(append([]ListOption(nil), opts...), WithStartingAfter(cursor))
		}

		page, err := s.List(ctx, pageOpts...)
		if err != nil {
			return nil, "", err
		}

		return page.Items, page.NextStartingAfter, nil
	})
}
