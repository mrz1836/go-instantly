// Package accountcampaign provides typed access to the Instantly.ai V2
// account-campaign-mappings endpoint: the campaigns a sending account belongs
// to.
//
//	svc := accountcampaign.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, "sender@example.com", accountcampaign.WithLimit(50))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package accountcampaign
