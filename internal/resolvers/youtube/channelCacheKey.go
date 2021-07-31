package youtube

import (
	"strings"
)

type channelID struct {
	ID string
	chanType channelType
}

// Gets the channel type from a cache key type segment - used to identify what YouTube API to use
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

func constructCacheKeyFromChannelID(ID channelID) string {
	return string(ID.chanType) + ":" + ID.ID
}

func deconstructChannelIDFromCacheKey(cacheKey string) channelID {
	splitKey := strings.Split(cacheKey, ":")

	if len(splitKey) < 2 {
		return channelID{ID: "", chanType: InvalidChannel}
	}

	return channelID{ID: splitKey[1], chanType: getChannelTypeFromString(splitKey[0])}
}

