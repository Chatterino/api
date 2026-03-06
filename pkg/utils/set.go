package utils

// SetFromSlice takes a slice and returns a set (represented as a map of struct{}s)
func SetFromSlice(slice []any) map[any]struct{} {
	set := make(map[any]struct{})
	for _, v := range slice {
		set[v] = struct{}{}
	}

	return set
}
