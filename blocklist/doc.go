// Package blocklist provides typed access to the Instantly.ai V2 Block List
// Entry API.
//
// It wraps the /api/v2/block-lists-entries endpoints: creating, listing,
// reading, patching, and deleting block list entries; bulk creating and deleting
// them; deleting every entry; and downloading the whole block list as CSV.
//
//	svc := blocklist.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, blocklist.WithDomainsOnly(true))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package blocklist
