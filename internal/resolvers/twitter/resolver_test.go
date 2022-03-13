package twitter

import (
	"context"
	"net/url"
	"testing"

	"github.com/Chatterino/api/internal/logger"
)

func TestShouldMatch(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())

	tests := []string{
		"https://twitter.com/pajlada",
		"http://www.twitter.com/forsen",
	}

	resolver := &TwitterResolver{}

	for _, test := range tests {
		u, err := url.Parse(test)
		if err != nil {
			t.Fatalf("invalid url %s", test)
		}

		if !resolver.Check(ctx, u) {
			t.Fatalf("url %s didn't match twitter check while it should have", test)
		}
	}
}

func TestShouldNotMatch(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())

	tests := []string{
		"https://google.com",
		"https://twitter.com/compose",
		"https://twitter.com/logout",
		"https://twitter.com/logout",
		"https://nontwitter.com/forsen",
	}

	resolver := &TwitterResolver{}

	for _, test := range tests {
		u, err := url.Parse(test)
		if err != nil {
			t.Fatalf("invalid url %s", test)
		}

		if resolver.Check(ctx, u) {
			t.Fatalf("url %s matched twitter check while it shouldn't have", test)
		}
	}
}
