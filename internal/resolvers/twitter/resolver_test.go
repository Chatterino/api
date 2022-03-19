package twitter

import (
	"context"
	"net/url"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	qt "github.com/frankban/quicktest"
)

func TestShouldMatch(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

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

		_, result := resolver.Check(ctx, u)
		c.Assert(result, qt.IsTrue, qt.Commentf("url %s didn't match twitter check while it should have", test))
	}
}

func TestShouldNotMatch(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

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

		_, result := resolver.Check(ctx, u)
		c.Assert(result, qt.IsFalse, qt.Commentf("url %s matched twitter check while it shouldn't have", test))
	}
}
