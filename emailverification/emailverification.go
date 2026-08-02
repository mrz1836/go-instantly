package emailverification

import (
	"context"
	"encoding/json"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the Email Verification API.
const basePath = "/api/v2/email-verification"

// Service provides access to the Instantly.ai V2 Email Verification API.
type Service struct {
	client *instantly.Client
}

// New builds an Email Verification API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Status is the verification status of an email address.
//
// It is the field to read when deciding whether an address is deliverable; the
// separate RequestStatus only reports whether the verification request itself
// succeeded.
type Status string

// The verification statuses an email address can be in.
const (
	// StatusPending means the verification has not finished yet, so the result
	// must be polled with Check or awaited on the request's webhook.
	StatusPending Status = "pending"

	// StatusVerified means the address was verified.
	StatusVerified Status = "verified"

	// StatusInvalid means the verification is invalid.
	StatusInvalid Status = "invalid"
)

// RequestStatus reports whether the verification request itself succeeded, which
// is distinct from the verification result carried by Status.
type RequestStatus string

// The request statuses a verification can report.
const (
	// RequestStatusSuccess means the verification request succeeded.
	RequestStatusSuccess RequestStatus = "success"

	// RequestStatusError means the verification request was unsuccessful.
	RequestStatusError RequestStatus = "error"
)

// Verification is the outcome of an email verification.
//
// Fields the API declares as nullable are pointers so an absent value stays
// distinguishable from a zero value.
type Verification struct {
	// Email is the address that was verified.
	Email string `json:"email"`

	// VerificationStatus is the verification status of the address. This is the
	// field to use to determine whether the address is deliverable.
	VerificationStatus Status `json:"verification_status"`

	// RequestStatus reports whether the verification request itself succeeded.
	// It is nil when the API reported nothing for it.
	RequestStatus *RequestStatus `json:"status,omitempty"`

	// CatchAll reports whether the address is a catch-all. The API delivers it as
	// a mixed true/false/"pending" enum, so it is preserved verbatim rather than
	// forced into a single Go type.
	CatchAll json.RawMessage `json:"catch_all,omitempty"`

	// Credits is the number of verification credits available after the
	// verification.
	Credits *float64 `json:"credits,omitempty"`

	// CreditsUsed is the number of verification credits the verification used.
	CreditsUsed *float64 `json:"credits_used,omitempty"`
}

// CreateRequest is the body of a create-email-verification request.
type CreateRequest struct {
	// Email is the address to verify. Required.
	Email string `json:"email"`

	// WebhookURL is a URL that receives the verification result when the
	// verification takes longer than ten seconds, in place of polling with Check.
	WebhookURL string `json:"webhook_url,omitempty"`
}

// Create submits an email address for verification.
//
// A verification that takes longer than ten seconds comes back with a
// VerificationStatus of StatusPending; poll it with Check, or set WebhookURL on
// the request to receive the result instead.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Verification, error) {
	return instantly.PostResult[Verification](ctx, s.client, basePath, req)
}

// Check returns the current verification result for an email address, which is
// how a StatusPending verification is polled to completion.
func (s *Service) Check(ctx context.Context, email string) (*Verification, error) {
	return instantly.GetResult[Verification](ctx, s.client, instantly.JoinPath(basePath, email))
}
