// Package inboxtest provides typed access to the Instantly.ai V2 Inbox Placement
// Test API.
//
// It wraps the /api/v2/inbox-placement-tests endpoints: creating, listing,
// reading, patching, and deleting inbox placement tests, and reading the email
// service provider options a test can target.
//
//	svc := inboxtest.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, inboxtest.WithStatus(inboxtest.StatusActive))
//
// A test reads back with extra metadata when Get is called with WithMetadata.
// The nested schedule and automation payloads are preserved verbatim as
// json.RawMessage so no detail is lost.
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package inboxtest
