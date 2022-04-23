package utils

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestTruncateString(t *testing.T) {
	type tType struct {
		input          string
		maxLength      int
		expectedOutput string
	}

	tests := []tType{
		{
			input:          "foobar",
			maxLength:      4,
			expectedOutput: "foo…",
		},
		{
			input:          "foobar",
			maxLength:      6,
			expectedOutput: "foobar",
		},
		{
			input:          "foobar",
			maxLength:      7,
			expectedOutput: "foobar",
		},
		{
			input:          "foobar",
			maxLength:      8,
			expectedOutput: "foobar",
		},
		{ // cut off on space
			input:          "foo bar",
			maxLength:      5,
			expectedOutput: "foo…",
		},
		{ // unicode
			input:          "⣿⣿⣿⣿⣿⣿⠿⠟⠛⠋⠉⠉⠉⠉⠉⠉⠉⠩⠭⠝⣛⠻⠿⣿⣿⣿⣿⣿ ⣿⣿⣿⣿⠏⣴⣶⣀⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⣹⣷⡌⣿⣿⣿⣿ ⣿⣿⣿⡿⠘⠻⠿⢿⣿⣷⣶⣶⣶⣶⣶⣶⣶⣶⣶⣿⣿⠿⠟⢃⡛⢿⣿⣿ ⣿⣿⠏⣴⣿⣿⣷⣶⣦⣬⣙⣛⣛⣛⣛⣛⣛⣛⣛⣉⣥⣾⣿⣿⣿⡧⢻⣿ ⣿⡏⢸⣿⣿⠟⠛⠿⣿⣿⣿⣿⣿⡿⠟⠛⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠸⣿ ⣿⣷⠘⠋⠄⠄⠄⠄⠙⣿⠿⠛⠁⠄⠄⠄⠄⠹⣿⣿⣿⣿⣿⣿⣿⣿⡇⣿ ⣿⡇⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠘⣿⣿⣿⣿⣿⣿⣿⣷⢸ ⣿⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠈⠛⠻⣿⣿⣿⣿⣿⢸ ⣿⡄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠈⢻⣿⣿⣿⢸ ⣿⣷⡀⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠹⣿⠏⣼ ⣿⣿⣿⣄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⣠⣍⣸⣿ ⣿⣿⣿⣿⣇⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⢰⣿⣿⣿⣿ ⣿⣿⣿⣿⣿⣷⣄⡀⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⠄⢀⣠⣴⣿⣿⣿⣿⣿ ⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣶⣦⣤⣤⣤⣤⣴⣶⣶⣾⣿⣿⣿⣿⣿⣿⣿⣿ HONEYDETECTED",
			maxLength:      5,
			expectedOutput: "⣿⣿⣿⣿…",
		},
	}

	for _, test := range tests {
		output := TruncateString(test.input, test.maxLength)
		if output != test.expectedOutput {
			t.Fatalf("got output '%s', expected '%s'", output, test.expectedOutput)
		}
	}
}

func TestStringPtr(t *testing.T) {
	c := qt.New(t)
	type tTest struct {
		input string
	}

	tests := []tTest{
		{
			input: "s",
		},
		{
			input: "",
		},
		{
			input: " ",
		},
		{
			input: "😂",
		},
	}

	for _, test := range tests {
		c.Run(test.input, func(c *qt.C) {
			output := StringPtr(test.input)
			c.Assert(output, qt.IsNotNil)
			c.Assert(*output, qt.Equals, test.input)
		})
	}
}
