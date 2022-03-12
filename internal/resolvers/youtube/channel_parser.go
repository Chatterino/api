package youtube

import (
	"fmt"
	"strings"
)

type Channel struct {
	ID   string
	Type channelType
}

func (c *Channel) ToCacheKey() string {
	return fmt.Sprintf("%s:%s", string(c.Type), c.ID)
}

// getChannelFromPath parsers a path to a YouTube channel to a Channel struct containing the ID of the channel,
// and a helper type which helps us figure out which API we need to send the ID to
func getChannelFromPath(path string) Channel {
	match := youtubeChannelRegex.FindStringSubmatch(path)
	if match == nil || len(match) != 3 {
		return Channel{
			ID:   "",
			Type: InvalidChannel,
		}
	}

	c := Channel{
		ID:   match[2],
		Type: getChannelTypeFromString(strings.TrimSuffix(match[1], "/")),
	}

	if c.Type == CustomChannel || c.Type == UserChannel {
		c.ID = strings.ToLower(c.ID)
	}

	return c
}

func getChannelFromCacheKey(cacheKey string) Channel {
	splitKey := strings.Split(cacheKey, ":")

	if len(splitKey) < 2 {
		return Channel{
			ID:   "",
			Type: InvalidChannel,
		}
	}

	return Channel{
		ID:   splitKey[1],
		Type: getChannelTypeFromString(splitKey[0]),
	}
}

// Gets the channel type from a cache key type segment - used to identify what YouTube API to use
func getChannelTypeFromString(channelType string) channelType {
	switch channelType {
	case "c":
		return CustomChannel
	case "user":
		return UserChannel
	case "channel":
		return IdentifierChannel
	case "":
		return CustomChannel
	}

	return InvalidChannel
}
