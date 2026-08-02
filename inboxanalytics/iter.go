package inboxanalytics

import (
	"context"
	"iter"

	"github.com/mrz1836/go-instantly"
)

// ListIter returns an iterator that walks every page of List for a test,
// following next_starting_after until the API stops returning a cursor.
//
// It is strictly additive: List still returns one page for callers who want to
// drive pagination themselves. The required test_id is bound once and carried
// onto every page. Iteration stops at the first error, which is yielded with a
// nil *Analytics, and stops early when the context is canceled.
func (s *Service) ListIter(ctx context.Context, testID string, opts ...ListOption) iter.Seq2[*Analytics, error] {
	list := func(ctx context.Context, o ...ListOption) (*ListResponse, error) {
		return s.List(ctx, testID, o...)
	}

	return instantly.Paginate(ctx, opts, WithStartingAfter, list)
}
