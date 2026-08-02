// Package webhookevent provides typed access to the Instantly.ai V2 Webhook
// Event API.
//
// It wraps the /api/v2/webhook-events endpoints: listing and reading webhook
// delivery events, and two aggregate views — an overall summary and a summary
// broken down by date.
//
//	svc := webhookevent.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, webhookevent.WithSuccess(false))
//
// The delivered event payload is free-form, so it is carried as json.RawMessage
// and preserved verbatim.
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package webhookevent
