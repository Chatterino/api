package twitch

import (
	"encoding/json"
	"net/http"
	"net/url"
)

func run(url *url.URL, r *http.Request) ([]byte, error) {
	clipSlug, err := parseClipSlug(url)
	if err != nil {
		return nil, err
	}

	apiResponse := clipCache.Get(clipSlug, r)
	return json.Marshal(apiResponse)
}
