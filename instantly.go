package instantly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// defaultBaseURL is the root endpoint of the Instantly.ai V2 API.
const defaultBaseURL = "https://api.instantly.ai"

// Client provides a connection to the Instantly.ai V2 API.
//
// Every request is authenticated with a single V2 API key sent as a bearer
// token. V2 keys are distinct from V1 keys and are not interchangeable.
//
// See https://developer.instantly.ai/getting-started/authorization
type Client struct {
	// HTTPClient executes the requests. NewClient sets this to &http.Client{};
	// a nil value falls back to http.DefaultClient.
	HTTPClient *http.Client

	// APIKey is the Instantly V2 API key sent as a bearer token on every request.
	APIKey string

	// BaseURL is the root API endpoint. NewClient sets this to
	// https://api.instantly.ai. Request paths carry their own /api/v2 prefix.
	BaseURL string
}

// NewClient builds a Client from the provided Instantly V2 API key, using a
// default HTTP client and the default API base URL.
func NewClient(apiKey string) *Client {
	return &Client{
		HTTPClient: &http.Client{},
		APIKey:     apiKey,
		BaseURL:    defaultBaseURL,
	}
}

// get performs a GET request and decodes the response into dst.
func (client *Client) get(ctx context.Context, path string, dst any) error {
	return client.doRequest(ctx, http.MethodGet, path, nil, dst)
}

// post performs a POST request with the given payload and decodes the response
// into dst.
func (client *Client) post(ctx context.Context, path string, payload, dst any) error {
	return client.doRequest(ctx, http.MethodPost, path, payload, dst)
}

// patch performs a PATCH request with the given payload and decodes the
// response into dst.
func (client *Client) patch(ctx context.Context, path string, payload, dst any) error {
	return client.doRequest(ctx, http.MethodPatch, path, payload, dst)
}

// put performs a PUT request with the given payload and decodes the response
// into dst.
func (client *Client) put(ctx context.Context, path string, payload, dst any) error {
	return client.doRequest(ctx, http.MethodPut, path, payload, dst)
}

// delete performs a DELETE request and decodes the response into dst.
func (client *Client) delete(ctx context.Context, path string, dst any) error {
	return client.doRequest(ctx, http.MethodDelete, path, nil, dst)
}

// doRequest performs a request against the Instantly.ai V2 API.
//
// It is the single choke point for every resource method: it marshals the
// payload, authenticates the request, and converts both API error wire shapes
// into a returned error before dst is ever populated. A nil payload sends no
// body, and a nil dst skips response decoding.
func (client *Client) doRequest(ctx context.Context, method, path string, payload, dst any) (err error) {
	baseURL := client.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	// Paths carry their own /api/v2 prefix, so the base URL is the bare host.
	requestURL := baseURL + path

	var payloadData []byte
	var bodyReader io.Reader
	if payload != nil {
		if payloadData, err = json.Marshal(payload); err != nil {
			return fmt.Errorf("instantly: failed to encode request payload: %w", err)
		}
		bodyReader = bytes.NewReader(payloadData)
	}

	var req *http.Request
	if req, err = http.NewRequestWithContext(
		ctx, method, requestURL, bodyReader,
	); err != nil {
		return err
	}

	if payloadData != nil {
		// Make the payload replayable on redirect or retry.
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(payloadData)), nil
		}
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// The OpenAPI document declares `security: null` on several operations even
	// though the documented authentication is a bearer key, so the header is
	// always sent regardless of what the spec claims for a given endpoint.
	req.Header.Set("Authorization", "Bearer "+client.APIKey)

	httpClient := client.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	var res *http.Response
	if res, err = httpClient.Do(req); err != nil {
		return err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	var body []byte
	if body, err = io.ReadAll(res.Body); err != nil {
		return err
	}

	// Failures arrive either as a 4xx envelope or inside a success body, and
	// both must surface as an error before dst is touched.
	if err = checkResponse(res.StatusCode, body); err != nil {
		return err
	}

	if dst == nil {
		return nil
	}

	if err = json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("instantly: failed to decode response with status %d: %w", res.StatusCode, err)
	}

	return nil
}
