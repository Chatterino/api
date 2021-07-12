package youtube

import (
	"strings"
)

type channelID struct {
	ID string
	chanType channelType
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

func constructCacheKeyFromChannelID(id channelID) string {
	return string(id.chanType) + ":" + id.ID
}

func deconstructChannelIdFromCacheKey(cacheKey string) channelID {
	splitKey := strings.Split(cacheKey, ":")

	if len(splitKey) < 2 {
		return channelID{ID: "", chanType: InvalidChannel}
	}

	return channelID{ID: splitKey[1], chanType: getChannelTypeFromString(splitKey[0])}
}

