package main

import (
	"testing"
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
			expectedOutput: "fooba…",
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
		output := truncateString(test.input, test.maxLength)
		if output != test.expectedOutput {
			t.Fatalf("got output '%s', expected '%s'", output, test.expectedOutput)
		}
	}
}
