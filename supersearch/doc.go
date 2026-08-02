// Package supersearch provides typed access to the Instantly.ai V2 SuperSearch
// Enrichment API.
//
// It wraps the /api/v2/supersearch-enrichment endpoints: creating standard and
// AI enrichments, reading an enrichment for a resource and its in-progress AI
// jobs, updating enrichment settings, running an enrichment, and the SuperSearch
// lead workflow — counting, previewing, and enriching leads from a SuperSearch
// query, plus faceting signal keywords and reading enrichment history.
//
//	svc := supersearch.New(instantly.NewClient("[API-KEY]"))
//	count, err := svc.CountLeads(ctx, supersearch.SearchRequest{SearchFilters: filters})
//
// The SuperSearch query DSL (search_filters, filters, integration_actions, and
// the enrichment payloads) is deeply nested and free-form, so those fields are
// carried as json.RawMessage and preserved verbatim; the flat, fully-documented
// shapes — leads, keywords, and the enrichment records — are typed.
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package supersearch
