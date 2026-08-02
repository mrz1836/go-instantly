package accountcampaign

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
// yielded with a nil *Mapping, and stops early when the context is canceled. The
// account email is fixed for the whole walk, so List is wrapped in a closure
// that captures it before being handed to the shared pagination helper.
func (s *Service) ListIter(
	ctx context.Context, email string, opts ...ListOption,
) iter.Seq2[*Mapping, error] {
	list := func(ctx context.Context, pageOpts ...ListOption) (*instantly.Page[Mapping], error) {
		return s.List(ctx, email, pageOpts...)
	}

	return instantly.Paginate(ctx, opts, WithStartingAfter, list)
}
