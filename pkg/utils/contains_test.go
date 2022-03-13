package utils

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestContains(t *testing.T) {
	c := qt.New(t)
	type tTest struct {
		haystack []string
		needle   string
		expected bool
	}

	tests := []tTest{
		{
			haystack: []string{"foo", "bar"},
			needle:   "foo",
			expected: true,
		},
		{
			haystack: []string{"foo", "bar"},
			needle:   "baz",
			expected: false,
		},
	}

	for _, test := range tests {
		c.Run("", func(c *qt.C) {
			output := Contains(test.haystack, test.needle)
			c.Assert(output, qt.Equals, test.expected)
		})
	}
}
