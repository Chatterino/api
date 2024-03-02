package humanize_test

import (
	"testing"

	"github.com/Chatterino/api/pkg/humanize"
	qt "github.com/frankban/quicktest"
)

func TestNumber(t *testing.T) {
	c := qt.New(t)
	type testCase struct {
		input    uint64
		expected string
	}
	cases := []testCase{
		{0, "0"},
		{1, "1"},
		{100, "100"},
		{1_000, "1,000"},
		{1_000_000, "1.0M"},
		{1_500_000, "1.5M"},
		{1_550_000, "1.6M"},
		{1_555_000, "1.6M"},
	}

	for _, tc := range cases {
		c.Run("", func(c *qt.C) {
			res := humanize.Number(tc.input)
			c.Assert(res, qt.Equals, tc.expected)
		})
	}
}

func TestNumberInt64(t *testing.T) {
	c := qt.New(t)
	type testCase struct {
		input    int64
		expected string
	}
	cases := []testCase{
		{0, "0"},
		{1, "1"},
		{100, "100"},
		{1_000, "1,000"},
		{1_000_000, "1.0M"},
		{1_500_000, "1.5M"},
		{1_550_000, "1.6M"},
		{1_555_000, "1.6M"},
	}

	for _, tc := range cases {
		c.Run("", func(c *qt.C) {
			res := humanize.NumberInt64(tc.input)
			c.Assert(res, qt.Equals, tc.expected)
		})
	}
}
