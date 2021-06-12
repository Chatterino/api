package imgur

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

func run(url *url.URL, r *http.Request) ([]byte, error) {
	imgurResponse, ok := imgurCache.Get(url.String(), r).(response)
	if !ok {
		return nil, errors.New("imgur cache load function is broken")
	}

	if imgurResponse.err != nil {
		return nil, imgurResponse.err
	}

	return json.Marshal(imgurResponse.resolverResponse)
}
