// Package leadlist provides typed access to the Instantly.ai V2 Lead List API.
//
// It wraps the /api/v2/lead-lists endpoints: creating, listing, reading,
// patching, and deleting lead lists, plus reading a list's verification stats.
//
//	svc := leadlist.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, leadlist.WithLimit(50))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package leadlist
