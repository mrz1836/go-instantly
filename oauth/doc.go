// Package oauth provides typed access to the Instantly.ai V2 OAuth API.
//
// It wraps the /api/v2/oauth endpoints for connecting a mailbox: starting a
// Google or Microsoft OAuth session, and polling that session's status until the
// user finishes (or abandons) the consent flow.
//
//	svc := oauth.New(instantly.NewClient("[API-KEY]"))
//	session, err := svc.InitGoogle(ctx)
//	status, err := svc.SessionStatus(ctx, session.SessionID)
//
// A session that ends in an error is reported as an *instantly.APIError rather
// than a decoded status: the API delivers the OAuth error code inside an
// otherwise successful HTTP 200 body, and the client converts that into an
// error. See SessionStatus for details.
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package oauth
