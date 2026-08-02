package instantly

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// TestAPIErrorStatusEnvelope verifies every documented 4xx envelope decodes into
// an APIError carrying all three fields.
func (s *clientTestSuite) TestAPIErrorStatusEnvelope() {
	tests := []struct {
		name       string
		statusCode int
		code       string
		message    string
	}{
		{"unauthorized", http.StatusUnauthorized, "Unauthorized", "Missing Authorization header"},
		{"payment required", http.StatusPaymentRequired, "Payment Required", "Workspace plan limit reached"},
		{"not found", http.StatusNotFound, "Not Found", "Email not found"},
		{"too many requests", http.StatusTooManyRequests, "Too Many Requests", "Rate limit exceeded"},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			path := "/api/v2/status-errors/" + strconv.Itoa(test.statusCode)
			body := fmt.Sprintf(
				`{"statusCode":%d,"error":%q,"message":%q}`, test.statusCode, test.code, test.message,
			)

			s.handle(http.MethodGet, path, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.statusCode)
				_, _ = w.Write([]byte(body))
			})

			var result map[string]string
			err := s.client.Get(context.Background(), path, &result)

			s.Require().Error(err)

			var apiErr *APIError
			s.Require().ErrorAs(err, &apiErr)
			s.Equal(int64(test.statusCode), apiErr.StatusCode)
			s.Equal(test.code, apiErr.Code)
			s.Equal(test.message, apiErr.Message)
			s.Empty(result, "the destination must not be populated from an error response")
		})
	}
}

// TestAPIErrorStatusCodeFallback verifies the transport status is used when the
// error body omits it.
func (s *clientTestSuite) TestAPIErrorStatusCodeFallback() {
	s.handle(http.MethodGet, "/api/v2/status-fallback", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"Not Found","message":"no statusCode in this body"}`))
	})

	err := s.client.Get(context.Background(), "/api/v2/status-fallback", nil)

	s.Require().Error(err)

	var apiErr *APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(int64(http.StatusNotFound), apiErr.StatusCode)
	s.Equal("Not Found", apiErr.Code)
}

// TestAPIErrorNonJSONBody verifies a failure whose body is not the documented
// envelope still returns an error naming the status code, never a nil error.
func (s *clientTestSuite) TestAPIErrorNonJSONBody() {
	s.handle(http.MethodGet, "/api/v2/html-error", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`<html><body>502 Bad Gateway</body></html>`))
	})

	var result map[string]string
	err := s.client.Get(context.Background(), "/api/v2/html-error", &result)

	s.Require().Error(err)
	s.Require().ErrorContains(err, "request failed with status 502")

	var apiErr *APIError
	s.Require().NotErrorAs(err, &apiErr, "a non-envelope body cannot masquerade as an APIError")
	s.Empty(result)
}

// TestAPIErrorInSuccessBody verifies an error code delivered inside an HTTP 200
// body becomes a returned error and leaves the destination untouched.
func (s *clientTestSuite) TestAPIErrorInSuccessBody() {
	codes := []string{
		ErrCodeAccountAuthError,
		ErrCodeAccountNotFound,
		ErrCodeAccountUnknownError,
	}

	for _, code := range codes {
		s.Run(code, func() {
			path := "/api/v2/success-errors/" + code

			s.handle(http.MethodPost, path, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(w, `{"error":%q}`, code)
			})

			var result struct {
				Status string `json:"status"`
			}
			err := s.client.Post(context.Background(), path, map[string]string{"eaccount": "a@b.com"}, &result)

			s.Require().Error(err, "an error code at HTTP 200 must still be an error")

			var apiErr *APIError
			s.Require().ErrorAs(err, &apiErr)
			s.Equal(code, apiErr.Code)
			s.Equal(int64(http.StatusOK), apiErr.StatusCode)
			s.Empty(apiErr.Message)
			s.Empty(result.Status, "the destination must stay zero-valued when the body carried an error")
		})
	}
}

// TestAPIErrorSuccessBodyDecodes verifies a genuine success is not mistaken for
// an error.
func (s *clientTestSuite) TestAPIErrorSuccessBodyDecodes() {
	s.handle(http.MethodPost, "/api/v2/success-body", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(successBody))
	})

	var result struct {
		Status string `json:"status"`
	}
	err := s.client.Post(context.Background(), "/api/v2/success-body", map[string]string{"ok": "yes"}, &result)

	s.Require().NoError(err)
	s.Equal("success", result.Status)
}

// TestAPIErrorSuccessBodyWithoutErrorField verifies a success body that is not a
// JSON object still decodes normally, since it carries no error field to find.
func (s *clientTestSuite) TestAPIErrorSuccessBodyWithoutErrorField() {
	s.handle(http.MethodGet, "/api/v2/success-array", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`["one","two"]`))
	})

	var result []string
	err := s.client.Get(context.Background(), "/api/v2/success-array", &result)

	s.Require().NoError(err)
	s.Equal([]string{"one", "two"}, result)
}

// TestAPIErrorErrorsAsSupport verifies the exported contract using the exact
// errors.As call shape a consumer of this library writes.
func (s *clientTestSuite) TestAPIErrorErrorsAsSupport() {
	s.handle(http.MethodGet, "/api/v2/errors-as", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"statusCode":401,"error":"Unauthorized","message":"Invalid API key"}`))
	})

	err := s.client.Get(context.Background(), "/api/v2/errors-as", nil)
	s.Require().Error(err)

	var apiErr *APIError
	matched := errors.As(err, &apiErr)

	s.Require().True(matched, "callers must be able to unwrap the error with errors.As")
	s.Equal("Unauthorized", apiErr.Code)
}

// TestAPIErrorMessage verifies both wire shapes render readably.
func (s *clientTestSuite) TestAPIErrorMessage() {
	tests := []struct {
		name     string
		apiErr   *APIError
		expected string
	}{
		{
			name:     "status envelope",
			apiErr:   &APIError{StatusCode: 401, Code: "Unauthorized", Message: "Missing Authorization header"},
			expected: "instantly: status 401: Unauthorized: Missing Authorization header",
		},
		{
			name:     "error code at http 200",
			apiErr:   &APIError{StatusCode: 200, Code: ErrCodeAccountAuthError},
			expected: "instantly: " + ErrCodeAccountAuthError,
		},
		{
			name:     "status without message",
			apiErr:   &APIError{StatusCode: 429, Code: "Too Many Requests"},
			expected: "instantly: status 429: Too Many Requests",
		},
		{
			name:     "empty error",
			apiErr:   &APIError{},
			expected: "instantly: unknown api error",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.Equal(test.expected, test.apiErr.Error())
		})
	}
}
