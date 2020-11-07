package resolver

import (
	"net/url"
	"strings"
)

func MatchesHosts(url *url.URL, hosts map[string]struct{}) (bool, string) {
	host := strings.ToLower(url.Host)
	if _, ok := hosts[host]; !ok {
		return false, ""
	}

	return true, host
}
