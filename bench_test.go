package instantly

import (
	"context"
	"net/http"
	"testing"
)

// BenchmarkQueryEncode measures encoding an accumulated query to its wire form,
// the last step of every list request path.
func BenchmarkQueryEncode(b *testing.B) {
	q := NewQuery()
	q.SetInt("limit", 50)
	q.SetString("search", "quick question")
	q.SetString("starting_after", "cursor-abc-123")

	for b.Loop() {
		_ = q.Encode()
	}
}

// BenchmarkBuildPath measures appending encoded parameters onto a base path.
func BenchmarkBuildPath(b *testing.B) {
	q := NewQuery()
	q.SetInt("limit", 50)
	q.SetString("search", "term")

	for b.Loop() {
		_ = q.Path("/api/v2/campaigns")
	}
}

// BenchmarkCheckResponse measures the success-body error probe every response
// runs through before it is decoded.
func BenchmarkCheckResponse(b *testing.B) {
	body := []byte(`{"status":"success"}`)

	for b.Loop() {
		if err := checkResponse(http.StatusOK, body); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkIterate measures walking a two-page cursor sequence to exhaustion.
func BenchmarkIterate(b *testing.B) {
	pages := []page{
		{items: []int{1, 2, 3, 4, 5}, next: "c2"},
		{items: []int{6, 7, 8, 9, 10}, next: ""},
	}

	for b.Loop() {
		calls := 0
		for _, err := range Iterate(context.Background(), fetcherFrom(pages, &calls)) {
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
