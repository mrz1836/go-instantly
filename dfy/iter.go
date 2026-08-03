package dfy

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
// yielded with a nil *Order, and stops early when the context is canceled.
func (s *Service) ListIter(ctx context.Context, opts ...ListOption) iter.Seq2[*Order, error] {
	return instantly.Paginate(ctx, opts, WithStartingAfter, s.List)
}

// ListAccountsIter returns an iterator that walks every page of ListAccounts,
// following next_starting_after until the API stops returning a cursor.
//
// The withPasswords choice is bound once and carried onto every page via an
// adapter closure, so ListAccounts's positional argument does not have to become
// a pagination option. Iteration stops at the first error, which is yielded with
// a nil *OrderedAccount, and stops early when the context is canceled.
func (s *Service) ListAccountsIter(
	ctx context.Context, withPasswords bool, opts ...ListOption,
) iter.Seq2[*OrderedAccount, error] {
	list := func(ctx context.Context, o ...ListOption) (*AccountsResponse, error) {
		return s.ListAccounts(ctx, withPasswords, o...)
	}

	return instantly.Paginate(ctx, opts, WithStartingAfter, list)
}
