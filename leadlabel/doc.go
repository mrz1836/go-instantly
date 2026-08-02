// Package leadlabel provides typed access to the Instantly.ai V2 Lead Label API.
//
// It wraps the /api/v2/lead-labels endpoints: creating, listing, reading,
// patching, and deleting lead labels, plus testing which label an AI reply
// classifier would assign to a reply.
//
//	svc := leadlabel.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, leadlabel.WithLimit(50))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package leadlabel
