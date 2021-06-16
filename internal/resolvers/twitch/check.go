package twitch

import (
	"fmt"
	"net/url"
	"strings"
)

func check(url *url.URL) bool {
	return twitchClipURLRegex.MatchString(fmt.Sprintf("%s%s", strings.ToLower(url.Host), url.Path))
}
