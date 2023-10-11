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

	v := url.Query().Get("v")

	if v == "" {
		// Early out so we can make assumptions about the response from Fields
		return ""
	}

	fields := strings.FieldsFunc(v, func(r rune) bool {
		return r == '?'
	})

	return fields[0]
}

func getYoutubeVideoIDFromURL2(url *url.URL) string {
	v := path.Base(url.Path)
	fields := strings.FieldsFunc(v, func(r rune) bool {
		return r == '?' || r == '&'
	})

	return fields[0]
}
