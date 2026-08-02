package leadlist

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
// yielded with a nil *LeadList, and stops early when the context is canceled.
func (s *Service) ListIter(ctx context.Context, opts ...ListOption) iter.Seq2[*LeadList, error] {
	return instantly.Paginate(ctx, opts, WithStartingAfter, s.List)
}
