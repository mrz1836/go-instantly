// Package instantly is the foundation of the Instantly.ai V2 API SDK.
//
// It provides the Client — an authenticated, immutable connection to the API —
// along with the low-level request plumbing (Do, Get, Post, Patch, Put, Delete,
// DoRaw), the shared query builder (Query), the generic pagination helper
// (Iterate), typed API errors (APIError), and cross-resource enums.
//
// Typed access to each API resource lives in its own subpackage — for example
// github.com/mrz1836/go-instantly/email — so importing one resource pulls in
// only that resource plus this foundation and the standard library, never the
// test-support code or the other resources.
//
// Construct a client once and hand it to a resource service:
//
//	client := instantly.NewClient("[API-KEY]")
//	svc := email.New(client)
//	page, err := svc.List(ctx, email.WithLimit(50))
package instantly
