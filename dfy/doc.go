// Package dfy provides typed access to the Instantly.ai V2 Done-For-You (DFY)
// Email Account Order API.
//
// It wraps the /api/v2/dfy-email-account-orders endpoints: placing an order for
// new mailboxes (or simulating one for a price quote), listing the orders and
// the accounts they produced, generating and checking candidate domains,
// listing pre-warmed-up domains, and canceling accounts.
//
//	svc := dfy.New(instantly.NewClient("[API-KEY]"))
//	result, err := svc.Create(ctx, dfy.CreateRequest{
//		OrderType:  dfy.OrderTypeDFY,
//		Simulation: instantly.Ptr(true),
//		Items:      []dfy.OrderItem{{Domain: "example.com"}},
//	})
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package dfy
