package youtube

import (
	"net/url"
	"path"
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

	// ex: https://www.youtube.com/shorts/nSW6scUfnFw
	if base, rest := path.Split(url.Path); base == "/shorts/" {
		return rest
	}

	return url.Query().Get("v")
}

func getYoutubeVideoIDFromURL2(url *url.URL) string {
	return path.Base(url.Path)
}
