package humanize

import "github.com/Chatterino/api/pkg/utils"

// Description formats the input description into a consistent format.
// Current operations is just limiting the string to 200 characters
func Description(description string) string {
	const MaxLength = 200

	return utils.TruncateString(description, MaxLength)
}

// ShortDescription formats the input description as a short description. Example uses is the `description` key from the Wikipedia page summary API, where the summary for Forsen is "Swedish esports player and Twitch streamer"
// Current operations is just limiting the string to 60 characters
func ShortDescription(description string) string {
	const MaxLength = 60

	return utils.TruncateString(description, MaxLength)
}
