// Package lead provides typed access to the Instantly.ai V2 Lead API.
//
// It wraps the /api/v2/leads endpoints: creating, reading, patching, and
// deleting leads; listing them (via POST /leads/list, whose filters and cursor
// travel in the request body); bulk add, delete, assign, move, and merge; and
// updating interest status and subsequence membership.
//
//	svc := lead.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, lead.ListRequest{Campaign: "campaign-id", Limit: 50})
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package lead
