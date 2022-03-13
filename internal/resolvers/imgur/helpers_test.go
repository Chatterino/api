package imgur

import (
	"net/http"
	"testing"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	qt "github.com/frankban/quicktest"
)

func TestInternalServerRror(t *testing.T) {
	c := qt.New(t)

	type tTest struct {
		label            string
		input            string
		expectedResponse *resolver.Response
		expectedDuration time.Duration
		expectedError    error
	}

	tests := []tTest{
		{
			label: "html",
			input: "error message <b>xD</b>",
			expectedResponse: &resolver.Response{
				Status:  http.StatusInternalServerError,
				Message: "imgur resolver error: error message &lt;b&gt;xD&lt;/b&gt;",
			},
			expectedDuration: cache.NoSpecialDur,
			expectedError:    nil,
		},
	}

	for _, test := range tests {
		c.Run(test.label, func(c *qt.C) {
			response, duration, err := internalServerError(test.input)
			c.Assert(response, qt.DeepEquals, test.expectedResponse)
			c.Assert(duration, qt.Equals, test.expectedDuration)
			c.Assert(err, qt.Equals, test.expectedError)
		})
	}
}

func TestBuildTooltip(t *testing.T) {
	c := qt.New(t)

	type tTest struct {
		label            string
		input            miniImage
		expectedResponse *resolver.Response
		expectedDuration time.Duration
		expectedError    error
	}

	tests := []tTest{
		{
			label: "empty image",
			input: miniImage{},
			expectedResponse: &resolver.Response{
				Status:  http.StatusOK,
				Tooltip: "%3Cdiv%20style=%22text-align:%20left%3B%22%3E%3Cli%3E%3Cb%3EUploaded:%3C%2Fb%3E%20%3C%2Fli%3E%3C%2Fdiv%3E",
			},
			expectedDuration: cache.NoSpecialDur,
			expectedError:    nil,
		},
	}

	for _, test := range tests {
		c.Run(test.label, func(c *qt.C) {
			response, duration, err := buildTooltip(test.input)
			c.Assert(response, qt.DeepEquals, test.expectedResponse)
			c.Assert(duration, qt.Equals, test.expectedDuration)
			c.Assert(err, qt.Equals, test.expectedError)
		})
	}
}
