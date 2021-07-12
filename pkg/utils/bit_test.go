package utils

import "testing"

func TestHasBits(t *testing.T) {
	type tType struct {
		input          int32
		bit            int32
		expectedOutput bool
	}

	tests := []tType{
		{
			input:          0b0100,
			bit:            0b0100,
			expectedOutput: true,
		},
		{
			input:          0b1100,
			bit:            0b0100,
			expectedOutput: true,
		},
		{
			input:          0b0000,
			bit:            0b0100,
			expectedOutput: false,
		},
		{
			input:          0b1000,
			bit:            0b0100,
			expectedOutput: false,
		},
	}

	for _, test := range tests {
		output := HasBits(test.input, test.bit)
		if output != test.expectedOutput {
			t.Fatalf("got output '%v', expected '%v'", output, test.expectedOutput)
		}
	}

}
