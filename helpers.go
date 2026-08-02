package instantly

import (
	"fmt"
	"net/url"
)

// buildURLWithQuery appends encoded query parameters to a path, returning the
// bare path when there are none.
//
// Query parameters reach this helper only as url.Values rendered by a
// resource's functional options, which keeps every query key and value
// compile-time checked.
func buildURLWithQuery(path string, queryParams url.Values) string {
	if len(queryParams) == 0 {
		return path
	}

	return fmt.Sprintf("%s?%s", path, queryParams.Encode())
}
