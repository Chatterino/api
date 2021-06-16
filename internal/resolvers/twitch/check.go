package twitch

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/Chatterino/api/pkg/utils"
)

var (
	clipSlugRegex = regexp.MustCompile(`^\/(\w{2,25}\/clip\/)?([a-zA-Z0-9]+(-[-\w]{16})?)$`)
)

func check(url *url.URL) bool {
	// Regardless of domain path needs to match anyway, so we do it here to avoid duplication
	matches := clipSlugRegex.FindStringSubmatch(url.Path)

	if len(matches) != 4 {
		return false
	}

	// Find clips that look like https://clips.twitch.tv/SlugHere
	if utils.IsDomain(url, "clips.twitch.tv") {
		// matches[1] contains "StreamerName/clip/" - we don't want it in this check though
		if matches[1] != "" {
			return false
		}

		return true
	}

	// Find clips that look like https://twitch.tv/StreamerName/clip/SlugHere
	if utils.IsDomain(url, "twitch.tv") {
		// matches[1] contains "StreamerName/clip/" - we need this in this check though
		if matches[1] == "" {
			return false
		}

		return true
	}

	return false
}
