package twitter

import (
	"net/url"
	"testing"
)

func TestShouldMatch(t *testing.T) {
	tests := []string{
		"https://twitter.com/pajlada",
		"http://www.twitter.com/forsen",
	}

	for _, test := range tests {
		u, err := url.Parse(test)
		if err != nil {
			t.Fatalf("invalid url %s", test)
		}

		if !check(u) {
			t.Fatalf("url %s didn't match twitter check while it should have", test)
		}
	}
}

func TestShouldNotMatch(t *testing.T) {
	tests := []string{
		"https://google.com",
		"https://twitter.com/compose",
		"https://twitter.com/logout",
		"https://twitter.com/logout",
		"https://nontwitter.com/forsen",
	}

	for _, test := range tests {
		u, err := url.Parse(test)
		if err != nil {
			t.Fatalf("invalid url %s", test)
		}

		if check(u) {
			t.Fatalf("url %s matched twitter check while it shouldn't have", test)
		}
	}
}
