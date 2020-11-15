package humanize

import "time"

// Date converts a date from a string in the RFC3339 format into one specified in the format string
func Date(format string, str string) string {
	date, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return ""
	}
	return date.Format(format)
}
