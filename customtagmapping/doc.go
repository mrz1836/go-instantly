// Package customtagmapping provides typed access to the Instantly.ai V2 Custom
// Tag Mapping API.
//
// It wraps the /api/v2/custom-tag-mappings endpoint, which lists the mappings
// between custom tags and the resources they are assigned to.
//
//	svc := customtagmapping.New(instantly.NewClient("[API-KEY]"))
//	page, err := svc.List(ctx, customtagmapping.WithResourceIDs("res-1,res-2"))
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package customtagmapping
