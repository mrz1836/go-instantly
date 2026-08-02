package subsequence

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
// yielded with a nil *Subsequence, and stops early when the context is canceled.
func (s *Service) ListIter(ctx context.Context, opts ...ListOption) iter.Seq2[*Subsequence, error] {
	return instantly.Iterate(ctx, func(ctx context.Context, cursor string) ([]Subsequence, string, error) {
		pageOpts := opts
		if cursor != "" {
			pageOpts = append(append([]ListOption(nil), opts...), WithStartingAfter(cursor))
		}

		page, err := s.List(ctx, pageOpts...)
		if err != nil {
			return nil, "", err
		}

		return page.Items, page.NextStartingAfter, nil
	})
}
