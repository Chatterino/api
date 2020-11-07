package frankerfacez

import (
	"net/url"

	"github.com/Chatterino/api/pkg/resolver"
)

func check(url *url.URL) bool {
	if match, _ := resolver.MatchesHosts(url, domains); !match {
		return false
	}

	if !emotePathRegex.MatchString(url.Path) {
		return false
	}

	return true
}
