package instantlytest_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-instantly/internal/instantlytest"
)

// get issues a GET against the router-backed server and returns the status and
// body.
func get(t *testing.T, server *httptest.Server, path string) (int, string) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+path, nil)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return res.StatusCode, string(body)
}

// TestRouterPathParameters verifies the router extracts :param segments and
// returns the empty string for a parameter that was not part of the pattern.
func TestRouterPathParameters(t *testing.T) {
	router := instantlytest.NewRouter()
	router.Get("/api/v2/emails/:id/detail", func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "email-123", instantlytest.PathParam(req, "id"))
		assert.Empty(t, instantlytest.PathParam(req, "missing"))
		_, _ = w.Write([]byte(instantlytest.SuccessBody))
	})

	server := httptest.NewServer(router)
	defer server.Close()

	status, body := get(t, server, "/api/v2/emails/email-123/detail")
	require.Equal(t, http.StatusOK, status)
	require.JSONEq(t, instantlytest.SuccessBody, body)
}

// TestRouterMultipleParameters verifies more than one parameter is captured.
func TestRouterMultipleParameters(t *testing.T) {
	router := instantlytest.NewRouter()
	router.Post("/api/v2/campaigns/:cid/leads/:lid", func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "c1", instantlytest.PathParam(req, "cid"))
		assert.Equal(t, "l1", instantlytest.PathParam(req, "lid"))
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(router)
	defer server.Close()

	req, err := http.NewRequestWithContext(
		t.Context(), http.MethodPost, server.URL+"/api/v2/campaigns/c1/leads/l1", nil,
	)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	require.Equal(t, http.StatusOK, res.StatusCode)
}

// TestRouterReplacesDuplicateRoutes verifies re-registering a method and pattern
// replaces the previous handler instead of shadowing it.
func TestRouterReplacesDuplicateRoutes(t *testing.T) {
	router := instantlytest.NewRouter()
	router.Get("/api/v2/replaced", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"first"}`))
	})
	router.Get("/api/v2/replaced", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"second"}`))
	})

	server := httptest.NewServer(router)
	defer server.Close()

	status, body := get(t, server, "/api/v2/replaced")
	require.Equal(t, http.StatusOK, status)
	require.JSONEq(t, `{"status":"second"}`, body)
}

// TestRouterUnknownRouteIsNotFound verifies unregistered paths fall through to a
// 404, and that a registered path with the wrong method also misses.
func TestRouterUnknownRouteIsNotFound(t *testing.T) {
	router := instantlytest.NewRouter()
	router.Get("/api/v2/known", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(router)
	defer server.Close()

	status, _ := get(t, server, "/api/v2/never-registered")
	require.Equal(t, http.StatusNotFound, status)

	// A registered path requested with an unregistered method still misses.
	req, err := http.NewRequestWithContext(t.Context(), http.MethodDelete, server.URL+"/api/v2/known", nil)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	require.Equal(t, http.StatusNotFound, res.StatusCode)
}

// TestRouterVerbHelpers verifies each verb helper registers under its method.
func TestRouterVerbHelpers(t *testing.T) {
	const path = "/api/v2/verbs"

	router := instantlytest.NewRouter()
	handler := func(expected string) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, expected, req.Method)
			w.WriteHeader(http.StatusOK)
		}
	}

	router.Get(path, handler(http.MethodGet))
	router.Post(path, handler(http.MethodPost))
	router.Put(path, handler(http.MethodPut))
	router.Patch(path, handler(http.MethodPatch))
	router.Delete(path, handler(http.MethodDelete))

	server := httptest.NewServer(router)
	defer server.Close()

	for _, method := range []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete,
	} {
		req, err := http.NewRequestWithContext(t.Context(), method, server.URL+path, nil)
		require.NoError(t, err)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusOK, res.StatusCode, method)
	}
}
