package humanize

import (
	"fmt"
	"log"
	"time"
)

// Duration takes a `time.Duration` and converts it to the nearest-second string output
// Example output: 01:59:59
func Duration(duration time.Duration) string {
	// Truncate away any non-second data
	duration = duration.Truncate(1 * time.Second)

	var hours, minutes, seconds time.Duration

	hours = duration / time.Hour
	duration -= hours * time.Hour
	minutes = duration / time.Minute
	duration -= minutes * time.Minute
	seconds = duration / time.Second

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// DurationSeconds takes a `time.Duration` and converts it to a string in the following format: %gs where %g is the number of seconds contained within this duration
// Example output: 53s
func DurationSeconds(duration time.Duration) string {
	// Truncate away any non-second data
	duration = duration.Truncate(1 * time.Second)

	if duration > 90*time.Second {
		log.Println("WARNING: humanize.DurationSeconds used for duration that's larger than 90 seconds")
	}

	return fmt.Sprintf("%gs", duration.Seconds())
}

// CreationDate returns the `time.Time`'s date formatted in the `02 Jan 2006` format
// Example output: 02 Dec 2016
func CreationDate(t time.Time) string {
	return t.Format("02 Jan 2006")
}

// CreationDateRFC3339 parses the incoming string as an RFC3339-formatted date and then formats it into the `02 Jan 2006` format
// If the given string is not a valid RFC3339-formatted date, we will return an empty string
// Example output: 02 Dec 2016
// See more: `CreationDate`
func CreationDateRFC3339(str string) string {
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return ""
	}
	return CreationDate(t)
}
