package twitch

import (
	"net/url"
	"regexp"

	"github.com/Chatterino/api/pkg/resolver"
)

var (
	clipSlugRegex = regexp.MustCompile(`^\/(\w{2,25}\/clip\/)?([a-zA-Z0-9]+(?:-[-\w]{16})?)$`)
)

func check(url *url.URL) bool {
	// Regardless of domain path needs to match anyway, so we do it here to avoid duplication
	matches := clipSlugRegex.FindStringSubmatch(url.Path)

	if len(matches) != 3 {
		return false
	}

	match, domain := resolver.MatchesHosts(url, domains)
	if !match {
		return false
	}

	// Find clips that look like https://clips.twitch.tv/SlugHere
	if domain == "clips.twitch.tv" {
		// matches[1] contains "StreamerName/clip/" - we don't want it in this check though
		return matches[1] == ""
	}

	// Find clips that look like https://twitch.tv/StreamerName/clip/SlugHere
	// matches[1] contains "StreamerName/clip/" - we need it in this check
	return matches[1] != ""
}
