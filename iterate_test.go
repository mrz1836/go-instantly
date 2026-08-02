package instantly

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// errBoom is a sentinel error a fetch page can fail with.
var errBoom = errors.New("boom")

// page is one canned page a fake fetcher hands back.
type page struct {
	items []int
	next  string
	err   error
}

// fetcherFrom returns a FetchPage backed by canned pages, recording how many
// times it was called through calls.
func fetcherFrom(pages []page, calls *int) FetchPage[int] {
	return func(_ context.Context, _ string) ([]int, string, error) {
		i := *calls
		*calls++

		if i >= len(pages) {
			return nil, "", nil
		}

		return pages[i].items, pages[i].next, pages[i].err
	}
}

// collect drains an iterator into the values it yielded and the first error.
func collect(seq func(func(*int, error) bool)) ([]int, error) {
	var got []int
	var iterErr error

	for v, err := range seq {
		if err != nil {
			iterErr = err
			break
		}
		got = append(got, *v)
	}

	return got, iterErr
}

func TestIterateWalksEveryPage(t *testing.T) {
	t.Parallel()

	calls := 0
	pages := []page{
		{items: []int{1, 2}, next: "c2"},
		{items: []int{3, 4}, next: "c3"},
		{items: []int{5}, next: ""},
	}

	got, err := collect(Iterate(t.Context(), fetcherFrom(pages, &calls)))

	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3, 4, 5}, got)
	require.Equal(t, 3, calls, "one call per page, none after the empty cursor")
}

func TestIterateStopsOnError(t *testing.T) {
	t.Parallel()

	calls := 0
	pages := []page{
		{items: []int{1}, next: "c2"},
		{err: errBoom},
	}

	got, err := collect(Iterate(t.Context(), fetcherFrom(pages, &calls)))

	require.Equal(t, []int{1}, got)
	require.ErrorIs(t, err, errBoom)
	require.Equal(t, 2, calls, "iteration ends at the failing page")
}

func TestIterateStopsOnEmptyPageWithCursor(t *testing.T) {
	t.Parallel()

	calls := 0
	pages := []page{
		{items: []int{1}, next: "c2"},
		{items: []int{}, next: "c3"}, // empty page still advertising a cursor
	}

	got, err := collect(Iterate(t.Context(), fetcherFrom(pages, &calls)))

	require.NoError(t, err)
	require.Equal(t, []int{1}, got)
	require.Equal(t, 2, calls, "an empty page ends iteration even when a cursor is offered")
}

func TestIterateEmptyFirstPage(t *testing.T) {
	t.Parallel()

	calls := 0
	pages := []page{{items: []int{}, next: "c2"}}

	got, err := collect(Iterate(t.Context(), fetcherFrom(pages, &calls)))

	require.NoError(t, err)
	require.Empty(t, got)
	require.Equal(t, 1, calls)
}

func TestIterateStopsOnBreak(t *testing.T) {
	t.Parallel()

	calls := 0
	pages := []page{{items: []int{1, 2, 3}, next: "c2"}}

	var got []int
	for v, err := range Iterate(t.Context(), fetcherFrom(pages, &calls)) {
		require.NoError(t, err)
		got = append(got, *v)
		if len(got) == 2 {
			break
		}
	}

	require.Equal(t, []int{1, 2}, got)
	require.Equal(t, 1, calls, "breaking mid-page fetches no further page")
}

func TestIterateStopsOnCancellationBeforeFirstPage(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	pages := []page{{items: []int{1}, next: "c2"}}

	got, err := collect(Iterate(ctx, fetcherFrom(pages, &calls)))

	require.Empty(t, got)
	require.ErrorIs(t, err, context.Canceled)
	require.Zero(t, calls, "a canceled context issues no request at all")
}

func TestIterateStopsOnCancellationBetweenPages(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	calls := 0
	pages := []page{
		{items: []int{1}, next: "c2"},
		{items: []int{2}, next: "c3"},
	}

	var got []int
	var iterErr error
	for v, err := range Iterate(ctx, fetcherFrom(pages, &calls)) {
		if err != nil {
			iterErr = err
			break
		}
		got = append(got, *v)
		cancel() // cancel between pages while a cursor is still pending
	}

	require.Equal(t, []int{1}, got)
	require.ErrorIs(t, iterErr, context.Canceled)
	require.Equal(t, 1, calls, "no page is requested after cancellation")
}

// FuzzIterate feeds arbitrary page shapes through Iterate, asserting it never
// panics, never yields past the first empty page, and stays bounded.
func FuzzIterate(f *testing.F) {
	f.Add(3, 2, true)
	f.Add(0, 5, false)
	f.Add(10, 0, true)
	f.Add(255, 1, false)

	f.Fuzz(func(t *testing.T, pageCount, pageSize int, endWithCursor bool) {
		// Bound the fuzzed shape so the test stays fast and cannot allocate wildly.
		pageCount = ((pageCount % 32) + 32) % 32
		pageSize = ((pageSize % 8) + 8) % 8

		pages := make([]page, 0, pageCount)
		for i := range pageCount {
			items := make([]int, pageSize)
			for j := range items {
				items[j] = i*pageSize + j
			}
			// The last page drops its cursor unless the shape says otherwise.
			next := "cursor"
			if i == pageCount-1 && !endWithCursor {
				next = ""
			}
			pages = append(pages, page{items: items, next: next})
		}

		calls := 0
		require.NotPanics(t, func() {
			got, err := collect(Iterate(t.Context(), fetcherFrom(pages, &calls)))
			require.NoError(t, err)

			// An empty page (pageSize 0, or running past the canned pages) halts
			// iteration, so the yield count can never exceed what the non-empty
			// prefix of pages holds.
			require.LessOrEqual(t, len(got), pageCount*pageSize)
			require.LessOrEqual(t, calls, pageCount+1)
		})
	})
}
