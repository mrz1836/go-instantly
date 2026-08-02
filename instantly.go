package instantly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	// defaultBaseURL is the root endpoint of the Instantly.ai V2 API.
	defaultBaseURL = "https://api.instantly.ai"

	// mediaTypeJSON is the media type sent for both Accept and Content-Type.
	mediaTypeJSON = "application/json"
)

// Client provides a connection to the Instantly.ai V2 API.
//
// Every request is authenticated with a single V2 API key sent as a bearer
// token. V2 keys are distinct from V1 keys and are not interchangeable.
//
// A Client is safe for concurrent use by multiple goroutines. It is immutable
// after construction: configure it once with the functional options passed to
// NewClient rather than mutating it afterwards.
//
// The low-level request plumbing (Do, Get, Post, Patch, Put, Delete, DoRaw) is
// exported so the resource subpackages — which live in their own packages — can
// build on it, and so consumers keep a forward-compatible escape hatch for any
// endpoint the typed resource packages do not yet wrap.
//
// See https://developer.instantly.ai/getting-started/authorization
type Client struct {
	// httpClient executes the requests. NewClient defaults this to a zero-value
	// http.Client; a nil value falls back to http.DefaultClient.
	httpClient *http.Client

	// apiKey is the Instantly V2 API key sent as a bearer token on every request.
	apiKey string

	// baseURL is the root API endpoint. NewClient defaults this to
	// https://api.instantly.ai. Request paths carry their own /api/v2 prefix.
	baseURL string

	// userAgent, when set, is sent as the User-Agent header on every request.
	userAgent string

	// headers are extra headers merged onto every request before the mandatory
	// Accept, Content-Type, and Authorization headers are applied.
	headers http.Header
}

// ClientOption configures a Client during construction.
//
// Options are applied in order, so a later option overrides an earlier one that
// sets the same field.
type ClientOption func(*Client)

// NewClient builds a Client from the provided Instantly V2 API key.
//
// Without options it uses a default HTTP client and the default API base URL.
// Pass functional options such as WithHTTPClient or WithBaseURL to customize it.
func NewClient(apiKey string, opts ...ClientOption) *Client {
	client := &Client{
		httpClient: &http.Client{},
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(client)
		}
	}

	return client
}

// WithHTTPClient sets the HTTP client used to execute requests, which is how a
// caller supplies a custom transport, proxy, or timeout.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

// WithBaseURL overrides the root API endpoint, chiefly to point the client at a
// mock server in tests. An empty value falls back to the default base URL.
func WithBaseURL(baseURL string) ClientOption {
	return func(client *Client) {
		client.baseURL = baseURL
	}
}

// WithUserAgent sets the User-Agent header sent on every request.
func WithUserAgent(userAgent string) ClientOption {
	return func(client *Client) {
		client.userAgent = userAgent
	}
}

// WithHTTPHeader adds an extra header sent on every request.
//
// The header is merged onto each request before the mandatory Accept,
// Content-Type, and Authorization headers, so it cannot override them. Calling
// it more than once with the same key adds multiple values for that header.
func WithHTTPHeader(key, value string) ClientOption {
	return func(client *Client) {
		if client.headers == nil {
			client.headers = http.Header{}
		}
		client.headers.Add(key, value)
	}
}

// Ptr returns a pointer to v.
//
// It is a convenience for populating the optional, pointer-typed fields the API
// models as nullable: instantly.Ptr(true) yields a *bool without a named local.
func Ptr[T any](v T) *T {
	return &v
}

// Do performs an API request and decodes the response into out.
//
// It is the single choke point every resource method routes through: it
// marshals the payload, authenticates the request, and converts both API error
// wire shapes into a returned error before out is ever populated. A nil payload
// sends no body, and a nil out skips response decoding.
func (c *Client) Do(ctx context.Context, method, path string, payload, out any) error {
	status, body, err := c.do(ctx, method, path, payload)
	if err != nil {
		return err
	}

	if out == nil {
		return nil
	}

	if err = json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("instantly: failed to decode response with status %d: %w", status, err)
	}

	return nil
}

// DoRaw performs an API request and returns the raw response body.
//
// It runs the same error handling as Do — a failing status or an error embedded
// in a success body is returned as an error — but hands back the undecoded bytes
// for endpoints that answer with a non-JSON payload, such as a CSV export.
func (c *Client) DoRaw(ctx context.Context, method, path string, payload any) ([]byte, error) {
	_, body, err := c.do(ctx, method, path, payload)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// Get performs a GET request and decodes the response into out.
func (c *Client) Get(ctx context.Context, path string, out any) error {
	return c.Do(ctx, http.MethodGet, path, nil, out)
}

// Post performs a POST request with the given payload and decodes the response
// into out.
func (c *Client) Post(ctx context.Context, path string, payload, out any) error {
	return c.Do(ctx, http.MethodPost, path, payload, out)
}

// Patch performs a PATCH request with the given payload and decodes the response
// into out.
func (c *Client) Patch(ctx context.Context, path string, payload, out any) error {
	return c.Do(ctx, http.MethodPatch, path, payload, out)
}

// Put performs a PUT request with the given payload and decodes the response
// into out.
func (c *Client) Put(ctx context.Context, path string, payload, out any) error {
	return c.Do(ctx, http.MethodPut, path, payload, out)
}

// Delete performs a DELETE request and decodes the response into out.
//
// Endpoints that delete with a request body are rare; call Do with
// http.MethodDelete and a payload directly for those.
func (c *Client) Delete(ctx context.Context, path string, out any) error {
	return c.Do(ctx, http.MethodDelete, path, nil, out)
}

// do executes a request and returns the status code and body once the response
// has been checked for an API error, whichever wire shape it arrived in.
//
// It is the shared core of Do and DoRaw: neither the JSON decode nor the raw
// hand-off happens until after checkResponse has cleared the response.
func (c *Client) do(ctx context.Context, method, path string, payload any) (int, []byte, error) {
	req, err := c.newRequest(ctx, method, path, payload)
	if err != nil {
		return 0, nil, err
	}

	res, err := c.httpClientOrDefault().Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("instantly: request failed: %w", err)
	}

	defer func() {
		_ = res.Body.Close()
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, nil, fmt.Errorf("instantly: failed to read response body: %w", err)
	}

	// Failures arrive either as a 4xx envelope or inside a success body, and
	// both must surface as an error before the body is handed back.
	if err = checkResponse(res.StatusCode, body); err != nil {
		return res.StatusCode, body, err
	}

	return res.StatusCode, body, nil
}

// newRequest builds an authenticated request, marshaling payload as its body.
func (c *Client) newRequest(ctx context.Context, method, path string, payload any) (*http.Request, error) {
	baseURL := c.baseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	var payloadData []byte
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("instantly: failed to encode request payload: %w", err)
		}
		payloadData = data
	}

	var bodyReader io.Reader
	if payloadData != nil {
		bodyReader = bytes.NewReader(payloadData)
	}

	// Paths carry their own /api/v2 prefix, so the base URL is the bare host.
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("instantly: failed to build request: %w", err)
	}

	if payloadData != nil {
		// Make the payload replayable on redirect or retry.
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(payloadData)), nil
		}
	}

	c.applyHeaders(req)

	return req, nil
}

// applyHeaders sets the mandatory request headers, after any caller-supplied
// extras so the mandatory ones always win.
func (c *Client) applyHeaders(req *http.Request) {
	for key, values := range c.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	req.Header.Set("Accept", mediaTypeJSON)
	req.Header.Set("Content-Type", mediaTypeJSON)

	// The OpenAPI document declares `security: null` on several operations even
	// though the documented authentication is a bearer key, so the header is
	// always sent regardless of what the spec claims for a given endpoint.
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
}

// httpClientOrDefault returns the configured HTTP client, or the shared default
// when none was supplied.
func (c *Client) httpClientOrDefault() *http.Client {
	if c.httpClient == nil {
		return http.DefaultClient
	}

	return c.httpClient
}
