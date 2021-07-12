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

func getChannelTypeFromString(channelType string) channelType  {
	switch channelType {
		case "c":
			return CustomChannel
		case "user":
			return UserChannel
		case "channel":
			return IdentifierChannel
	}

	return InvalidChannel
}

func constructCacheKeyFromChannelId(id channelId) string {
	return string(id.chanType) + ":" + id.id
}

func deconstructChannelIdFromCacheKey(cacheKey string) channelId  {
	splitKey := strings.Split(cacheKey, ":")

	if len(splitKey) < 2 {
		return channelId{id: "", chanType: InvalidChannel}
	}

	return channelId{id: splitKey[1], chanType: getChannelTypeFromString(splitKey[0])}
}

func getYoutubeChannelIdFromURL(url *url.URL) channelId {
	pattern, err := regexp.Compile(`(user|c(?:hannel)?)/([\w-]+)`)
	if err != nil {
		return channelId{id: "", chanType: InvalidChannel}
	}

	match := pattern.FindStringSubmatch(url.Path)
	if match == nil || len(match) < 3 {
		return channelId{id: "", chanType: InvalidChannel}
	}

	return channelId{id: match[2], chanType: getChannelTypeFromString(match[1])}
}
