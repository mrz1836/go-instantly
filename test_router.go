package instantly

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

// TestRouter provides simple HTTP routing for testing purposes.
// It supports method-based routing and path parameter extraction.
//
// Patterns carry the full request path, including the /api/v2 prefix the client
// sends, for example "/api/v2/emails/:id".
type TestRouter struct {
	routes []route
}

type route struct {
	method  string
	pattern string
	handler http.HandlerFunc
	regex   *regexp.Regexp
	params  []string
}

// NewTestRouter creates a new test router.
func NewTestRouter() *TestRouter {
	return &TestRouter{
		routes: make([]route, 0),
	}
}

// ServeHTTP implements the http.Handler interface.
func (r *TestRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, rt := range r.routes {
		if rt.method != req.Method {
			continue
		}

		if r.matchRoute(rt, req, w) {
			return
		}
	}

	// No route found
	http.NotFound(w, req)
}

// HandleFunc registers a handler function for the given method and pattern.
//
// Registering the same method and pattern twice replaces the previous handler,
// so a suite can re-register a route with a different fixture.
func (r *TestRouter) HandleFunc(method, pattern string, handler http.HandlerFunc) {
	rt := route{
		method:  method,
		pattern: pattern,
		handler: handler,
	}

	// Check if pattern contains parameters (e.g., :id)
	if strings.Contains(pattern, ":") {
		rt.regex, rt.params = compilePattern(pattern)
	}

	// Replace existing route with same method+pattern
	for i, existing := range r.routes {
		if existing.method == method && existing.pattern == pattern {
			r.routes[i] = rt
			return
		}
	}

	r.routes = append(r.routes, rt)
}

// Get registers a GET handler for the given pattern.
func (r *TestRouter) Get(pattern string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodGet, pattern, handler)
}

// Post registers a POST handler for the given pattern.
func (r *TestRouter) Post(pattern string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodPost, pattern, handler)
}

// Put registers a PUT handler for the given pattern.
func (r *TestRouter) Put(pattern string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodPut, pattern, handler)
}

// Delete registers a DELETE handler for the given pattern.
func (r *TestRouter) Delete(pattern string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodDelete, pattern, handler)
}

// Patch registers a PATCH handler for the given pattern.
func (r *TestRouter) Patch(pattern string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodPatch, pattern, handler)
}

// matchRoute attempts to match a route and execute its handler.
func (r *TestRouter) matchRoute(route route, req *http.Request, w http.ResponseWriter) bool {
	if route.regex == nil {
		return r.matchExactRoute(route, req, w)
	}
	return r.matchPatternRoute(route, req, w)
}

// matchExactRoute handles routes without parameters.
func (r *TestRouter) matchExactRoute(route route, req *http.Request, w http.ResponseWriter) bool {
	if route.pattern == req.URL.Path {
		route.handler(w, req)
		return true
	}
	return false
}

// matchPatternRoute handles routes with path parameters.
func (r *TestRouter) matchPatternRoute(route route, req *http.Request, w http.ResponseWriter) bool {
	matches := route.regex.FindStringSubmatch(req.URL.Path)
	if matches == nil {
		return false
	}

	// Extract path parameters and add to request context
	if len(matches) > 1 && len(route.params) > 0 {
		ctx := req.Context()
		for i, param := range route.params {
			if i+1 < len(matches) {
				ctx = context.WithValue(ctx, paramKey(param), matches[i+1])
			}
		}
		req = req.WithContext(ctx)
	}
	route.handler(w, req)
	return true
}

// compilePattern converts a pattern like "/api/v2/emails/:id"
// to a regex and extracts parameter names.
func compilePattern(pattern string) (*regexp.Regexp, []string) {
	var params []string
	regexPattern := "^"

	parts := strings.Split(pattern, "/")
	for _, part := range parts {
		if part == "" {
			continue
		}

		if strings.HasPrefix(part, ":") {
			// Parameter segment
			paramName := part[1:] // Remove the ':'
			params = append(params, paramName)
			regexPattern += "/([^/]+)"
		} else {
			// Literal segment
			regexPattern += "/" + regexp.QuoteMeta(part)
		}
	}

	regexPattern += "$"

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		panic("invalid pattern: " + pattern)
	}

	return regex, params
}

// paramKey is used as the key type for storing path parameters in request context.
type paramKey string

// GetPathParam extracts a path parameter from the request context.
// This is only available for handlers registered with parameterized routes.
func GetPathParam(r *http.Request, key string) string {
	if value := r.Context().Value(paramKey(key)); value != nil {
		if param, ok := value.(string); ok {
			return param
		}
	}
	return ""
}
