package utils

import (
	"net/url"
	"testing"

	qt "github.com/frankban/quicktest"
)

func makeUrl(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}

func TestIsSubdomainOf(t *testing.T) {
	c := qt.New(t)
	type tTest struct {
		u *url.URL

		parents  []string
		expected bool
	}

	tests := []tTest{
		{
			u:        makeUrl("https://www.youtube.com/watch?v=aTts9CnsAv8"),
			parents:  []string{"youtube.com"},
			expected: true,
		},
		{
			u:        makeUrl("https://www.twitter.com/forsen"),
			parents:  []string{"youtube.com"},
			expected: false,
		},
		{
			u:        makeUrl("https://www.twitter.com/forsen"),
			parents:  []string{"youtube.com", "google.com"},
			expected: false,
		},
		{
			u:        makeUrl("https://www.twitter.com/forsen"),
			parents:  []string{"youtube.com", "twitter.com"},
			expected: true,
		},
		{
			u:        makeUrl("https://www.twitter.com/forsen"),
			parents:  []string{"twitter.com", "youtube.com"},
			expected: true,
		},
	}

	for _, test := range tests {
		c.Run(test.u.String(), func(c *qt.C) {
			actual := IsSubdomainOf(test.u, test.parents...)
			c.Assert(actual, qt.Equals, test.expected)
		})
	}

}
