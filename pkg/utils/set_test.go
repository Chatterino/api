package utils

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestSetFromSlice(t *testing.T) {
	c := qt.New(t)
	type tTest struct {
		label    string
		input    []any
		expected map[any]struct{}
	}

	tests := []tTest{
		{
			label: "Set 1",
			input: []any{"foo", "bar"},
			expected: map[any]struct{}{
				"foo": {},
				"bar": {},
			},
		},
		{
			label: "Handle duplicate",
			input: []any{"foo", "bar", "bar"},
			expected: map[any]struct{}{
				"foo": {},
				"bar": {},
			},
		},
	}

	for _, test := range tests {
		c.Run("", func(c *qt.C) {
			output := SetFromSlice(test.input)
			c.Assert(output, qt.DeepEquals, test.expected)
		})
	}
}
