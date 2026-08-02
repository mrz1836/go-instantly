package instantly

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Paths the query-building cases are rendered against. More than one endpoint is
// used so the builder is exercised against the shapes it will see in practice.
const (
	testPath      = "/api/v2/emails"
	testCountPath = "/api/v2/emails/unread/count"
)

func TestQuerySetters(t *testing.T) {
	q := NewQuery()
	q.SetString("search", "hello world")
	q.SetInt("limit", 50)
	q.SetBool("is_unread", true)

	assert.Equal(t, 3, q.Len())
	assert.Equal(t, "hello world", q.Get("search"))
	assert.Equal(t, "50", q.Get("limit"))
	assert.Equal(t, "true", q.Get("is_unread"))
}

func TestQueryAddStringRepeats(t *testing.T) {
	q := NewQuery()
	q.AddString("emails", "a@b.com")
	q.AddString("emails", "c@d.com")

	assert.Equal(t, 1, q.Len(), "repeated values share one key")
	assert.Equal(t, "a@b.com", q.Get("emails"), "Get returns the first value")
	assert.Equal(t, "emails=a%40b.com&emails=c%40d.com", q.Encode())
}

func TestQuerySetLastValueWins(t *testing.T) {
	q := NewQuery()
	q.SetInt("limit", 25)
	q.SetInt("limit", 100)

	assert.Equal(t, 1, q.Len(), "setting the same key twice keeps one parameter")
	assert.Equal(t, "100", q.Get("limit"), "the last value supplied wins")
}

func TestQuerySettersAreChainable(t *testing.T) {
	q := NewQuery().SetString("a", "1").SetInt("b", 2).SetBool("c", false)

	assert.Equal(t, 3, q.Len())
	assert.Equal(t, "1", q.Get("a"))
	assert.Equal(t, "2", q.Get("b"))
	assert.Equal(t, "false", q.Get("c"))
}

func TestQueryGetUnset(t *testing.T) {
	q := NewQuery()
	assert.Empty(t, q.Get("missing"), "an unset key reads as the empty string")
	assert.Zero(t, q.Len())
}

func TestSetEnum(t *testing.T) {
	q := NewQuery()
	SetEnum(q, "sort_order", SortOrderDesc)

	assert.Equal(t, "desc", q.Get("sort_order"))
	assert.Equal(t, 1, q.Len())
}

func TestQueryEncode(t *testing.T) {
	q := NewQuery()
	q.SetString("search", "hello world")
	q.SetString("eaccount", "sender@example.com")

	encoded := q.Encode()
	assert.Contains(t, encoded, "search=hello+world")
	assert.Contains(t, encoded, "%40")
	assert.NotContains(t, encoded, "sender@example.com")
}

func TestQueryPathEmpty(t *testing.T) {
	assert.Equal(t, testPath, NewQuery().Path(testPath), "no parameters returns the bare path")
}

func TestQueryPathNilReceiver(t *testing.T) {
	var q *Query
	assert.Equal(t, testPath, q.Path(testPath), "a nil query returns the bare path")
}

func TestQueryPathWithParams(t *testing.T) {
	q := NewQuery().SetInt("limit", 50)
	assert.Equal(t, "/api/v2/emails?limit=50", q.Path(testPath))
}

func TestBuildPathNilParams(t *testing.T) {
	assert.Equal(t, testCountPath, BuildPath(testCountPath, nil), "nil params returns the path unchanged")
}

func TestBuildPathEmptyParams(t *testing.T) {
	assert.Equal(t, testPath, BuildPath(testPath, url.Values{}), "empty params returns the path unchanged")
}

func TestBuildPathSingleParam(t *testing.T) {
	params := url.Values{}
	params.Set("limit", "50")

	assert.Equal(t, "/api/v2/emails?limit=50", BuildPath(testPath, params))
}

func TestBuildPathMultipleParams(t *testing.T) {
	params := url.Values{}
	params.Set("limit", "50")
	params.Set("is_unread", "true")
	params.Set("mode", "emode_focused")

	result := BuildPath(testPath, params)

	assert.Contains(t, result, "/api/v2/emails?")
	assert.Contains(t, result, "limit=50")
	assert.Contains(t, result, "is_unread=true")
	assert.Contains(t, result, "mode=emode_focused")
}

func TestBuildPathEncodesSpecialCharacters(t *testing.T) {
	params := url.Values{}
	params.Set("search", "hello world")
	params.Set("eaccount", "sender@example.com")

	result := BuildPath(testPath, params)

	// Spaces and reserved characters must be percent-encoded, never raw.
	assert.Contains(t, result, "search=hello+world")
	assert.Contains(t, result, "%40")
	assert.NotContains(t, result, "sender@example.com")
}

func TestBuildPathPreservesEmptyValue(t *testing.T) {
	params := url.Values{}
	params.Set("search", "")

	result := BuildPath(testCountPath, params)
	require.Equal(
		t, testCountPath+"?search=", result, "an explicitly set empty value is still a parameter",
	)
}
