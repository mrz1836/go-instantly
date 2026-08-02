// Package customtag provides typed access to the Instantly.ai V2 Custom Tag API.
//
// It wraps the /api/v2/custom-tags endpoints: creating, listing, reading,
// patching, and deleting custom tags, and assigning or unassigning tags to
// resources in bulk.
//
//	svc := customtag.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, customtag.WithSearch("vip"))
//
// The optional selected-all filter on a toggle request is a free-form payload,
// so it is carried as json.RawMessage and preserved verbatim.
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package customtag
