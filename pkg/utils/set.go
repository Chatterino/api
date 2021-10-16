package utils

// SetFromSlice takes a slice and returns a set (represented as a map of struct{}s)
func SetFromSlice(slice []interface{}) map[interface{}]struct{} {
	set := make(map[interface{}]struct{})
	for _, v := range slice {
		set[v] = struct{}{}
	}

	return set
}
