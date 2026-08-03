// Package apikey provides typed access to the Instantly.ai V2 API Key API.
//
// It wraps the /api/v2/api-keys endpoints: creating an API key with a set of
// scopes, listing the keys in a workspace, and deleting one.
//
//	svc := apikey.New(instantly.NewClient("[API-KEY]"))
//	key, err := svc.Create(ctx, apikey.CreateRequest{
//		Name:   "CI deploy key",
//		Scopes: []apikey.Scope{apikey.ScopeCampaignsRead, apikey.ScopeEmailsRead},
//	})
//
// Scopes are enumerated as named constants (see scopes.go) for autocomplete and
// compile-time safety, while remaining a plain string type so a scope the API
// adds later still works.
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package apikey
