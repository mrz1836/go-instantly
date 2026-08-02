// Package workspace provides typed access to the Instantly.ai V2 Workspace API.
//
// It wraps the /api/v2/workspaces/current endpoints, which all operate on the
// single workspace the API key authenticates against — there is no workspace id
// argument. It covers reading and patching the workspace, scheduling and
// canceling its removal, managing its whitelabel agency domain, and changing its
// owner.
//
//	svc := workspace.New(instantly.NewClient("[API-KEY]"))
//	ws, err := svc.Get(ctx)
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package workspace
