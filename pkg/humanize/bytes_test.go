package humanize_test

import (
	"testing"

	"github.com/Chatterino/api/pkg/humanize"
	qt "github.com/frankban/quicktest"
)

func TestBytes(t *testing.T) {
	c := qt.New(t)
	type testCase struct {
		input    uint64
		expected string
	}
	cases := []testCase{
		{0, "0.0 B"},
		{1, "1.0 B"},
		{1000, "1.0 KB"},
		{1001, "1.0 KB"},
		{1501, "1.5 KB"},
		{1234 * 1000, "1.2 MB"},
		{1234 * 1000 * 1000, "1.2 GB"},
		{1234 * 1000 * 1000 * 1000, "1.2 TB"},
		{1234 * 1000 * 1000 * 1000 * 1000, "1234.0 TB"},
	}

	for _, tc := range cases {
		c.Run("", func(c *qt.C) {
			res := humanize.Bytes(tc.input)
			c.Assert(res, qt.Equals, tc.expected)
		})
	}
}
