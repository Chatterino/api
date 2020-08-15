package main

import (
	"bytes"
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

func insertCommas(str string, n int) string {
	var buffer bytes.Buffer
	var remainder = n - 1
	var lenght = len(str) - 2
	for i, rune := range str {
		buffer.WriteRune(rune)
		if (lenght-i)%n == remainder {
			buffer.WriteRune(',')
		}
	}
	return buffer.String()
}

func formatDate(format string, str string) string {
	date, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return ""
	} else {
		return date.Format(format)
	}
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
