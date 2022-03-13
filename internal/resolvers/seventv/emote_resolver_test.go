package seventv

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/utils"
	qt "github.com/frankban/quicktest"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
)

func TestEmoteResolver(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

	pool, _ := pgxmock.NewPool()

	ts := testServer()
	defer ts.Close()
	cfg := config.APIConfig{
		BaseURL: "https://example.com/chatterino/",
	}
	apiURL := utils.MustParseURL(ts.URL + "/v2/gql")

	resolver := NewEmoteResolver(ctx, cfg, pool, apiURL)

	c.Assert(resolver, qt.IsNotNil)

	c.Run("Name", func(c *qt.C) {
		c.Assert(resolver.Name(), qt.Equals, "seventv:emote")
	})

	c.Run("Check", func(c *qt.C) {
		type checkTest struct {
			label    string
			input    *url.URL
			expected bool
		}

		tests := []checkTest{
			{
				label:    "Matching domain",
				input:    utils.MustParseURL("https://7tv.app/emotes/604281c81ae70f000d47ffd9"),
				expected: true,
			},
			{
				label:    "Matching domain, non-matching path",
				input:    utils.MustParseURL("https://7tv.app/users/60bca831e7ecd2f892c9b9ab"),
				expected: false,
			},
			{
				label:    "Non-matching domain",
				input:    utils.MustParseURL("https://example.com/emotes/604281c81ae70f000d47ffd9"),
				expected: false,
			},
		}

		for _, test := range tests {
			c.Run(test.label, func(c *qt.C) {
				output := resolver.Check(ctx, test.input)
				c.Assert(output, qt.Equals, test.expected)
			})
		}
	})

	c.Run("Run", func(c *qt.C) {
		c.Run("Error", func(c *qt.C) {
			type runTest struct {
				label          string
				inputURL       *url.URL
				inputEmoteHash string
				inputReq       *http.Request
				expectedError  error
			}

			tests := []runTest{
				{
					label:         "Non-matching link",
					inputURL:      utils.MustParseURL("https://betterttv.com/user/566ca04265dbbdab32ec054a"),
					expectedError: errInvalidSevenTVEmotePath,
				},
			}

			const q = `SELECT value FROM cache WHERE key=$1`

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					outputBytes, outputError := resolver.Run(ctx, test.inputURL, test.inputReq)
					c.Assert(outputError, qt.Equals, test.expectedError)
					c.Assert(outputBytes, qt.IsNil)
				})
			}
		})
		c.Run("Cached", func(c *qt.C) {
			type runTest struct {
				label          string
				inputURL       *url.URL
				inputEmoteHash string
				inputReq       *http.Request
				// expectedBytes will be returned from the cache, and expected to be returned in the same form
				expectedBytes []byte
				expectedError error
			}

			tests := []runTest{
				{
					label:          "Matching link - cached",
					inputURL:       utils.MustParseURL("https://betterttv.com/emotes/566ca04265dbbdab32ec054a"),
					inputEmoteHash: "566ca04265dbbdab32ec054a",
					inputReq:       nil,
					expectedBytes:  []byte(`{"status":200,"thumbnail":"https://cdn.betterttv.net/emote/566ca04265dbbdab32ec054a/3x","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%3Cb%3EKKona%3C%2Fb%3E%3Cbr%3E%3Cb%3EGlobal%20BetterTTV%20Emote%3C%2Fb%3E%3Cbr%3E%3Cb%3EBy:%3C%2Fb%3E%20NightDev%3C%2Fdiv%3E"}`),
					expectedError:  nil,
				},
				{
					label:          "Matching link - cached 2",
					inputURL:       utils.MustParseURL("https://betterttv.com/emotes/566ca04265dbbdab32ec054a"),
					inputEmoteHash: "566ca04265dbbdab32ec054a",
					inputReq:       nil,
					expectedBytes:  []byte(`{"status":200,"thumbnail":"https://cdn.betterttv.net/emote/566ca04265dbbdab32ec054a/3x","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%3Cb%3EKKona%3C%2Fb%3E%3Cbr%3E%3Cb%3EGlobal%20BetterTTV%20Emote%3C%2Fb%3E%3Cbr%3E%3Cb%3EBy:%3C%2Fb%3E%20NightDev%3C%2Fdiv%3E"}`),
					expectedError:  nil,
				},
				{
					label:          "Matching link - 404",
					inputURL:       utils.MustParseURL("https://betterttv.com/emotes/404"),
					inputEmoteHash: "404",
					inputReq:       nil,
					expectedBytes:  []byte(`{"status":404,"message":"No BetterTTV emote with this hash found"}`),
					expectedError:  nil,
				},
			}

			const q = `SELECT value FROM cache WHERE key=$1`

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					rows := pgxmock.NewRows([]string{"value"}).AddRow(test.expectedBytes)
					pool.ExpectQuery("SELECT").
						WithArgs("seventv:emote:" + test.inputEmoteHash).
						WillReturnRows(rows)
					outputBytes, outputError := resolver.Run(ctx, test.inputURL, test.inputReq)
					c.Assert(outputError, qt.Equals, test.expectedError)
					c.Assert(outputBytes, qt.DeepEquals, test.expectedBytes)
				})
			}
		})

		c.Run("Not cached", func(c *qt.C) {
			type runTest struct {
				label          string
				inputURL       *url.URL
				inputEmoteHash string
				inputReq       *http.Request
				expectedBytes  []byte
				expectedError  error
				rowsReturned   int
			}

			tests := []runTest{
				{
					label:          "Private",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/604281c81ae70f000d47ffd9"),
					inputEmoteHash: "604281c81ae70f000d47ffd9",
					inputReq:       nil,
					expectedBytes:  []byte(`{"status":200,"thumbnail":"https://example.com/chatterino/thumbnail/https%3A%2F%2Fcdn.7tv.app%2Femote%2F604281c81ae70f000d47ffd9%2F4x","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EPajawalk%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EPrivate%20SevenTV%20Emote%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EBy:%3C%2Fb%3E%20durado_%0A%3C%2Fdiv%3E","link":"https://7tv.app/emotes/604281c81ae70f000d47ffd9"}`),
					expectedError:  nil,
				},
				{
					label:          "Hidden",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/60ae8d9ff39a7552b658b60d"),
					inputEmoteHash: "60ae8d9ff39a7552b658b60d",
					inputReq:       nil,
					expectedBytes:  []byte(`{"status":200,"tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EBedge%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EPrivate%20SevenTV%20Emote%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EBy:%3C%2Fb%3E%20Paruna%0A%3Cli%3E%3Cb%3E%3Cspan%20style=%22color:%20red%3B%22%3EUNLISTED%3C%2Fspan%3E%3C%2Fb%3E%3C%2Fli%3E%0A%3C%2Fdiv%3E","link":"https://7tv.app/emotes/60ae8d9ff39a7552b658b60d"}`),
					expectedError:  nil,
				},
				{
					label:          "Global",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/6042998c1d4963000d9dae34"),
					inputEmoteHash: "6042998c1d4963000d9dae34",
					inputReq:       nil,
					expectedBytes:  []byte(`{"status":200,"thumbnail":"https://example.com/chatterino/thumbnail/https%3A%2F%2Fcdn.7tv.app%2Femote%2F6042998c1d4963000d9dae34%2F4x","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EFeelsOkayMan%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EGlobal%20SevenTV%20Emote%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EBy:%3C%2Fb%3E%20MegaKill3%0A%3C%2Fdiv%3E","link":"https://7tv.app/emotes/6042998c1d4963000d9dae34"}`),
					expectedError:  nil,
				},
				{
					label:          "Missing visibility",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/603cb219c20d020014423c34"),
					inputEmoteHash: "603cb219c20d020014423c34",
					inputReq:       nil,
					expectedBytes:  []byte(`{"status":200,"thumbnail":"https://example.com/chatterino/thumbnail/https%3A%2F%2Fcdn.7tv.app%2Femote%2F603cb219c20d020014423c34%2F4x","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EmonkaE%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EShared%20SevenTV%20Emote%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EBy:%3C%2Fb%3E%20Zhark%0A%3C%2Fdiv%3E","link":"https://7tv.app/emotes/603cb219c20d020014423c34"}`),
					expectedError:  nil,
				},
				{
					label:          "Matching link - 404",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/404"),
					inputEmoteHash: "404",
					inputReq:       nil,
					expectedBytes:  []byte(`{"status":404,"message":"No SevenTV emote with this id found"}`),
					expectedError:  nil,
				},
			}

			const q = `SELECT value FROM cache WHERE key=$1`

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
					pool.ExpectExec("INSERT INTO cache").
						WithArgs("seventv:emote:"+test.inputEmoteHash, test.expectedBytes, pgxmock.AnyArg()).
						WillReturnResult(pgxmock.NewResult("INSERT", 1))
					outputBytes, outputError := resolver.Run(ctx, test.inputURL, test.inputReq)
					c.Assert(outputError, qt.Equals, test.expectedError)
					c.Assert(outputBytes, qt.DeepEquals, test.expectedBytes)
				})
			}
		})
	})
}
