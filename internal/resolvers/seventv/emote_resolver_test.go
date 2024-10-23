package seventv

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
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
	apiURL := utils.MustParseURL(ts.URL + "/v3/emotes")

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
				label:    "Matching domain (ulid)",
				input:    utils.MustParseURL("https://7tv.app/emotes/01F01WNXA00001NSRF006MFZYS"),
				expected: true,
			},
			{
				label:    "Matching domain, non-matching path (ulid)",
				input:    utils.MustParseURL("https://7tv.app/users/01F7GF1ZV8000EFV6JZ29CKEDB"),
				expected: false,
			},
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
				_, output := resolver.Check(ctx, test.input)
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
				// expectedResponse will be returned from the cache, and expected to be returned in the same form
				expectedResponse *cache.Response
				expectedError    error
			}

			tests := []runTest{
				{
					label:          "Matching link - cached",
					inputURL:       utils.MustParseURL("https://betterttv.com/emotes/566ca04265dbbdab32ec054a"),
					inputEmoteHash: "566ca04265dbbdab32ec054a",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://cdn.betterttv.net/emote/566ca04265dbbdab32ec054a/3x","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%3Cb%3EKKona%3C%2Fb%3E%3Cbr%3E%3Cb%3EGlobal%20BetterTTV%20Emote%3C%2Fb%3E%3Cbr%3E%3Cb%3EBy:%3C%2Fb%3E%20NightDev%3C%2Fdiv%3E"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Matching link - cached 2",
					inputURL:       utils.MustParseURL("https://betterttv.com/emotes/566ca04265dbbdab32ec054a"),
					inputEmoteHash: "566ca04265dbbdab32ec054a",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://cdn.betterttv.net/emote/566ca04265dbbdab32ec054a/3x","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%3Cb%3EKKona%3C%2Fb%3E%3Cbr%3E%3Cb%3EGlobal%20BetterTTV%20Emote%3C%2Fb%3E%3Cbr%3E%3Cb%3EBy:%3C%2Fb%3E%20NightDev%3C%2Fdiv%3E"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Matching link - 404",
					inputURL:       utils.MustParseURL("https://betterttv.com/emotes/404"),
					inputEmoteHash: "404",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":404,"message":"No BetterTTV emote with this hash found"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
			}

			const q = `SELECT value FROM cache WHERE key=$1`

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					rows := pgxmock.NewRows([]string{"value", "http_status_code", "http_content_type"}).AddRow(test.expectedResponse.Payload, test.expectedResponse.StatusCode, test.expectedResponse.ContentType)
					pool.ExpectQuery("SELECT").
						WithArgs("seventv:emote:" + test.inputEmoteHash).
						WillReturnRows(rows)
					outputBytes, outputError := resolver.Run(ctx, test.inputURL, test.inputReq)
					c.Assert(outputError, qt.Equals, test.expectedError)
					c.Assert(outputBytes, qt.DeepEquals, test.expectedResponse)
				})
			}
		})

		c.Run("Not cached", func(c *qt.C) {
			type runTest struct {
				label            string
				inputURL         *url.URL
				inputEmoteHash   string
				inputReq         *http.Request
				expectedResponse *cache.Response
				expectedError    error
				rowsReturned     int
			}

			tests := []runTest{
				{
					label:          "Private (ulid)",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/01F01WNXA00001NSRF006MFZYS"),
					inputEmoteHash: "01F01WNXA00001NSRF006MFZYS",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://example.com/chatterino/thumbnail/https%3A%2F%2Fcdn.7tv.app%2Femote%2F01F01WNXA00001NSRF006MFZYS%2Fbest.webp","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EPajawalk%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EPrivate%207TV%20Emote%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EBy:%3C%2Fb%3E%20durado_%0A%3C%2Fdiv%3E","link":"https://7tv.app/emotes/01F01WNXA00001NSRF006MFZYS"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Unlisted (ulid)",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/01F6MXJD8R000F76KNAAV5HDGD"),
					inputEmoteHash: "01F6MXJD8R000F76KNAAV5HDGD",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EBedge%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EShared%207TV%20Emote%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EBy:%3C%2Fb%3E%20Paruna%0A%3Cli%3E%3Cb%3E%3Cspan%20style=%22color:%20red%3B%22%3EUNLISTED%3C%2Fspan%3E%3C%2Fb%3E%3C%2Fli%3E%0A%3C%2Fdiv%3E","link":"https://7tv.app/emotes/01F6MXJD8R000F76KNAAV5HDGD"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Regular (ulid)",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/01EZPHFCD8000C438200A44F1M"),
					inputEmoteHash: "01EZPHFCD8000C438200A44F1M",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://example.com/chatterino/thumbnail/https%3A%2F%2Fcdn.7tv.app%2Femote%2F01EZPHFCD8000C438200A44F1M%2Fbest.webp","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EmonkaE%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EShared%207TV%20Emote%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EBy:%3C%2Fb%3E%20Zhark%0A%3C%2Fdiv%3E","link":"https://7tv.app/emotes/01EZPHFCD8000C438200A44F1M"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Regular,global (ulid)",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/01GB9W8JN80004CKF2H1TWA99H"),
					inputEmoteHash: "01GB9W8JN80004CKF2H1TWA99H",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://example.com/chatterino/thumbnail/https%3A%2F%2Fcdn.7tv.app%2Femote%2F01GB9W8JN80004CKF2H1TWA99H%2Fbest.webp","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EFeelsDankMan%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EShared%207TV%20Emote%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EBy:%3C%2Fb%3E%20clyverE%0A%3C%2Fdiv%3E","link":"https://7tv.app/emotes/01GB9W8JN80004CKF2H1TWA99H"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Regular, no images (ulid)",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/01F6MA6Y100002B6P5MWZ5D916"),
					inputEmoteHash: "01F6MA6Y100002B6P5MWZ5D916",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EHmm%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EShared%207TV%20Emote%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EBy:%3C%2Fb%3E%20lnsc%0A%3C%2Fdiv%3E","link":"https://7tv.app/emotes/01F6MA6Y100002B6P5MWZ5D916"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Unlisted, Private (ulid)",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/01F7GJ0N4R00074A83FVHRDMDB"),
					inputEmoteHash: "01F7GJ0N4R00074A83FVHRDMDB",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EOkayge%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EPrivate%207TV%20Emote%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EBy:%3C%2Fb%3E%20joonwi%0A%3Cli%3E%3Cb%3E%3Cspan%20style=%22color:%20red%3B%22%3EUNLISTED%3C%2Fspan%3E%3C%2Fb%3E%3C%2Fli%3E%0A%3C%2Fdiv%3E","link":"https://7tv.app/emotes/01F7GJ0N4R00074A83FVHRDMDB"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Matching link - 404 (ulid)",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/11F01WNXA00001NSRF006MFZYS"),
					inputEmoteHash: "11F01WNXA00001NSRF006MFZYS",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":404,"message":"No 7TV emote with this id found"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Private",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/604281c81ae70f000d47ffd9"),
					inputEmoteHash: "604281c81ae70f000d47ffd9",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://example.com/chatterino/thumbnail/https%3A%2F%2Fcdn.7tv.app%2Femote%2F604281c81ae70f000d47ffd9%2Fbest.webp","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EPajawalk%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EPrivate%207TV%20Emote%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EBy:%3C%2Fb%3E%20durado_%0A%3C%2Fdiv%3E","link":"https://7tv.app/emotes/604281c81ae70f000d47ffd9"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Unlisted",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/60ae8d9ff39a7552b658b60d"),
					inputEmoteHash: "60ae8d9ff39a7552b658b60d",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EBedge%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EShared%207TV%20Emote%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EBy:%3C%2Fb%3E%20Paruna%0A%3Cli%3E%3Cb%3E%3Cspan%20style=%22color:%20red%3B%22%3EUNLISTED%3C%2Fspan%3E%3C%2Fb%3E%3C%2Fli%3E%0A%3C%2Fdiv%3E","link":"https://7tv.app/emotes/60ae8d9ff39a7552b658b60d"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Regular",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/603cb219c20d020014423c34"),
					inputEmoteHash: "603cb219c20d020014423c34",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://example.com/chatterino/thumbnail/https%3A%2F%2Fcdn.7tv.app%2Femote%2F603cb219c20d020014423c34%2Fbest.webp","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EmonkaE%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EShared%207TV%20Emote%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EBy:%3C%2Fb%3E%20Zhark%0A%3C%2Fdiv%3E","link":"https://7tv.app/emotes/603cb219c20d020014423c34"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Regular,global",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/63071bb9464de28875c52531"),
					inputEmoteHash: "63071bb9464de28875c52531",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://example.com/chatterino/thumbnail/https%3A%2F%2Fcdn.7tv.app%2Femote%2F63071bb9464de28875c52531%2Fbest.webp","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EFeelsDankMan%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EShared%207TV%20Emote%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EBy:%3C%2Fb%3E%20clyverE%0A%3C%2Fdiv%3E","link":"https://7tv.app/emotes/63071bb9464de28875c52531"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Regular, no images",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/60ae3e54259ac5a73e56a426"),
					inputEmoteHash: "60ae3e54259ac5a73e56a426",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EHmm%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EShared%207TV%20Emote%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EBy:%3C%2Fb%3E%20lnsc%0A%3C%2Fdiv%3E","link":"https://7tv.app/emotes/60ae3e54259ac5a73e56a426"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Unlisted, Private",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/60bcb44f7229037ee386d1ab"),
					inputEmoteHash: "60bcb44f7229037ee386d1ab",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EOkayge%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EPrivate%207TV%20Emote%3C%2Fb%3E%3Cbr%3E%0A%3Cb%3EBy:%3C%2Fb%3E%20joonwi%0A%3Cli%3E%3Cb%3E%3Cspan%20style=%22color:%20red%3B%22%3EUNLISTED%3C%2Fspan%3E%3C%2Fb%3E%3C%2Fli%3E%0A%3C%2Fdiv%3E","link":"https://7tv.app/emotes/60bcb44f7229037ee386d1ab"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Matching link - 404",
					inputURL:       utils.MustParseURL("https://7tv.app/emotes/70bcb44f7229037ee386d1ab"),
					inputEmoteHash: "70bcb44f7229037ee386d1ab",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":404,"message":"No 7TV emote with this id found"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
			}

			const q = `SELECT value FROM cache WHERE key=$1`

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
					pool.ExpectExec("INSERT INTO cache").
						WithArgs("seventv:emote:"+test.inputEmoteHash, test.expectedResponse.Payload, test.expectedResponse.StatusCode, test.expectedResponse.ContentType, pgxmock.AnyArg()).
						WillReturnResult(pgxmock.NewResult("INSERT", 1))
					outputBytes, outputError := resolver.Run(ctx, test.inputURL, test.inputReq)
					c.Assert(outputError, qt.Equals, test.expectedError)
					c.Assert(outputBytes, qt.DeepEquals, test.expectedResponse)
					c.Assert(pool.ExpectationsWereMet(), qt.IsNil)
				})
			}
		})
	})
}
