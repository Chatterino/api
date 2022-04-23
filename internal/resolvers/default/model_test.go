package defaultresolver

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestTooltipData(t *testing.T) {
	c := qt.New(t)

	c.Run("Truncate", func(c *qt.C) {
		tests := []struct {
			input    tooltipData
			expected tooltipData
		}{
			{
				input: tooltipData{
					Title:       "foo",
					Description: "bar",
				},
				expected: tooltipData{
					Title:       "foo",
					Description: "bar",
				},
			},
			{
				input: tooltipData{
					Title:       "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxy",
					Description: "bar",
				},
				expected: tooltipData{
					Title:       "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxy",
					Description: "bar",
				},
			},
			{
				input: tooltipData{
					Title:       "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxyz",
					Description: "bar",
				},
				expected: tooltipData{
					Title:       "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx…",
					Description: "bar",
				},
			},
		}

		for _, test := range tests {
			c.Run("", func(c *qt.C) {
				test.input.Truncate()
				c.Assert(test.input, qt.DeepEquals, test.expected)
			})
		}
	})
}
