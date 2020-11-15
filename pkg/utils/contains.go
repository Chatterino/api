package utils

// Contains returns true if the given string array (haystack) contains given string (needle)
func Contains(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}

	return false
}
