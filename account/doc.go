// Package account provides typed access to the Instantly.ai V2 Account API.
//
// It wraps the /api/v2/accounts endpoints: creating, listing, reading, patching,
// and deleting sending accounts; pausing (one at a time or in bulk), resuming,
// and marking accounts fixed; enabling and disabling warmup; moving accounts
// between workspaces; testing account vitals; and reading warmup, daily, and
// custom-tracking-domain analytics.
//
// Construct a Service from an *instantly.Client and call its methods:
//
//	svc := account.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, account.WithLimit(50), account.WithStatus(account.StatusActive))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package account
