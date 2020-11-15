package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func unescapeURLArgument(r *http.Request, key string) (string, error) {
	vars := mux.Vars(r)
	escapedURL := vars[key]
	url, err := url.PathUnescape(escapedURL)
	if err != nil {
		return "", err
	}

	return url, nil
}

func marshalNoDur(i interface{}) ([]byte, error, time.Duration) {
	data, err := json.Marshal(i)
	return data, err, noSpecialDur
}

// truncateString truncates string down to the maximum length with a unicode triple dot if truncation took place
func truncateString(s string, maxLength int) string {
	runes := []rune(s)
	if len(runes) < maxLength {
		return s
	}

	return strings.TrimSpace(string(runes[:maxLength-1])) + "…"
}
