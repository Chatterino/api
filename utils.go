package main

import (
	"encoding/json"
	"fmt"
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

func formatDuration(dur string) string {
	dur = strings.ToLower(dur)
	dur = strings.Replace(dur, "pt", "", 1)
	d, _ := time.ParseDuration(dur)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func formatDate(format string, str string) string {
	date, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return ""
	}
	return date.Format(format)
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
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
