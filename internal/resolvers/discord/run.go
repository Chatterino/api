package discord

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func run(url *url.URL, r *http.Request) ([]byte, error) {
	matches := discordInviteURLRegex.FindStringSubmatch(fmt.Sprintf("%s%s", strings.ToLower(url.Host), url.Path))
	if len(matches) != 4 {
		return nil, errInvalidDiscordInvite
	}

	inviteCode := matches[3]

	apiResponse := inviteCache.Get(inviteCode, r)
	return json.Marshal(apiResponse)
}
