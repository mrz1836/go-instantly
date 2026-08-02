package instantly

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// widget is a throwaway resource type the generic list helpers are exercised
// against, standing in for a real resource's model.
type widget struct {
	ID string `json:"id"`
}

// widgetOption is a throwaway per-resource option, matching the ~func(*Query)
// shape every real ListOption has.
type widgetOption func(*Query)

// captureTransport records the request it intercepts and answers with a canned
// response, so the result helpers can be asserted without a network round trip.
type captureTransport struct {
	method   string
	path     string
	query    string
	body     string
	status   int
	response string
}

// RoundTrip implements http.RoundTripper.
func (c *captureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	c.method = req.Method
	c.path = req.URL.Path
	c.query = req.URL.RawQuery

	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		c.body = string(body)
	}

	return jsonResponse(c.status, c.response), nil
}

// captureClient points a client at a captureTransport answering with the given
// status and body.
func captureClient(status int, response string) (*Client, *captureTransport) {
	ct := &captureTransport{status: status, response: response}
	client := NewClient(testAPIKey, WithHTTPClient(&http.Client{Transport: ct}))

	return client, ct
}

func TestJoinPath(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		base     string
		segments []string
		want     string
	}{
		"no segments returns the bare base": {base: "/api/v2/x", segments: nil, want: "/api/v2/x"},
		"one segment is appended":           {base: "/api/v2/x", segments: []string{"id-1"}, want: "/api/v2/x/id-1"},
		"several segments join in order": {
			base: "/api/v2/x", segments: []string{"id-1", "verb"}, want: "/api/v2/x/id-1/verb",
		},
		"a slash in a segment is escaped": {
			base: "/api/v2/x", segments: []string{"a/b"}, want: "/api/v2/x/a%2Fb",
		},
		"an email segment is escaped": {
			base: "/api/v2/x", segments: []string{"a b@c.com"}, want: "/api/v2/x/a%20b@c.com",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, JoinPath(tc.base, tc.segments...))
		})
	}
}

func TestApplyOptions(t *testing.T) {
	t.Parallel()

	t.Run("applies each option and skips a nil one", func(t *testing.T) {
		t.Parallel()

		q := ApplyOptions(
			widgetOption(func(q *Query) { q.SetString("limit", "2") }),
			nil,
			widgetOption(func(q *Query) { q.SetString("search", "hello") }),
		)

		require.Equal(t, "2", q.Get("limit"))
		require.Equal(t, "hello", q.Get("search"))
		require.Equal(t, 2, q.Len())
	})

	t.Run("no options yields an empty query", func(t *testing.T) {
		t.Parallel()
		require.Zero(t, ApplyOptions[widgetOption]().Len())
	})
}

func TestResultHelpersRouteEachMethod(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const body = `{"id":"w1"}`

	t.Run("GetResult issues a GET and decodes the body", func(t *testing.T) {
		t.Parallel()
		client, ct := captureClient(http.StatusOK, body)

		got, err := GetResult[widget](ctx, client, "/api/v2/widgets/w1")

		require.NoError(t, err)
		require.Equal(t, &widget{ID: "w1"}, got)
		require.Equal(t, http.MethodGet, ct.method)
		require.Equal(t, "/api/v2/widgets/w1", ct.path)
	})

	t.Run("PostResult issues a POST carrying the payload", func(t *testing.T) {
		t.Parallel()
		client, ct := captureClient(http.StatusOK, body)

		got, err := PostResult[widget](ctx, client, "/api/v2/widgets", widget{ID: "req"})

		require.NoError(t, err)
		require.Equal(t, &widget{ID: "w1"}, got)
		require.Equal(t, http.MethodPost, ct.method)
		require.JSONEq(t, `{"id":"req"}`, ct.body)
	})

	t.Run("PatchResult issues a PATCH carrying the payload", func(t *testing.T) {
		t.Parallel()
		client, ct := captureClient(http.StatusOK, body)

		got, err := PatchResult[widget](ctx, client, "/api/v2/widgets/w1", widget{ID: "req"})

		require.NoError(t, err)
		require.Equal(t, &widget{ID: "w1"}, got)
		require.Equal(t, http.MethodPatch, ct.method)
		require.JSONEq(t, `{"id":"req"}`, ct.body)
	})

	t.Run("DeleteResult issues a DELETE and decodes the body", func(t *testing.T) {
		t.Parallel()
		client, ct := captureClient(http.StatusOK, body)

		got, err := DeleteResult[widget](ctx, client, "/api/v2/widgets/w1")

		require.NoError(t, err)
		require.Equal(t, &widget{ID: "w1"}, got)
		require.Equal(t, http.MethodDelete, ct.method)
	})
}

func TestResultHelperReturnsNilOnError(t *testing.T) {
	t.Parallel()

	client, _ := captureClient(http.StatusNotFound, `{"statusCode":404,"error":"Not Found"}`)

	got, err := GetResult[widget](context.Background(), client, "/api/v2/widgets/missing")

	require.Error(t, err)
	require.Nil(t, got, "a failed request must never hand back a partly populated value")
}

func TestPaginateWalksEveryPageCarryingOptions(t *testing.T) {
	t.Parallel()

	cursors := make([]string, 0, 2)
	limits := make([]string, 0, 2)

	withCursor := func(cursor string) widgetOption {
		return func(q *Query) { q.SetString("starting_after", cursor) }
	}
	list := func(_ context.Context, opts ...widgetOption) (*Page[widget], error) {
		q := ApplyOptions(opts...)
		cursors = append(cursors, q.Get("starting_after"))
		limits = append(limits, q.Get("limit"))

		if q.Get("starting_after") == "" {
			return &Page[widget]{Items: []widget{{ID: "w1"}, {ID: "w2"}}, NextStartingAfter: "c2"}, nil
		}

		return &Page[widget]{Items: []widget{{ID: "w3"}}}, nil
	}

	callerOpts := []widgetOption{func(q *Query) { q.SetString("limit", "2") }}

	seen := make([]string, 0, 3)
	for got, err := range Paginate(context.Background(), callerOpts, withCursor, list) {
		require.NoError(t, err)
		seen = append(seen, got.ID)
	}

	require.Equal(t, []string{"w1", "w2", "w3"}, seen)
	require.Equal(t, []string{"", "c2"}, cursors, "the page cursor overrides the caller's on later pages")
	require.Equal(t, []string{"2", "2"}, limits, "the caller's filters survive every page")
	require.Len(t, callerOpts, 1, "the caller's option slice is never mutated")
}

func TestPaginateStopsOnError(t *testing.T) {
	t.Parallel()

	withCursor := func(cursor string) widgetOption {
		return func(q *Query) { q.SetString("starting_after", cursor) }
	}
	list := func(_ context.Context, opts ...widgetOption) (*Page[widget], error) {
		if ApplyOptions(opts...).Get("starting_after") == "" {
			return &Page[widget]{Items: []widget{{ID: "w1"}}, NextStartingAfter: "c2"}, nil
		}

		return nil, errBoom
	}

	seen := make([]string, 0, 1)
	var iterErr error
	for got, err := range Paginate(context.Background(), []widgetOption(nil), withCursor, list) {
		if err != nil {
			iterErr = err
			require.Nil(t, got)

			break
		}
		seen = append(seen, got.ID)
	}

	require.Equal(t, []string{"w1"}, seen)
	require.ErrorIs(t, iterErr, errBoom)
}

func BenchmarkJoinPath(b *testing.B) {
	for b.Loop() {
		_ = JoinPath("/api/v2/campaigns", "id-1234-5678")
	}
}

func BenchmarkApplyOptions(b *testing.B) {
	opts := []widgetOption{
		func(q *Query) { q.SetInt("limit", 50) },
		func(q *Query) { q.SetString("search", "term") },
	}

	for b.Loop() {
		_ = ApplyOptions(opts...).Path("/api/v2/campaigns")
	}
}
