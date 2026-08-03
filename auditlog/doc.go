// Package auditlog provides typed access to the Instantly.ai V2 Audit Log API.
//
// It wraps the /api/v2/audit-logs endpoint, which lists the audit records that
// track activity across a workspace, filtered by activity type, search term, or
// date range.
//
//	svc := auditlog.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, auditlog.WithActivityType(auditlog.ActivityTypeUserLogin))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package auditlog
