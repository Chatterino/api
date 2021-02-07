package imgur

import (
	"encoding/json"
	"errors"
	"net/url"
)

func run(url *url.URL) ([]byte, error) {
	imgurResponse, ok := imgurCache.Get(url.String(), nil).(response)
	if !ok {
		return nil, errors.New("imgur cache load function is broken")
	}

	if imgurResponse.err != nil {
		return nil, imgurResponse.err
	}

	return json.Marshal(imgurResponse.resolverResponse)
}
