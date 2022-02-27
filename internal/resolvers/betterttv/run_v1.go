package betterttv

import (
	"net/http"
	"net/url"
)

func run(url *url.URL, r *http.Request) ([]byte, error) {
	matches := emotePathRegex.FindStringSubmatch(url.Path)
	if len(matches) != 2 {
		return nil, errInvalidBTTVEmotePath
	}

	emoteHash := matches[1]

	return emoteCache.Get(emoteHash, r)
}
