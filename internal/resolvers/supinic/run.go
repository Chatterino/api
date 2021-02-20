package supinic

import (
	"encoding/json"
	"net/url"
)

func run(url *url.URL) ([]byte, error) {
	matches := trackPathRegex.FindStringSubmatch(url.Path)
	if len(matches) != 2 {
		return nil, errInvalidTrackPath
	}

	trackID := matches[1]

	apiResponse := trackListCache.Get(trackID, nil)
	return json.Marshal(apiResponse)
}
