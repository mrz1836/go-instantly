// Package crm provides typed access to the Instantly.ai V2 CRM Actions API.
//
// It wraps the /api/v2/crm-actions endpoints for phone numbers: listing the
// phone numbers an organization owns and deleting one.
//
//	svc := crm.New(instantly.NewClient("[API-KEY]"))
//	numbers, err := svc.ListPhoneNumbers(ctx)
//
// Importing this package pulls in only github.com/mrz1836/go-instantly and the
// standard library.
package crm
