// Package workspacebilling provides typed access to the Instantly.ai V2
// Workspace Billing API.
//
// It wraps the /api/v2/workspace-billing endpoints: reading the current
// workspace's plan details and its subscription details.
//
//	svc := workspacebilling.New(instantly.NewClient("[API-KEY]"))
//	plan, err := svc.PlanDetails(ctx)
//
// The plan and subscription payloads are provider-shaped and not documented as a
// fixed schema, so they are carried as json.RawMessage and preserved verbatim.
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package workspacebilling
