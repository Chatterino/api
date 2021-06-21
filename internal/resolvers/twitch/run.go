package twitch

import (
	"encoding/json"
	"net/http"
	"net/url"
)

func run(url *url.URL, r *http.Request) ([]byte, error) {
	matches := clipSlugRegex.FindStringSubmatch(url.Path)

	if len(matches) != 3 {
		return nil, errInvalidTwitchClip
	}

	clipSlug := matches[2]

	apiResponse := clipCache.Get(clipSlug, r)
	return json.Marshal(apiResponse)
}
