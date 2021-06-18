package frankerfacez

import (
	"encoding/json"
	"net/http"
	"net/url"
)

func run(url *url.URL, r *http.Request) ([]byte, error) {
	matches := emotePathRegex.FindStringSubmatch(url.Path)
	if len(matches) != 4 {
		return nil, errInvalidFrankerFaceZEmotePath
	}

	emoteHash := matches[1]

	apiResponse := emoteCache.Get(emoteHash, r)
	return json.Marshal(apiResponse)

}
