// Package inboxreport provides typed access to the Instantly.ai V2 Inbox
// Placement blacklist and SpamAssassin report API.
//
// It wraps the /api/v2/inbox-placement-reports endpoints: listing and reading
// the blacklist and SpamAssassin reports produced for an inbox placement test.
//
//	svc := inboxreport.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, testID, inboxreport.WithSkipSpamAssassinReport(true))
//
// List and its ListIter require a test_id filter, so it is a required positional
// argument rather than an option. The nested blacklist and SpamAssassin reports
// are preserved verbatim as json.RawMessage — the SpamAssassin per-rule score
// arrives as a string, which raw preservation keeps intact.
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package inboxreport
