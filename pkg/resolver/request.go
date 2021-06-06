package resolver

import (
	"net/http"
	"strings"
	"time"
)

var (
	httpClient = &http.Client{
		Timeout: 15 * time.Second,
	}
)

func RequestGET(url string) (response *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// ensures websites return pages in english (e.g. twitter would return french preview
	// when the request came from a french IP.)
	req.Header.Add("Accept-Language", "en-US, en;q=0.9, *;q=0.5")
	req.Header.Set("User-Agent", "chatterino-api-cache/1.0 link-resolver")

	return httpClient.Do(req)
}

func RequestGETWithHeaders(url string, extraHeaders map[string]string) (response *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// ensures websites return pages in english (e.g. twitter would return french preview
	// when the request came from a french IP.)
	req.Header.Add("Accept-Language", "en-US, en;q=0.9, *;q=0.5")
	req.Header.Set("User-Agent", "chatterino-api-cache/1.0 link-resolver")

	for headerKey, headerValue := range extraHeaders {
		req.Header.Set(headerKey, headerValue)
	}

	return httpClient.Do(req)
}

func RequestPOST(url, body string) (response *http.Response, err error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return httpClient.Do(req)
}

func HTTPClient() *http.Client {
	return httpClient
}
