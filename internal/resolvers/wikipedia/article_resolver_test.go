package wikipedia

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

func TestArticleResolver(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

	pool, _ := pgxmock.NewPool()

	cfg := config.APIConfig{}
	ts := testServer()
	defer ts.Close()
	apiURL := ts.URL + "/api/rest_v1/page/summary/%s/%s"

	r := NewArticleResolver(ctx, cfg, pool, apiURL)

	c.Assert(r, qt.IsNotNil)

	c.Run("Name", func(c *qt.C) {
		c.Assert(r.Name(), qt.Equals, "wikipedia:article")
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
				input:    utils.MustParseURL("https://wikipedia.org/wiki/ArticleID"),
				expected: true,
			},
			{
				label:    "Matching domain, WWW",
				input:    utils.MustParseURL("https://www.wikipedia.org/wiki/ArticleID"),
				expected: true,
			},
			{
				label:    "Matching domain, English",
				input:    utils.MustParseURL("https://en.wikipedia.org/wiki/ArticleID"),
				expected: true,
			},
			{
				label:    "Matching domain, German",
				input:    utils.MustParseURL("https://de.wikipedia.org/wiki/Gurke"),
				expected: true,
			},
			{
				label:    "Matching domain, non-matching path",
				input:    utils.MustParseURL("https://wikipedia.org/bad"),
				expected: false,
			},
			{
				label:    "Non-matching domain",
				input:    utils.MustParseURL("https://example.com/wiki/ArticleID"),
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
		c.Run("Error", func(c *qt.C) {
			type runTest struct {
				label         string
				inputURL      *url.URL
				expectedError error
				rowsReturned  int
			}

			tests := []runTest{
				// {
				// 	label:          "Emote",
				// 	inputURL:       utils.MustParseURL("https://betterttv.com/emotes/566ca04265dbbdab32ec054b"),
				// 	inputReq:       nil,
				// 	expectedBytes:  []byte(`{"status":200,"thumbnail":"https://cdn.betterttv.net/emote/566ca04265dbbdab32ec054b/3x","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%3Cb%3EKKona%3C%2Fb%3E%3Cbr%3E%3Cb%3EGlobal%20BetterTTV%20Emote%3C%2Fb%3E%3Cbr%3E%3Cb%3EBy:%3C%2Fb%3E%20zneix%3C%2Fdiv%3E"}`),
				// 	expectedError:  nil,
				// },
				// {
				// 	label:         "404",
				// 	inputURL:      utils.MustParseURL("https://en.wikipedia.org/wiki/404"),
				// 	expectedError: resolver.ErrDontHandle,
				// },
				// {
				// 	label:          "Bad JSON",
				// 	inputURL:       utils.MustParseURL("https://betterttv.com/emotes/bad"),
				// 	inputReq:       nil,
				// 	expectedBytes:  []byte(`{"status":500,"message":"betterttv api unmarshal error: invalid character \u0026#39;x\u0026#39; looking for beginning of value"}`),
				// 	expectedError:  nil,
				// },
			}

			const q = `SELECT value FROM cache WHERE key=$1`

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
					ctx, checkResult := r.Check(ctx, test.inputURL)
					c.Assert(checkResult, qt.IsTrue)
					outputBytes, outputError := r.Run(ctx, test.inputURL, nil)
					c.Assert(outputError, qt.Equals, test.expectedError)
					c.Assert(outputBytes, qt.IsNil)
				})
			}
		})
		c.Run("Not cached", func(c *qt.C) {
			type runTest struct {
				label         string
				inputURL      *url.URL
				inputReq      *http.Request
				expectedBytes []byte
				expectedError error
				rowsReturned  int
			}

			tests := []runTest{
				// {
				// 	label:          "Emote",
				// 	inputURL:       utils.MustParseURL("https://betterttv.com/emotes/566ca04265dbbdab32ec054b"),
				// 	inputReq:       nil,
				// 	expectedBytes:  []byte(`{"status":200,"thumbnail":"https://cdn.betterttv.net/emote/566ca04265dbbdab32ec054b/3x","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%3Cb%3EKKona%3C%2Fb%3E%3Cbr%3E%3Cb%3EGlobal%20BetterTTV%20Emote%3C%2Fb%3E%3Cbr%3E%3Cb%3EBy:%3C%2Fb%3E%20zneix%3C%2Fdiv%3E"}`),
				// 	expectedError:  nil,
				// },
				{
					label:         "404",
					inputURL:      utils.MustParseURL("https://wikipedia.org/wiki/404"),
					inputReq:      nil,
					expectedBytes: []byte(`{"status":404,"message":"No Wikipedia article found"}`),
					expectedError: nil,
				},
				{
					label:         "Normal page (HTML)",
					inputURL:      utils.MustParseURL("https://wikipedia.org/wiki/test_html"),
					inputReq:      nil,
					expectedBytes: []byte(`{"status":200,"tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%3Cb%3E\u0026lt%3Bb\u0026gt%3BTest%20title\u0026lt%3B%2Fb\u0026gt%3B\u0026nbsp%3B%E2%80%A2\u0026nbsp%3B\u0026lt%3Bb\u0026gt%3BTest%20description\u0026lt%3B%2Fb\u0026gt%3B%3C%2Fb%3E%3Cbr%3E\u0026lt%3Bb\u0026gt%3BTest%20extract\u0026lt%3B%2Fb\u0026gt%3B%3C%2Fdiv%3E"}`),
					expectedError: nil,
				},
				{
					label:         "Normal page (No description)",
					inputURL:      utils.MustParseURL("https://wikipedia.org/wiki/test_no_description"),
					inputReq:      nil,
					expectedBytes: []byte(`{"status":200,"tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%3Cb%3ETest%20title%3C%2Fb%3E%3Cbr%3ETest%20extract%3C%2Fdiv%3E"}`),
					expectedError: nil,
				},
				// {
				// 	label:          "Bad JSON",
				// 	inputURL:       utils.MustParseURL("https://betterttv.com/emotes/bad"),
				// 	inputReq:       nil,
				// 	expectedBytes:  []byte(`{"status":500,"message":"betterttv api unmarshal error: invalid character \u0026#39;x\u0026#39; looking for beginning of value"}`),
				// 	expectedError:  nil,
				// },
			}

			const q = `SELECT value FROM cache WHERE key=$1`

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
					pool.ExpectExec("INSERT INTO cache").
						WithArgs(pgxmock.AnyArg(), test.expectedBytes, pgxmock.AnyArg()).
						WillReturnResult(pgxmock.NewResult("INSERT", 1))
					ctx, checkResult := r.Check(ctx, test.inputURL)
					c.Assert(checkResult, qt.IsTrue)
					outputBytes, outputError := r.Run(ctx, test.inputURL, test.inputReq)
					c.Assert(outputError, qt.Equals, test.expectedError)
					c.Assert(outputBytes, qt.DeepEquals, test.expectedBytes)
				})
			}
		})
	})
}
