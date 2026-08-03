package oauth

import (
	"context"
	"time"

	"github.com/mrz1836/go-instantly"
)

// basePath is the root path of the OAuth API.
const basePath = "/api/v2/oauth"

// Service provides access to the Instantly.ai V2 OAuth API.
type Service struct {
	client *instantly.Client
}

// New builds an OAuth API service from an Instantly client.
func New(client *instantly.Client) *Service {
	return &Service{client: client}
}

// Status is the state an OAuth session is in.
type Status string

// The states an OAuth session can be in.
const (
	// StatusPending means the user has not yet finished the consent flow.
	StatusPending Status = "pending"

	// StatusSuccess means the account was connected successfully.
	StatusSuccess Status = "success"

	// StatusError means the session ended in an error. The API delivers this
	// status alongside a top-level error code, which the client surfaces as an
	// *instantly.APIError rather than a decoded status; see SessionStatus.
	StatusError Status = "error"

	// StatusExpired means the session expired before the user finished.
	StatusExpired Status = "expired"
)

// InitResult is the outcome of starting an OAuth session.
type InitResult struct {
	// AuthURL is the provider authorization URL to redirect the user to.
	AuthURL string `json:"auth_url"`

	// SessionID identifies the session, used to poll its status.
	SessionID string `json:"session_id"`

	// ExpiresAt is when the session expires, ten minutes after creation.
	ExpiresAt string `json:"expires_at"`
}

// ParsedExpiresAt parses ExpiresAt as an RFC 3339 time.
//
// The raw string field is left untouched so a decoded result re-encodes
// byte-for-byte; call this accessor when a time.Time is needed.
func (r *InitResult) ParsedExpiresAt() (time.Time, error) {
	return time.Parse(time.RFC3339, r.ExpiresAt)
}

// SessionStatus is the current state of an OAuth session.
//
// Email and Name are populated on success. Error and ErrorDescription are part
// of the documented schema but are never populated through SessionStatus: a
// session that ends in an error is reported as an *instantly.APIError instead,
// because the API delivers the error code inside an HTTP 200 body.
type SessionStatus struct {
	// Status is the current state of the session.
	Status Status `json:"status"`

	// Email is the address of the connected account, on success.
	Email string `json:"email,omitempty"`

	// Name is the name of the account owner, on success.
	Name string `json:"name,omitempty"`

	// Error is the OAuth error code, on error.
	Error string `json:"error,omitempty"`

	// ErrorDescription is the human-readable error detail, on error.
	ErrorDescription string `json:"error_description,omitempty"`
}

// InitGoogle starts a Google OAuth session and returns the authorization URL to
// redirect the user to, along with the session id used to poll its status.
//
// The request carries no body.
func (s *Service) InitGoogle(ctx context.Context) (*InitResult, error) {
	return instantly.PostResult[InitResult](ctx, s.client, basePath+"/google/init", nil)
}

// InitMicrosoft starts a Microsoft OAuth session and returns the authorization
// URL to redirect the user to, along with the session id used to poll its
// status.
//
// The request carries no body.
func (s *Service) InitMicrosoft(ctx context.Context) (*InitResult, error) {
	return instantly.PostResult[InitResult](ctx, s.client, basePath+"/microsoft/init", nil)
}

// SessionStatus polls the status of an OAuth session by its id.
//
// A pending, success, or expired session decodes into a SessionStatus. A
// session that ends in an error is instead reported as an *instantly.APIError,
// because the API delivers the OAuth error code inside an otherwise successful
// HTTP 200 body and the client converts that into an error. The error code
// (for example access_denied) is available on the APIError; its human-readable
// description is not surfaced:
//
//	status, err := svc.SessionStatus(ctx, sessionID)
//	if err != nil {
//		var apiErr *instantly.APIError
//		if errors.As(err, &apiErr) {
//			// apiErr.Code is the OAuth error code, apiErr.StatusCode is 200
//		}
//		return err
//	}
func (s *Service) SessionStatus(ctx context.Context, sessionID string) (*SessionStatus, error) {
	path := instantly.JoinPath(basePath, "session", "status", sessionID)

	return instantly.GetResult[SessionStatus](ctx, s.client, path)
}
