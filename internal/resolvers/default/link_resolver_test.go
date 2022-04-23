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

func TestLinkResolver(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

	cfg := config.APIConfig{
		MaxContentLength: 1024,
	}
	pool, _ := pgxmock.NewPool()

	resolver.InitializeStaticResponses(ctx, cfg)

	router := chi.NewRouter()

	r := New(ctx, cfg, pool, nil)

	router.Get("/link_resolver/{url}", r.HandleRequest)
	router.Get("/thumbnail/{url}", r.HandleThumbnailRequest)

	var resolverResponses = map[string]string{}

	resolverResponses["/"] = "<html><head><title>/ title</title></head><body>xD</body></html>"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if response, ok := resolverResponses[r.URL.Path]; ok {
			w.Write([]byte(response))
			return
		}
		http.Error(w, http.StatusText(404), 404)
	}))
	defer ts.Close()

	c.Run("Request", func(c *qt.C) {
		tests := []struct {
			inputReq     *http.Request
			inputLinkKey string
			expected     resolver.Response
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
			},
		}

		for _, test := range tests {
			c.Run("", func(c *qt.C) {
				respRec := httptest.NewRecorder()

				pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
				pool.ExpectExec("INSERT INTO cache").
					WithArgs("default:link:"+test.inputLinkKey, pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				router.ServeHTTP(respRec, test.inputReq)
				resp := respRec.Result()
				response := resolver.Response{}
				err := json.NewDecoder(resp.Body).Decode(&response)
				c.Assert(err, qt.IsNil)

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

	c.Run("Request with early error", func(c *qt.C) {
		tests := []struct {
			inputReq     *http.Request
			inputLinkKey string
			expected     resolver.Response
		}{
			{
				inputReq:     newLinkResolverRequest(t, ctx, "GET", " :", nil),
				inputLinkKey: ts.URL,
				expected: resolver.Response{
					Status:  500,
					Link:    "",
					Message: `Could not fetch link info: Invalid URL`,
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
}
