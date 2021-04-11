package utils

import "strings"

// TruncateString truncates string down to the maximum length with a unicode triple dot if truncation took place
func TruncateString(s string, maxLength int) string {
	runes := []rune(s)
	if len(runes) < maxLength {
		return s
	}

	return strings.TrimSpace(string(runes[:maxLength-1])) + "…"
}

func StringPtr(s string) *string {
	return &s
}
