package imgur

import (
	"context"
	"net/url"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/resolver"
	qt "github.com/frankban/quicktest"
)

func testCheck(ctx context.Context, resolver resolver.Resolver, c *qt.C, urlString string) bool {
	u, err := url.Parse(urlString)
	c.Assert(u, qt.Not(qt.IsNil))
	c.Assert(err, qt.IsNil)

	_, result := resolver.Check(ctx, u)

	return result
}

func TestCheck(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

	resolver := &Resolver{}

	shouldCheck := []string{
		"https://imgur.com",
		"https://www.imgur.com",
		"https://i.imgur.com",
	}

	for _, u := range shouldCheck {
		c.Assert(testCheck(ctx, resolver, c, u), qt.IsTrue)
	}

	shouldNotCheck := []string{
		"https://imgurr.com",
		"https://www.imgur.bad.com",
		"https://iimgur.com",
		"https://google.com",
		"https://i.imgur.org",
	}

	for _, u := range shouldNotCheck {
		c.Assert(testCheck(ctx, resolver, c, u), qt.IsFalse)
	}
}
