package wikipedia

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

func run(url *url.URL, r *http.Request) ([]byte, error) {
	wikiResponse, ok := wikipediaCache.Get(url.String(), r).(response)
	if !ok {
		return nil, errors.New("wikipedia cache load function is broken")
	}

	if wikiResponse.err != nil {
		return nil, wikiResponse.err
	}

	return json.Marshal(wikiResponse.resolverResponse)
}
