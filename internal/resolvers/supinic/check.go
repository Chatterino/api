package supinic

import (
	"net/url"
	"strings"
)

func check(url *url.URL) bool {
	host := strings.ToLower(url.Host)

	if _, ok := trackListDomains[host]; !ok {
		return false
	}

	if !trackPathRegex.MatchString(url.Path) {
		return false
	}

	return true
}
