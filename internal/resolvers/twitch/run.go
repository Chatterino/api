package twitch

import (
	"net/http"
	"net/url"
)

func run(url *url.URL, r *http.Request) ([]byte, error) {
	clipSlug, err := parseClipSlug(url)
	if err != nil {
		return nil, err
	}

	return clipCache.Get(clipSlug, r)
}
