package supinic

import (
	"net/http"
	"net/url"
)

func run(url *url.URL, r *http.Request) ([]byte, error) {
	matches := trackPathRegex.FindStringSubmatch(url.Path)
	if len(matches) != 2 {
		return nil, errInvalidTrackPath
	}

	trackID := matches[1]

	return trackListCache.Get(trackID, r)
}
