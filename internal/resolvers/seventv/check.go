package seventv

import (
	"net/url"
	"strings"
)

func check(url *url.URL) bool {
	return seventvEmoteURLRegex.MatchString(strings.ToLower(url.Host) + url.Path)
}
