package twitch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func run(url *url.URL, r *http.Request) ([]byte, error) {
	matches := twitchClipURLRegex.FindStringSubmatch(fmt.Sprintf("%s%s", strings.ToLower(url.Host), url.Path))
	if len(matches) != 4 {
		return nil, errInvalidTwitchClip
	}

	clipSlug := matches[2]

	apiResponse := clipCache.Get(clipSlug, r)
	return json.Marshal(apiResponse)
}
