package imgur

import (
	"net/url"
	"testing"

	qt "github.com/frankban/quicktest"
)

func testCheck(c *qt.C, urlString string) bool {
	u, err := url.Parse(urlString)
	c.Assert(u, qt.Not(qt.IsNil))
	c.Assert(err, qt.IsNil)

	return check(u)
}

func TestCheck(t *testing.T) {
	c := qt.New(t)

	shouldCheck := []string{
		"https://imgur.com",
		"https://www.imgur.com",
		"https://i.imgur.com",
	}

	for _, u := range shouldCheck {
		c.Assert(testCheck(c, u), qt.IsTrue)
	}

	shouldNotCheck := []string{
		"https://imgurr.com",
		"https://www.imgur.bad.com",
		"https://iimgur.com",
		"https://google.com",
		"https://i.imgur.org",
	}

	for _, u := range shouldNotCheck {
		c.Assert(testCheck(c, u), qt.IsFalse)
	}
}
