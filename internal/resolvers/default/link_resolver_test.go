package defaultresolver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	qt "github.com/frankban/quicktest"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
)

func newRequest(t *testing.T, ctx context.Context, method string, url string, payload io.Reader) *http.Request {
	req, err := http.NewRequestWithContext(ctx, method, url, payload)
	if err != nil {
		t.Fatal("Unable to create request")
	}

	return req
}

func newLinkResolverRequest(t *testing.T, ctx context.Context, method string, u string, payload io.Reader) *http.Request {
	finalURL := "/link_resolver/" + url.QueryEscape(u)
	fmt.Println("Final URL", finalURL)
	req, err := http.NewRequestWithContext(ctx, method, finalURL, payload)
	if err != nil {
		t.Fatal("Unable to create request")
	}

	return req
}

func newThumbnailRequest(t *testing.T, ctx context.Context, method string, u string, payload io.Reader) *http.Request {
	finalURL := "/thumbnail/" + url.QueryEscape(u)
	fmt.Println("Thumbnail request to", finalURL)
	req, err := http.NewRequestWithContext(ctx, method, finalURL, payload)
	if err != nil {
		t.Fatal("Unable to create request")
	}

	return req
}

func TestLinkResolver(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

	cfg := config.APIConfig{
		MaxContentLength: 5 * 1024 * 1024, // 5 MB
	}
	pool, _ := pgxmock.NewPool()

	resolver.InitializeStaticResponses(ctx, cfg)

	router := chi.NewRouter()

	ignoredHosts := map[string]struct{}{
		"ignoredhost.com": {},
	}

	r := New(ctx, cfg, pool, nil, ignoredHosts)

	router.Get("/link_resolver/{url}", r.HandleRequest)
	router.Get("/thumbnail/{url}", r.HandleThumbnailRequest)

	resolverResponses := map[string]string{}

	resolverResponses["/"] = "<html><head><title>/ title</title></head><body>xD</body></html>"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if response, ok := resolverResponses[r.URL.Path]; ok {
			w.Write([]byte(response))
			return
		}

		fmt.Println("404 xd")
		http.Error(w, http.StatusText(404), 404)
	}))
	defer ts.Close()

	thumbnailResponses := map[string][]byte{}

	thumbnailResponses["/thumb1.png"] = []byte{'\x89', '\x50', '\x4e', '\x47', '\x0d', '\x0a', '\x1a', '\x0a', '\x00', '\x00', '\x00', '\x0d', '\x49', '\x48', '\x44', '\x52', '\x00', '\x00', '\x01', '\x00', '\x00', '\x00', '\x01', '\x00', '\x01', '\x03', '\x00', '\x00', '\x00', '\x66', '\xbc', '\x3a', '\x25', '\x00', '\x00', '\x00', '\x03', '\x50', '\x4c', '\x54', '\x45', '\xb5', '\xd0', '\xd0', '\x63', '\x04', '\x16', '\xea', '\x00', '\x00', '\x00', '\x1f', '\x49', '\x44', '\x41', '\x54', '\x68', '\x81', '\xed', '\xc1', '\x01', '\x0d', '\x00', '\x00', '\x00', '\xc2', '\xa0', '\xf7', '\x4f', '\x6d', '\x0e', '\x37', '\xa0', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\xbe', '\x0d', '\x21', '\x00', '\x00', '\x01', '\x9a', '\x60', '\xe1', '\xd5', '\x00', '\x00', '\x00', '\x00', '\x49', '\x45', '\x4e', '\x44', '\xae', '\x42', '\x60', '\x82'}
	thumbnailResponses["/toobig.png"] = []byte{'\x89', '\x50', '\x4e', '\x47', '\x0d', '\x0a', '\x1a', '\x0a', '\x00', '\x00', '\x00', '\x0d', '\x49', '\x48', '\x44', '\x52', '\x00', '\x00', '\x01', '\x00', '\x00', '\x00', '\x01', '\x00', '\x01', '\x03', '\x00', '\x00', '\x00', '\x66', '\xbc', '\x3a', '\x25', '\x00', '\x00', '\x00', '\x03', '\x50', '\x4c', '\x54', '\x45', '\xb5', '\xd0', '\xd0', '\x63', '\x04', '\x16', '\xea', '\x00', '\x00', '\x00', '\x1f', '\x49', '\x44', '\x41', '\x54', '\x68', '\x81', '\xed', '\xc1', '\x01', '\x0d', '\x00', '\x00', '\x00', '\xc2', '\xa0', '\xf7', '\x4f', '\x6d', '\x0e', '\x37', '\xa0', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\xbe', '\x0d', '\x21', '\x00', '\x00', '\x01', '\x9a', '\x60', '\xe1', '\xd5', '\x00', '\x00', '\x00', '\x00', '\x49', '\x45', '\x4e', '\x44', '\xae', '\x42', '\x60', '\x82'}
	thumbnailResponses["/unsupported-thumbnail-format.foo"] = []byte{'\x89', '\x50', '\x4e', '\x47', '\x0d', '\x0a', '\x1a', '\x0a', '\x00', '\x00', '\x00', '\x0d', '\x49', '\x48', '\x44', '\x52', '\x00', '\x00', '\x01', '\x00', '\x00', '\x00', '\x01', '\x00', '\x01', '\x03', '\x00', '\x00', '\x00', '\x66', '\xbc', '\x3a', '\x25', '\x00', '\x00', '\x00', '\x03', '\x50', '\x4c', '\x54', '\x45', '\xb5', '\xd0', '\xd0', '\x63', '\x04', '\x16', '\xea', '\x00', '\x00', '\x00', '\x1f', '\x49', '\x44', '\x41', '\x54', '\x68', '\x81', '\xed', '\xc1', '\x01', '\x0d', '\x00', '\x00', '\x00', '\xc2', '\xa0', '\xf7', '\x4f', '\x6d', '\x0e', '\x37', '\xa0', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\xbe', '\x0d', '\x21', '\x00', '\x00', '\x01', '\x9a', '\x60', '\xe1', '\xd5', '\x00', '\x00', '\x00', '\x00', '\x49', '\x45', '\x4e', '\x44', '\xae', '\x42', '\x60', '\x82'}

	thumbnailTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/toobig.png":
			w.Header().Add("Content-Length", "999999999999999999")

		case "/unsupported-thumbnail-format.foo":
			// Use some dummy Content-Type. If we used a real Content-Type that we later add support
			// for, these tests would break
			w.Header().Set("Content-Type", "application/foo")
		}

		if response, ok := thumbnailResponses[r.URL.Path]; ok {
			w.Write(response)
			return
		}

		http.Error(w, http.StatusText(404), 404)
	}))
	defer thumbnailTs.Close()

	c.Run("Request", func(c *qt.C) {
		tests := []struct {
			inputReq            *http.Request
			inputLinkKey        string
			expected            resolver.Response
			expectedStatusCode  int
			expectedContentType string
		}{
			{
				inputReq:     newLinkResolverRequest(t, ctx, "GET", ts.URL, nil),
				inputLinkKey: ts.URL,
				expected: resolver.Response{
					Status: 200,
					Link:   ts.URL,
					Tooltip: `<div style="text-align: left;">

<b>/ title</b><hr>


<b>URL:</b> http://127\.0\.0\.1:[\d]{2,7}</div>`,
				},
				expectedStatusCode:  http.StatusOK,
				expectedContentType: "application/json",
			},
			{
				inputReq:     newLinkResolverRequest(t, ctx, "GET", ts.URL+"/404", nil),
				inputLinkKey: ts.URL + "/404",
				expected: resolver.Response{
					Status:  http.StatusNotFound,
					Message: `Could not fetch link info: No link info found`,
				},
				expectedStatusCode:  http.StatusOK,
				expectedContentType: "application/json",
			},
		}

		for _, test := range tests {
			c.Run(test.inputLinkKey, func(c *qt.C) {
				respRec := httptest.NewRecorder()

				pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
				pool.ExpectExec("INSERT INTO cache").
					WithArgs("default:link:"+test.inputLinkKey, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				router.ServeHTTP(respRec, test.inputReq)
				resp := respRec.Result()
				response := resolver.Response{}
				bod, _ := io.ReadAll(resp.Body)
				fmt.Println(string(bod))
				// err := json.NewDecoder(resp.Body).Decode(&response)
				err := json.Unmarshal([]byte(bod), &response)
				c.Assert(err, qt.IsNil)

				c.Assert(response.Status, qt.Equals, test.expected.Status)
				c.Assert(response.Link, qt.Equals, test.expected.Link)

				c.Assert(resp.Header.Get("Content-Type"), qt.Equals, test.expectedContentType)
				c.Assert(resp.StatusCode, qt.Equals, test.expectedStatusCode)

				unescapedTooltip, err := url.QueryUnescape(response.Tooltip)
				c.Assert(err, qt.IsNil)

				if test.expected.Tooltip != "" {
					c.Assert(unescapedTooltip, MatchesRegexp, regexp.MustCompile(test.expected.Tooltip), qt.Commentf("%s does not match %s", unescapedTooltip, test.expected.Tooltip))
				}
				if test.expected.Message != "" {
					c.Assert(response.Message, qt.Matches, test.expected.Message, qt.Commentf("%s does not match %s", response.Message, test.expected.Message))
				}

				c.Assert(pool.ExpectationsWereMet(), qt.IsNil)
			})
		}
	})

	c.Run("Early error", func(c *qt.C) {
		tests := []struct {
			inputReq     *http.Request
			inputLinkKey string
			expected     resolver.Response
		}{
			{
				inputReq:     newLinkResolverRequest(t, ctx, "GET", " :", nil),
				inputLinkKey: ts.URL,
				expected: resolver.Response{
					Status:  http.StatusBadRequest,
					Link:    "",
					Message: `Could not fetch link info: Invalid URL`,
				},
			},
			{
				inputReq:     newLinkResolverRequest(t, ctx, "GET", "https://ignoredhost.com/forsen", nil),
				inputLinkKey: ts.URL,
				expected: resolver.Response{
					Status:  http.StatusForbidden,
					Link:    "",
					Message: `Link forbidden`,
				},
			},
			{
				inputReq:     newLinkResolverRequest(t, ctx, "GET", "https://IgnoredHost.com/forsen", nil),
				inputLinkKey: ts.URL,
				expected: resolver.Response{
					Status:  http.StatusForbidden,
					Link:    "",
					Message: `Link forbidden`,
				},
			},
		}

		for _, test := range tests {
			c.Run("", func(c *qt.C) {
				respRec := httptest.NewRecorder()

				router.ServeHTTP(respRec, test.inputReq)
				resp := respRec.Result()
				response := resolver.Response{}
				err := json.NewDecoder(resp.Body).Decode(&response)
				c.Assert(err, qt.IsNil)

				c.Assert(resp.StatusCode, qt.Equals, test.expected.Status)
				c.Assert(response.Status, qt.Equals, test.expected.Status)
				c.Assert(response.Link, qt.Equals, test.expected.Link)

				unescapedTooltip, err := url.QueryUnescape(response.Tooltip)
				c.Assert(err, qt.IsNil)

				if test.expected.Tooltip != "" {
					c.Assert(unescapedTooltip, MatchesRegexp, regexp.MustCompile(test.expected.Tooltip), qt.Commentf("%s does not match %s", unescapedTooltip, test.expected.Tooltip))
				}
				if test.expected.Message != "" {
					c.Assert(response.Message, qt.Matches, test.expected.Message, qt.Commentf("%s does not match %s", response.Message, test.expected.Message))
				}

				c.Assert(pool.ExpectationsWereMet(), qt.IsNil)
			})
		}
	})

	c.Run("Thumbnail", func(c *qt.C) {
		tests := []struct {
			inputReq            *http.Request
			inputLinkKey        string
			expected            resolver.Response
			expectedContentType string
			expectedStatusCode  int
		}{
			{
				inputReq:     newThumbnailRequest(t, ctx, "GET", thumbnailTs.URL+"/thumb404.png", nil),
				inputLinkKey: thumbnailTs.URL + "/thumb404.png",
				expected: resolver.Response{
					Status:  404,
					Message: `Could not fetch thumbnail`,
				},
				expectedContentType: "application/json",
				expectedStatusCode:  http.StatusNotFound,
			},
			{
				inputReq:     newThumbnailRequest(t, ctx, "GET", thumbnailTs.URL+"/toobig.png", nil),
				inputLinkKey: thumbnailTs.URL + "/toobig.png",
				expected: resolver.Response{
					Status:  500,
					Message: `Could not fetch link info: Response too large (>5MB)`,
				},
				expectedContentType: "application/json",
				expectedStatusCode:  http.StatusInternalServerError,
			},
			{
				inputReq:     newThumbnailRequest(t, ctx, "GET", thumbnailTs.URL+"/unsupported-thumbnail-format.foo", nil),
				inputLinkKey: thumbnailTs.URL + "/unsupported-thumbnail-format.foo",
				expected: resolver.Response{
					Status:  415,
					Message: `Unsupported thumbnail type`,
				},
				expectedContentType: "application/json",
				expectedStatusCode:  http.StatusOK,
			},
		}

		for _, test := range tests {
			c.Run("", func(c *qt.C) {
				respRec := httptest.NewRecorder()

				pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
				pool.ExpectExec("INSERT INTO cache").
					WithArgs("default:thumbnail:"+test.inputLinkKey, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				router.ServeHTTP(respRec, test.inputReq)
				resp := respRec.Result()
				response := resolver.Response{}
				// bod, _ := io.ReadAll(resp.Body)
				// fmt.Println(string(bod))
				err := json.NewDecoder(resp.Body).Decode(&response)
				c.Assert(err, qt.IsNil)

				c.Assert(resp.StatusCode, qt.Equals, test.expectedStatusCode)
				c.Assert(resp.Header.Get("Content-Type"), qt.Equals, test.expectedContentType)
				c.Assert(response.Status, qt.Equals, test.expected.Status)
				c.Assert(response.Link, qt.Equals, test.expected.Link)
				c.Assert(response.Message, qt.Equals, test.expected.Message, qt.Commentf("%s does not match %s", response.Message, test.expected.Message))

				// unescapedTooltip, err := url.QueryUnescape(response.Tooltip)
				// c.Assert(err, qt.IsNil)

				c.Assert(pool.ExpectationsWereMet(), qt.IsNil)
			})
		}
	})
}
