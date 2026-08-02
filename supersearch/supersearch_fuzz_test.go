package supersearch_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly"
	"github.com/mrz1836/go-instantly/internal/instantlytest"
	"github.com/mrz1836/go-instantly/supersearch"
)

// FuzzSuperSearchSerialization round trips arbitrary field values through the
// request bodies, asserting the encoding never panics and never drifts. The
// free-form raw payloads are held at a fixed valid value so only the typed
// fields vary.
func FuzzSuperSearchSerialization(f *testing.F) {
	f.Add("res-1", "ai_summary", "gpt-4o", "My List", 1)
	f.Add("", "", "", "", 0)
	f.Add("Ünïcödé", "col\r\n", "\x00model", "list", 2)

	f.Fuzz(func(t *testing.T, resourceID, column, model, listName string, kind int) {
		lossless := utf8.ValidString(resourceID) && utf8.ValidString(column) &&
			utf8.ValidString(model) && utf8.ValidString(listName)

		filters := json.RawMessage(`{"q":"x"}`)

		instantlytest.RequireStableRoundTrip(t, supersearch.CreateRequest{
			ResourceID: resourceID,
			Type:       supersearch.EnrichmentType(model),
			Limit:      instantly.Ptr(float64(kind)),
			CustomFlow: []string{listName},
			Filters:    filters,
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, supersearch.AIRequest{
			ResourceID:   resourceID,
			OutputColumn: column,
			ResourceType: supersearch.ResourceType(kind%2 + 1),
			ModelVersion: supersearch.ModelVersion(model),
			Prompt:       listName,
			Status:       instantly.Ptr(supersearch.AIStatus(kind%4 + 1)),
			Filters:      filters,
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, supersearch.EnrichLeadsRequest{
			SearchFilters:    filters,
			Limit:            float64(kind),
			ListName:         listName,
			ResourceID:       resourceID,
			CustomFlow:       []string{column},
			AIEnrichment:     json.RawMessage(`{"model":"x"}`),
			SignalEnrichment: json.RawMessage(`[]`),
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, supersearch.RunRequest{
			ResourceID:  resourceID,
			ColumnName:  column,
			Count:       instantly.Ptr(int64(kind)),
			StartingRow: instantly.Ptr(int64(kind)),
			LeadIDs:     []string{resourceID},
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, supersearch.FacetRequest{
			Category: column,
			Field:    listName,
			Prefix:   model,
		}, lossless)

		instantlytest.RequireStableRoundTrip(t, supersearch.SettingsRequest{
			AutoUpdate:  instantly.Ptr(kind%2 == 0),
			IsEvergreen: instantly.Ptr(kind%2 == 1),
		}, lossless)
	})
}

// FuzzSuperSearchResponseDecoding feeds arbitrary bytes back as supersearch
// responses, asserting the client never panics and never hands back a partly
// decoded value.
func FuzzSuperSearchResponseDecoding(f *testing.F) {
	f.Add(createFixture)
	f.Add(enrichmentFixture)
	f.Add(resourceFixture)
	f.Add(resourceFixtureNulls)
	f.Add(aiFixture)
	f.Add(enrichLeadsFixture)
	f.Add(`[` + aiInProgressFixture + `]`)
	f.Add(`{"number_of_leads":1234}`)
	f.Add(`{"keywords":[{"keyword":"ai","count":42}]}`)
	f.Add(`{`)
	f.Add(``)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{"resource_type":"not a number"}`)
	f.Add(`{"status":2.5}`)

	f.Fuzz(func(t *testing.T, body string) {
		ctx := context.Background()
		svc := supersearch.New(instantlytest.FuzzClient(http.StatusOK, body))

		requireNilOnError(t, func() (any, error) { return svc.Get(ctx, resourceID) })
		requireNilOnError(t, func() (any, error) { return svc.Create(ctx, supersearch.CreateRequest{}) })
		requireNilOnError(t, func() (any, error) { return svc.Run(ctx, supersearch.RunRequest{}) })
		requireNilOnError(t, func() (any, error) { return svc.CreateAI(ctx, supersearch.AIRequest{}) })
		requireNilOnError(t, func() (any, error) { return svc.CountLeads(ctx, supersearch.SearchRequest{}) })
		requireNilOnError(t, func() (any, error) { return svc.PreviewLeads(ctx, supersearch.SearchRequest{}) })
		requireNilOnError(t, func() (any, error) { return svc.EnrichLeads(ctx, supersearch.EnrichLeadsRequest{}) })
		requireNilOnError(t, func() (any, error) {
			return svc.UpdateSettings(ctx, resourceID, supersearch.SettingsRequest{})
		})

		require.NotPanics(t, func() {
			_, _ = svc.AIInProgress(ctx, resourceID)
			_, _ = svc.SignalKeywords(ctx, supersearch.FacetRequest{})
			_, _ = svc.History(ctx, resourceID)
		})
	})
}

// requireNilOnError asserts a single-resource call hands back a nil pointer
// whenever it reports an error, so a decode failure never yields a partly
// populated value. It works over any *T return via a typed-nil reflection check.
func requireNilOnError(t *testing.T, call func() (any, error)) {
	t.Helper()

	got, err := call()
	if err != nil {
		require.True(t, isNilPointer(got), "a decode failure must never hand back a partly populated value")
	} else {
		require.False(t, isNilPointer(got), "a successful decode must hand back a value")
	}
}

// isNilPointer reports whether v is a nil pointer, so the fuzz check can span the
// several distinct *T return types with one helper.
func isNilPointer(v any) bool {
	switch p := v.(type) {
	case *supersearch.Enrichment:
		return p == nil
	case *supersearch.ResourceEnrichment:
		return p == nil
	case *supersearch.AIEnrichment:
		return p == nil
	case *supersearch.EnrichLeadsResponse:
		return p == nil
	case *supersearch.LeadCount:
		return p == nil
	case *supersearch.Preview:
		return p == nil
	default:
		return false
	}
}
