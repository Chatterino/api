package seventv

import (
	"net/url"

	"github.com/Chatterino/api/pkg/resolver"
)

func check(url *url.URL) bool {
	if match, _ := resolver.MatchesHosts(url, domains); !match {
		return false
	}

	return emotePathRegex.MatchString(url.Path)
}
