package wikipedia

import (
	"net/http"
	"net/url"
)

func run(url *url.URL, r *http.Request) ([]byte, error) {
	return wikipediaCache.Get(url.String(), r)
}
