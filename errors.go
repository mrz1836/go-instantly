package instantly

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Error codes the API delivers inside an HTTP 200 response body.
const (
	// ErrCodeAccountAuthError indicates the sending account failed to authenticate.
	ErrCodeAccountAuthError = "ACC_AUTH_ERROR"

	// ErrCodeAccountNotFound indicates the requested sending account does not exist.
	ErrCodeAccountNotFound = "ACC_NOT_FOUND"

	// ErrCodeAccountUnknownError indicates an unclassified sending account failure.
	ErrCodeAccountUnknownError = "ACC_UNKNOWN_ERROR"
)

// APIError represents an error returned by the Instantly.ai API.
//
// It covers both wire shapes the API uses: a 4xx envelope carrying
// {statusCode, error, message}, and an error code delivered inside an HTTP 200
// body as {"error": "ACC_AUTH_ERROR"}. Both are returned as an error, so
// err != nil catches every real failure.
//
// Callers can inspect the code with errors.As:
//
//	var apiErr *instantly.APIError
//	if errors.As(err, &apiErr) && apiErr.Code == instantly.ErrCodeAccountAuthError {
//		// the sending account failed to authenticate
//	}
type APIError struct {
	// StatusCode is the HTTP status code of the response. It is 200 when the
	// error arrived inside an otherwise successful response.
	StatusCode int64 `json:"statusCode,omitempty"`

	// Code is the machine-readable error code, such as ACC_AUTH_ERROR for an
	// error delivered at HTTP 200, or Unauthorized for a 4xx envelope.
	Code string `json:"error,omitempty"`

	// Message is the human-readable error detail, when the API provides one.
	Message string `json:"message,omitempty"`
}

// Error returns a readable description of the API error for both wire shapes.
func (e *APIError) Error() string {
	parts := make([]string, 0, 3)

	// A 200 is not an HTTP failure, so reporting it alongside the code is noise.
	if e.StatusCode > 0 && e.StatusCode != http.StatusOK {
		parts = append(parts, fmt.Sprintf("status %d", e.StatusCode))
	}
	if e.Code != "" {
		parts = append(parts, e.Code)
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}

	if len(parts) == 0 {
		return "instantly: unknown api error"
	}

	return "instantly: " + strings.Join(parts, ": ")
}

// checkResponse converts an API failure into an error, whichever wire shape it
// arrived in. It returns nil when the response is a genuine success.
func checkResponse(statusCode int, body []byte) error {
	if statusCode >= http.StatusBadRequest {
		apiErr := &APIError{}
		if err := json.Unmarshal(body, apiErr); err != nil {
			// The body is not the documented envelope, but the request still
			// failed: never let a non-JSON error body become a nil error.
			return fmt.Errorf("instantly: request failed with status %d: %w", statusCode, err)
		}

		// Fall back to the transport status when the body omits it.
		if apiErr.StatusCode == 0 {
			apiErr.StatusCode = int64(statusCode)
		}

		return apiErr
	}

	return successBodyError(statusCode, body)
}

// successBodyError probes a successful response for an error code delivered
// inside the body, which some endpoints return instead of a 4xx status.
func successBodyError(statusCode int, body []byte) error {
	var probe struct {
		Error string `json:"error"`
	}

	// A body that is not a JSON object carries no error field to find, and
	// decoding it properly is the caller's job, so a probe failure is not itself
	// an error here.
	if err := json.Unmarshal(body, &probe); err == nil && probe.Error != "" {
		return &APIError{
			StatusCode: int64(statusCode),
			Code:       probe.Error,
		}
	}

	return nil
}
