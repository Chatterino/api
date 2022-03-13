package utils

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestSetFromSlice(t *testing.T) {
	c := qt.New(t)
	type tTest struct {
		label    string
		input    []interface{}
		expected map[interface{}]struct{}
	}

	tests := []tTest{
		{
			label: "Set 1",
			input: []interface{}{"foo", "bar"},
			expected: map[interface{}]struct{}{
				"foo": {},
				"bar": {},
			},
		},
		{
			label: "Handle duplicate",
			input: []interface{}{"foo", "bar", "bar"},
			expected: map[interface{}]struct{}{
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
