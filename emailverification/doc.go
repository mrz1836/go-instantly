// Package emailverification provides typed access to the Instantly.ai V2 Email
// Verification API.
//
// It wraps the /api/v2/email-verification endpoints: submitting an email address
// for verification and checking the status of a verification that is still
// pending.
//
//	svc := emailverification.New(instantly.NewClient("[API-KEY]"))
//	result, err := svc.Create(ctx, emailverification.CreateRequest{Email: "example@example.com"})
//
// Verification can be asynchronous: an address that takes longer than ten
// seconds to verify comes back with a VerificationStatus of StatusPending, at
// which point Check polls the result, or a webhook_url on the request receives
// it instead.
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package emailverification
