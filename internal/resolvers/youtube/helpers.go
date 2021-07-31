package youtube

import (
	"net/url"
	"path"
	"regexp"
	"strings"
)

type channelType string

const (
	// InvalidChannel channel isn't of a known type or doesn't exist
	InvalidChannel channelType = ""
	// UserChannel channel ID is a username
	UserChannel = "user"
	// IdentifierChannel channel uses the YouTube channel ID format (UC*)
	IdentifierChannel = "channel"
	// CustomChannel channel uses a custom URL and requires a Search call for the ID
	CustomChannel = "c"
)

func getYoutubeVideoIDFromURL(url *url.URL) string {
	if strings.Contains(url.Path, "embed") {
		return path.Base(url.Path)
	}

	return url.Query().Get("v")
}

func getYoutubeVideoIDFromURL2(url *url.URL) string {
	return path.Base(url.Path)
}

func getYoutubeChannelIDFromURL(url *url.URL) channelID {
	pattern, err := regexp.Compile(`(user|c(?:hannel)?)/([\w-]+)`)
	if err != nil {
		return channelID{ID: "", chanType: InvalidChannel}
	}

	match := pattern.FindStringSubmatch(url.Path)
	if match == nil || len(match) < 3 {
		return channelID{ID: "", chanType: InvalidChannel}
	}

	return channelID{ID: match[2], chanType: getChannelTypeFromString(match[1])}
}
