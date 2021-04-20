package supinic

import (
	"net/url"

	"github.com/Chatterino/api/pkg/utils"
)

func check(url *url.URL) bool {
	if !utils.IsDomains(url, trackListDomains) {
		return false
	}

	if !trackPathRegex.MatchString(url.Path) {
		return false
	}

	return true
}
