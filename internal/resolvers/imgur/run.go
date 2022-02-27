package imgur

import (
	"net/http"
	"net/url"
)

func run(url *url.URL, r *http.Request) ([]byte, error) {
	return imgurCache.Get(url.String(), r)
}
