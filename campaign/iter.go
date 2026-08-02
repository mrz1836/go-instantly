package campaign

import (
	"context"
	"iter"

	"github.com/mrz1836/go-instantly"
)

// ListIter returns an iterator that walks every page of List, following
// next_starting_after until the API stops returning a cursor.
//
// It is strictly additive: List still returns one page for callers who want to
// drive pagination themselves. Iteration stops at the first error, which is
// yielded with a nil *Campaign, and stops early when the context is canceled.
func (s *Service) ListIter(ctx context.Context, opts ...ListOption) iter.Seq2[*Campaign, error] {
	return instantly.Iterate(ctx, func(ctx context.Context, cursor string) ([]Campaign, string, error) {
		pageOpts := opts
		if cursor != "" {
			// The cursor is appended last so it overrides any starting cursor the
			// caller supplied. A fresh slice keeps the caller's options unmutated.
			pageOpts = append(append([]ListOption(nil), opts...), WithStartingAfter(cursor))
		}

		page, err := s.List(ctx, pageOpts...)
		if err != nil {
			return nil, "", err
		}

		return page.Items, page.NextStartingAfter, nil
	})
}
