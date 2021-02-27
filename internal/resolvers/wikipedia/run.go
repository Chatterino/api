package wikipedia

import (
	"encoding/json"
	"errors"
	"net/url"
)

func run(url *url.URL) ([]byte, error) {
	wikiResponse, ok := wikipediaCache.Get(url.String(), nil).(response)
	if !ok {
		return nil, errors.New("wikipedia cache load function is broken")
	}

	if wikiResponse.err != nil {
		return nil, wikiResponse.err
	}

	return json.Marshal(wikiResponse.resolverResponse)
}
