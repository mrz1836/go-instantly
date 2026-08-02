// Package inboxanalytics provides typed access to the Instantly.ai V2 Inbox
// Placement Analytics API.
//
// It wraps the /api/v2/inbox-placement-analytics endpoints: listing and reading
// individual placement events for a test, and three aggregate stats endpoints —
// by test id, by date, and deliverability insights.
//
//	svc := inboxanalytics.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, testID, inboxanalytics.WithDateFrom("2026-08-01"))
//
// List and its ListIter require a test_id filter, so it is a required positional
// argument rather than an option. The three stats endpoints POST typed enum
// slices and return bare arrays.
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package inboxanalytics
