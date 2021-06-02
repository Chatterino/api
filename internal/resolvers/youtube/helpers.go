package youtube

import (
	"net/url"
	"path"
	"strings"
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

func getYoutubeChannelIdFromURL(url *url.URL) string {
	segments := strings.Split(url.Path, "/")

	// Get the segment of the path after the channel or user segment
	for i, segment := range segments {
		if i == len(segment) - 1 {
			break
		}

		if segment == "channel" || segment == "user" {
			return segments[i + 1]
		}
	}

	return ""
}
