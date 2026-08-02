// Package email provides typed access to the Instantly.ai V2 Email API.
//
// It wraps the /api/v2/emails endpoints: sending test emails, listing and
// reading emails, patching and deleting them, replying and forwarding, counting
// unread messages, and marking a thread as read.
//
// Construct a Service from an *instantly.Client and call its methods:
//
//	svc := email.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, email.WithLimit(50), email.WithIsUnread(true))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package email
