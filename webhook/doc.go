// Package webhook provides typed access to the Instantly.ai V2 Webhook API.
//
// It wraps the /api/v2/webhooks endpoints: creating, listing, reading, patching,
// and deleting webhooks; listing the available event types; sending a test
// delivery; and resuming a webhook disabled by delivery failures.
//
//	svc := webhook.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, webhook.WithEventType(webhook.EventReplyReceived))
//
// The custom headers a webhook sends are a free-form map, so they are carried as
// json.RawMessage and preserved verbatim.
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package webhook
