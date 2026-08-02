package instantly

// SortOrder is the direction results are sorted in. It is shared across the
// resource packages because more than one list endpoint sorts the same way.
type SortOrder string

// The directions results can be sorted in.
const (
	// SortOrderAsc sorts results oldest first.
	SortOrderAsc SortOrder = "asc"

	// SortOrderDesc sorts results newest first.
	SortOrderDesc SortOrder = "desc"
)
