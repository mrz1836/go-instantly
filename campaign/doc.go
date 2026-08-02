// Package campaign provides typed access to the Instantly.ai V2 Campaign API.
//
// It wraps the /api/v2/campaigns endpoints: the full campaign lifecycle (create,
// list, read, patch, delete), activating and pausing, duplicating, sharing,
// exporting and creating from an export, adding variables, reading the sending
// status and launched count, searching campaigns by contact, and the four
// campaign analytics reports (per-campaign, overview, daily, and by step).
//
// Construct a Service from an *instantly.Client and call its methods:
//
//	svc := campaign.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, campaign.WithLimit(50), campaign.WithStatus(campaign.StatusActive))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package campaign
