// Package workspacemember provides typed access to the Instantly.ai V2 Workspace
// Member API.
//
// It wraps the /api/v2/workspace-members endpoints: inviting, listing, reading,
// patching, and removing the members of the current workspace.
//
//	svc := workspacemember.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, workspacemember.WithAccepted(true))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package workspacemember
