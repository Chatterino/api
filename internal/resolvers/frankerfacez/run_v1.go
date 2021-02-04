package frankerfacez

import (
	"encoding/json"
	"net/url"
)

func run(url *url.URL) ([]byte, error) {
	matches := emotePathRegex.FindStringSubmatch(url.Path)
	if len(matches) != 4 {
		return nil, errInvalidFrankerFaceZEmotePath
	}

	emoteHash := matches[1]

	apiResponse := emoteCache.Get(emoteHash, nil)
	return json.Marshal(apiResponse)

}
