package lead

import (
	"context"
	"iter"

	"github.com/mrz1836/go-instantly"
)

// ListIter returns an iterator that walks every page of List, following
// next_starting_after until the API stops returning a cursor.
//
// Because listing leads is a POST, the pagination cursor travels in the request
// body: each page copies the caller's request and overrides StartingAfter with
// the next cursor, so the caller's request value is never mutated. Iteration
// stops at the first error, which is yielded with a nil *Lead, and stops early
// when the context is canceled.
func (s *Service) ListIter(ctx context.Context, req ListRequest) iter.Seq2[*Lead, error] {
	return instantly.Iterate(ctx, func(ctx context.Context, cursor string) ([]Lead, string, error) {
		pageReq := req
		if cursor != "" {
			pageReq.StartingAfter = cursor
		}

		page, err := s.List(ctx, pageReq)
		if err != nil {
			return nil, "", err
		}

		return page.Items, page.NextStartingAfter, nil
	})
}
