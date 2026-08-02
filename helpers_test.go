package instantly

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Paths the query-building cases are rendered against. More than one endpoint
// is used so the helper is exercised against the shapes it will see in practice.
const (
	testPath      = "/api/v2/emails"
	testCountPath = "/api/v2/emails/unread/count"
)

func TestBuildURLWithQuery_NilParams(t *testing.T) {
	result := buildURLWithQuery(testCountPath, nil)
	assert.Equal(t, testCountPath, result, "nil params should return the path unchanged")
}

func TestBuildURLWithQuery_EmptyParams(t *testing.T) {
	result := buildURLWithQuery(testPath, url.Values{})
	assert.Equal(t, testPath, result, "empty params should return the path unchanged")
}

func TestBuildURLWithQuery_SingleParam(t *testing.T) {
	params := url.Values{}
	params.Set("limit", "50")

	result := buildURLWithQuery(testPath, params)
	assert.Equal(t, "/api/v2/emails?limit=50", result)
}

func TestBuildURLWithQuery_MultipleParams(t *testing.T) {
	params := url.Values{}
	params.Set("limit", "50")
	params.Set("is_unread", "true")
	params.Set("mode", "emode_focused")

	result := buildURLWithQuery(testPath, params)

	assert.Contains(t, result, "/api/v2/emails?")
	assert.Contains(t, result, "limit=50")
	assert.Contains(t, result, "is_unread=true")
	assert.Contains(t, result, "mode=emode_focused")
}

func TestBuildURLWithQuery_EncodesSpecialCharacters(t *testing.T) {
	params := url.Values{}
	params.Set("search", "hello world")
	params.Set("eaccount", "sender@example.com")

	result := buildURLWithQuery(testPath, params)

	// Spaces and reserved characters must be percent-encoded, never raw.
	assert.Contains(t, result, "search=hello+world")
	assert.Contains(t, result, "%40")
	assert.NotContains(t, result, "sender@example.com")
}

func TestBuildURLWithQuery_PreservesEmptyValue(t *testing.T) {
	params := url.Values{}
	params.Set("search", "")

	result := buildURLWithQuery(testCountPath, params)
	assert.Equal(
		t, testCountPath+"?search=", result, "an explicitly set empty value is still a parameter",
	)
}
