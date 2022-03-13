package resolver

import (
	"net/http"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

func TestErrorf(t *testing.T) {
	c := qt.New(t)

	tests := []struct {
		label            string
		format           string
		args             []interface{}
		expectedResponse *Response
		expectedDuration time.Duration
		expectedError    error
	}{
		{
			"normal",
			"error",
			[]interface{}{},
			&Response{
				Status:  http.StatusInternalServerError,
				Message: "error",
			},
			NoSpecialDur,
			nil,
		},
		{
			"args",
			"error: %s",
			[]interface{}{"hello"},
			&Response{
				Status:  http.StatusInternalServerError,
				Message: "error: hello",
			},
			NoSpecialDur,
			nil,
		},
		{
			"html",
			"<b>error</b>",
			[]interface{}{},
			&Response{
				Status:  http.StatusInternalServerError,
				Message: "&lt;b&gt;error&lt;/b&gt;",
			},
			NoSpecialDur,
			nil,
		},
	}

	for _, t := range tests {
		c.Run(t.label, func(c *qt.C) {
			response, duration, err := Errorf(t.format, t.args...)
			c.Assert(response, qt.DeepEquals, t.expectedResponse)
			c.Assert(duration, qt.Equals, t.expectedDuration)
			c.Assert(err, qt.Equals, t.expectedError)
		})
	}
}
