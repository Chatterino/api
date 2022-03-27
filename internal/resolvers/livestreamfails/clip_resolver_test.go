package livestreamfails

import (
	"context"
	"net/url"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/utils"
	qt "github.com/frankban/quicktest"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
)

func TestClipResolver(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

	pool, _ := pgxmock.NewPool()

	cfg := config.APIConfig{}
	ts := testServer()
	defer ts.Close()
	apiURLFormat := ts.URL + "/clip/%s"

	r := NewClipResolver(ctx, cfg, pool, apiURLFormat)

	c.Assert(r, qt.IsNotNil)

	c.Run("Name", func(c *qt.C) {
		c.Assert(r.Name(), qt.Equals, "livestreamfails:clip")
	})

	c.Run("Check", func(c *qt.C) {
		type checkTest struct {
			label    string
			input    *url.URL
			expected bool
		}

		tests := []checkTest{
			{
				label:    "Matching domain, no WWW",
				input:    utils.MustParseURL("https://livestreamfails.com/clip/123"),
				expected: true,
			},
			{
				label:    "Matching domain, WWW",
				input:    utils.MustParseURL("https://www.livestreamfails.com/clip/123"),
				expected: true,
			},
			{
				label:    "Matching domain, non-matching path",
				input:    utils.MustParseURL("https://livestreamfails.com/categories"),
				expected: false,
			},
			{
				label:    "Matching domain, non-matching path",
				input:    utils.MustParseURL("https://livestreamfails.com/clip/"),
				expected: false,
			},
			{
				label:    "Non-matching domain",
				input:    utils.MustParseURL("https://example.com/emotes/566ca04265dbbdab32ec054a"),
				expected: false,
			},
		}

		for _, test := range tests {
			c.Run(test.label, func(c *qt.C) {
				_, output := r.Check(ctx, test.input)
				c.Assert(output, qt.Equals, test.expected)
			})
		}
	})

	c.Run("Run", func(c *qt.C) {
		c.Run("Missing context value", func(c *qt.C) {
			type runTest struct {
				label         string
				inputURL      *url.URL
				inputClipID   *string
				expectedError error
			}

			tests := []runTest{
				{
					label:         "Missing clip ID",
					inputURL:      utils.MustParseURL("https://livestreamfails.com/clip/"),
					inputClipID:   nil,
					expectedError: errMissingClipID,
				},
			}

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					// pool.ExpectExec("INSERT INTO cache").
					// 	WithArgs("livestreamfails:clip:"+test.inputClipID, test.expectedBytes, pgxmock.AnyArg()).
					// 	WillReturnResult(pgxmock.NewResult("INSERT", 1))
					outputBytes, outputError := r.Run(ctx, test.inputURL, nil)
					c.Assert(outputError, qt.Equals, test.expectedError)
					c.Assert(outputBytes, qt.IsNil)
					c.Assert(pool.ExpectationsWereMet(), qt.IsNil)
				})
			}
		})
		c.Run("Not cached", func(c *qt.C) {
			type runTest struct {
				label         string
				inputURL      *url.URL
				inputClipID   string
				expectedBytes []byte
				expectedError error
				rowsReturned  int
			}

			tests := []runTest{
				{
					label:         "normal",
					inputURL:      utils.MustParseURL("https://livestreamfails.com/clip/123"),
					inputClipID:   "123",
					expectedBytes: []byte(`{"status":200,"thumbnail":"https://livestreamfails-image-prod.b-cdn.net/image/asd","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%0A%3Cb%3EClip%20Label%3C%2Fb%3E%3Chr%3E%0A%3Cb%3EStreamer:%3C%2Fb%3E%20Streamer%20Label%3Cbr%3E%0A%3Cb%3ECategory:%3C%2Fb%3E%20Category%20Label%3Cbr%3E%0A%3Cb%3EPlatform:%3C%2Fb%3E%20Twitch%3Cbr%3E%0A%3Cb%3EReddit%20score:%3C%2Fb%3E%2069%3Cbr%3E%0A%3Cb%3ECreated:%3C%2Fb%3E%2027%20Mar%202022%0A%3C%2Fdiv%3E"}`),
					expectedError: nil,
				},
				{
					label:         "normal NSFW",
					inputURL:      utils.MustParseURL("https://livestreamfails.com/clip/905"),
					inputClipID:   "905",
					expectedBytes: []byte(`{"status":200,"tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cli%3E%3Cb%3E%3Cspan%20style=%22color:%20red%22%3ENSFW%3C%2Fspan%3E%3C%2Fb%3E%3C%2Fli%3E%0A%3Cb%3EClip%20Label%3C%2Fb%3E%3Chr%3E%0A%3Cb%3EStreamer:%3C%2Fb%3E%20Streamer%20Label%3Cbr%3E%0A%3Cb%3ECategory:%3C%2Fb%3E%20Category%20Label%3Cbr%3E%0A%3Cb%3EPlatform:%3C%2Fb%3E%20Twitch%3Cbr%3E%0A%3Cb%3EReddit%20score:%3C%2Fb%3E%2069%3Cbr%3E%0A%3Cb%3ECreated:%3C%2Fb%3E%2027%20Mar%202022%0A%3C%2Fdiv%3E"}`),
					expectedError: nil,
				},
				{
					label:         "404",
					inputURL:      utils.MustParseURL("https://livestreamfails.com/clip/404"),
					inputClipID:   "404",
					expectedBytes: []byte(`{"status":404,"message":"No LivestreamFails Clip with this ID found"}`),
					expectedError: nil,
				},
				{
					label:         "500",
					inputURL:      utils.MustParseURL("https://livestreamfails.com/clip/500"),
					inputClipID:   "500",
					expectedBytes: []byte(`{"status":500,"message":"Livestreamfails unhandled HTTP status code: 500"}`),
					expectedError: nil,
				},
				{
					label:         "bad json",
					inputURL:      utils.MustParseURL("https://livestreamfails.com/clip/666"),
					inputClipID:   "666",
					expectedBytes: []byte(`{"status":500,"message":"Livestreamfails API response decode error: invalid character \u0026#39;x\u0026#39; looking for beginning of value"}`),
					expectedError: nil,
				},
			}

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
					pool.ExpectExec("INSERT INTO cache").
						WithArgs("livestreamfails:clip:"+test.inputClipID, test.expectedBytes, pgxmock.AnyArg()).
						WillReturnResult(pgxmock.NewResult("INSERT", 1))
					ctx, checkResult := r.Check(ctx, test.inputURL)
					c.Assert(checkResult, qt.IsTrue)
					outputBytes, outputError := r.Run(ctx, test.inputURL, nil)
					c.Assert(outputError, qt.Equals, test.expectedError)
					c.Assert(outputBytes, qt.DeepEquals, test.expectedBytes)
					c.Assert(pool.ExpectationsWereMet(), qt.IsNil)
				})
			}
		})
	})
}
