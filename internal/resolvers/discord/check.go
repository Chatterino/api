package discord

import (
	"fmt"
	"net/url"
	"strings"
)

func check(url *url.URL) bool {
	return discordInviteURLRegex.MatchString(fmt.Sprintf("%s%s", strings.ToLower(url.Host), url.Path))
}
