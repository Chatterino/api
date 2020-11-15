package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
)

func FormatThumbnailURL(baseURL string, r *http.Request, urlString string) string {
	if baseURL == "" {
		scheme := "https://"
		if r.TLS == nil {
			scheme = "http://" // https://github.com/golang/go/issues/28940#issuecomment-441749380
		}
		return fmt.Sprintf("%s%s/thumbnail/%s", scheme, r.Host, url.QueryEscape(urlString))
	}

	return fmt.Sprintf("%s/thumbnail/%s", strings.TrimSuffix(baseURL, "/"), url.QueryEscape(urlString))
}

func UnescapeURLArgument(r *http.Request, key string) (string, error) {
	vars := mux.Vars(r)
	escapedURL := vars[key]
	url, err := url.PathUnescape(escapedURL)
	if err != nil {
		return "", err
	}

	return url, nil
}
