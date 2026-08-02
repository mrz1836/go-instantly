// Package subsequence provides typed access to the Instantly.ai V2 Campaign
// Subsequence API.
//
// It wraps the /api/v2/subsequences endpoints: creating, listing, reading,
// patching, and deleting campaign subsequences; duplicating them; pausing and
// resuming; and reading a subsequence's sending status.
//
//	svc := subsequence.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, subsequence.WithParentCampaign("campaign-id"))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package subsequence
