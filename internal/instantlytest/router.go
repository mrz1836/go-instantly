// Package instantlytest provides the shared test harness for the go-instantly
// SDK: a small HTTP mock router, a reusable testify suite, and assertion
// helpers.
//
// It imports testify and is therefore test-support code only. It is reachable
// exclusively from _test.go files, so importing a resource package never drags
// testify into a consumer's build. It must never be imported by non-test code.
package instantlytest

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

// Router is a small HTTP mock router for tests. It supports method-based routing
// and path-parameter extraction.
//
// Patterns carry the full request path, including the /api/v2 prefix the client
// sends, for example "/api/v2/emails/:id".
type Router struct {
	routes []route
}

type route struct {
	method  string
	pattern string
	handler http.HandlerFunc
	regex   *regexp.Regexp
	params  []string
}

// NewRouter creates a new mock router.
func NewRouter() *Router {
	return &Router{
		routes: make([]route, 0),
	}
}

// ServeHTTP implements the http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, rt := range r.routes {
		if rt.method != req.Method {
			continue
		}

		if r.matchRoute(rt, req, w) {
			return
		}
	}

	// No route found.
	http.NotFound(w, req)
}

// HandleFunc registers a handler function for the given method and pattern.
//
// Registering the same method and pattern twice replaces the previous handler,
// so a suite can re-register a route with a different fixture.
func (r *Router) HandleFunc(method, pattern string, handler http.HandlerFunc) {
	rt := route{
		method:  method,
		pattern: pattern,
		handler: handler,
	}

	// Check if pattern contains parameters (e.g., :id).
	if strings.Contains(pattern, ":") {
		rt.regex, rt.params = compilePattern(pattern)
	}

	// Replace existing route with same method+pattern.
	for i, existing := range r.routes {
		if existing.method == method && existing.pattern == pattern {
			r.routes[i] = rt
			return
		}
	}

	r.routes = append(r.routes, rt)
}

// Get registers a GET handler for the given pattern.
func (r *Router) Get(pattern string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodGet, pattern, handler)
}

// Post registers a POST handler for the given pattern.
func (r *Router) Post(pattern string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodPost, pattern, handler)
}

// Put registers a PUT handler for the given pattern.
func (r *Router) Put(pattern string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodPut, pattern, handler)
}

// Delete registers a DELETE handler for the given pattern.
func (r *Router) Delete(pattern string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodDelete, pattern, handler)
}

// Patch registers a PATCH handler for the given pattern.
func (r *Router) Patch(pattern string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodPatch, pattern, handler)
}

// matchRoute attempts to match a route and execute its handler.
func (r *Router) matchRoute(rt route, req *http.Request, w http.ResponseWriter) bool {
	if rt.regex == nil {
		return r.matchExactRoute(rt, req, w)
	}
	return r.matchPatternRoute(rt, req, w)
}

// matchExactRoute handles routes without parameters.
func (r *Router) matchExactRoute(rt route, req *http.Request, w http.ResponseWriter) bool {
	if rt.pattern == req.URL.Path {
		rt.handler(w, req)
		return true
	}
	return false
}

// matchPatternRoute handles routes with path parameters.
func (r *Router) matchPatternRoute(rt route, req *http.Request, w http.ResponseWriter) bool {
	matches := rt.regex.FindStringSubmatch(req.URL.Path)
	if matches == nil {
		return false
	}

	// Extract path parameters and add them to the request context.
	if len(matches) > 1 && len(rt.params) > 0 {
		ctx := req.Context()
		for i, param := range rt.params {
			if i+1 < len(matches) {
				ctx = context.WithValue(ctx, paramKey(param), matches[i+1])
			}
		}
		req = req.WithContext(ctx)
	}
	rt.handler(w, req)
	return true
}

// compilePattern converts a pattern like "/api/v2/emails/:id" to a regex and
// extracts the parameter names.
func compilePattern(pattern string) (*regexp.Regexp, []string) {
	var params []string
	regexPattern := "^"

	parts := strings.Split(pattern, "/")
	for _, part := range parts {
		if part == "" {
			continue
		}

		if strings.HasPrefix(part, ":") {
			// Parameter segment.
			paramName := part[1:] // Remove the ':'.
			params = append(params, paramName)
			regexPattern += "/([^/]+)"
		} else {
			// Literal segment.
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

// paramKey is the key type for storing path parameters in the request context.
type paramKey string

// PathParam extracts a path parameter from the request context. It is only
// populated for handlers registered with a parameterized route.
func PathParam(r *http.Request, key string) string {
	if value := r.Context().Value(paramKey(key)); value != nil {
		if param, ok := value.(string); ok {
			return param
		}
	}
	return ""
}
