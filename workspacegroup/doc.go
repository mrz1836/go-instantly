// Package workspacegroup provides typed access to the Instantly.ai V2 Workspace
// Group Member API.
//
// It wraps the /api/v2/workspace-group-members endpoints: inviting a sub
// workspace into the current workspace's group, listing and reading group
// members, removing them, and reading the group's admin workspace.
//
//	svc := workspacegroup.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, workspacegroup.WithLimit(50))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package workspacegroup
