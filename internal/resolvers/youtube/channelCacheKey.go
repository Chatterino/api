package youtube

import (
	"strings"
)

type channelId struct {
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

func constructCacheKeyFromChannelId(id channelId) string {
	return string(id.chanType) + ":" + id.ID
}

func deconstructChannelIdFromCacheKey(cacheKey string) channelId  {
	splitKey := strings.Split(cacheKey, ":")

	if len(splitKey) < 2 {
		return channelId{ID: "", chanType: InvalidChannel}
	}

	return channelId{ID: splitKey[1], chanType: getChannelTypeFromString(splitKey[0])}
}

