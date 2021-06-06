package seventv

import (
	"encoding/json"
	"net/url"
)

func run(url *url.URL) ([]byte, error) {
	matches := emotePathRegex.FindStringSubmatch(url.Path)
	if len(matches) != 2 {
		return nil, errInvalidSevenTVEmotePath
	}

	emoteHash := matches[1]

	apiResponse := emoteCache.Get(emoteHash, nil)
	return json.Marshal(apiResponse)
}
