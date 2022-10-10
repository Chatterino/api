package twitter

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
	qt "github.com/frankban/quicktest"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
)

func TestResolver(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

	datetime := time.Date(2019, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()
	fmt.Println(datetime)

	cfg := config.APIConfig{
		BaseURL:            "https://example.com/",
		TwitterBearerToken: "aaa",
	}

	pool, _ := pgxmock.NewPool()

	ts := testServer()
	defer ts.Close()

	r := NewTwitterResolver(ctx,
		cfg,
		pool,
		ts.URL+"/1.1/users/show.json?screen_name=%s",
		ts.URL+"/1.1/statuses/show.json?id=%s&tweet_mode=extended",
	)

	c.Assert(r, qt.IsNotNil)

	c.Run("Name", func(c *qt.C) {
		c.Assert(r.Name(), qt.Equals, "twitter")
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
				input:    utils.MustParseURL("https://twitter.com/pajlada"),
				expected: true,
			},
			{
				label:    "Matching domain, WWW",
				input:    utils.MustParseURL("https://www.twitter.com/pajlada"),
				expected: true,
			},
			{
				label:    "Matching domain, tweet",
				input:    utils.MustParseURL("https://twitter.com/pajlada/status/1507648130682077194"),
				expected: true,
			},
			{
				label:    "Matching domain, no WWW, http",
				input:    utils.MustParseURL("http://twitter.com/pajlada"),
				expected: true,
			},
			{
				label:    "Matching domain, WWW, http",
				input:    utils.MustParseURL("http://www.twitter.com/pajlada"),
				expected: true,
			},
			{
				label:    "Non-matching domain",
				input:    utils.MustParseURL("https://google.com"),
				expected: false,
			},
			{
				label:    "Non-matching domain",
				input:    utils.MustParseURL("https://nontwitter.com"),
				expected: false,
			},
			{
				label:    "Matching domain, no path",
				input:    utils.MustParseURL("https://twitter.com/"),
				expected: false,
			},
			{
				label:    "Matching domain, ignored path",
				input:    utils.MustParseURL("https://twitter.com/compose"),
				expected: false,
			},
			{
				label:    "Matching domain, ignored path",
				input:    utils.MustParseURL("https://twitter.com/logout"),
				expected: false,
			},
		}

		for _, test := range tests {
			c.Run(test.label, func(c *qt.C) {
				_, output := r.Check(ctx, test.input)
				c.Assert(output, qt.Equals, test.expected, qt.Commentf("URL was not handled as expected: %s", test.input))
			})
		}
	})

	c.Run("Run", func(c *qt.C) {
		c.Run("Dont handle", func(c *qt.C) {
			type runTest struct {
				label          string
				inputURL       *url.URL
				inputEmoteHash string
			}

			tests := []runTest{
				{
					label:    "missing params",
					inputURL: utils.MustParseURL("https://twitter.com"),
				},
			}

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					outputBytes, outputError := r.Run(ctx, test.inputURL, nil)
					c.Assert(outputError, qt.Equals, resolver.ErrDontHandle)
					c.Assert(outputBytes, qt.IsNil)
					c.Assert(pool.ExpectationsWereMet(), qt.IsNil)
				})
			}
		})

		c.Run("Tweet", func(c *qt.C) {
			type runTest struct {
				label            string
				inputURL         *url.URL
				inputTweet       string
				expectedResponse *cache.Response
				expectedError    error
				rowsReturned     int
			}

			tests := []runTest{
				{
					label:      "good",
					inputURL:   utils.MustParseURL("https://twitter.com/pajlada/status/1507648130682077194"),
					inputTweet: "1507648130682077194",
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EPAJLADA%20%28@pajlada%29%3C%2Fb%3E%0A%3Cspan%20style=%22white-space:%20pre-wrap%3B%20word-wrap:%20break-word%3B%22%3E%0ADigging%20a%20hole%0A%3C%2Fspan%3E%0A%3Cspan%20style=%22color:%20%23808892%3B%22%3E69%20likes\u0026nbsp%3B%E2%80%A2\u0026nbsp%3B420%20retweets\u0026nbsp%3B%E2%80%A2\u0026nbsp%3B26%20Mar%202022%20%E2%80%A2%2017:15%20UTC%3C%2Fspan%3E%0A%3C%2Fdiv%3E%0A"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:      "with entities",
					inputURL:   utils.MustParseURL("https://twitter.com/pajlada/status/1506968434134953986"),
					inputTweet: "1506968434134953986",
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://pbs.twimg.com/media/FOnTzeQWUAMU6L1?format=jpg\u0026name=medium","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EPAJLADA%20%28@pajlada%29%3C%2Fb%3E%0A%3Cspan%20style=%22white-space:%20pre-wrap%3B%20word-wrap:%20break-word%3B%22%3E%0A%0A%3C%2Fspan%3E%0A%3Cspan%20style=%22color:%20%23808892%3B%22%3E69%20likes\u0026nbsp%3B%E2%80%A2\u0026nbsp%3B420%20retweets\u0026nbsp%3B%E2%80%A2\u0026nbsp%3B26%20Mar%202022%20%E2%80%A2%2017:15%20UTC%3C%2Fspan%3E%0A%3C%2Fdiv%3E%0A"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:      "Poorly formatted timestamp",
					inputURL:   utils.MustParseURL("https://twitter.com/pajlada/status/1505121705290874881"),
					inputTweet: "1505121705290874881",
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EPAJLADA%20%28@pajlada%29%3C%2Fb%3E%0A%3Cspan%20style=%22white-space:%20pre-wrap%3B%20word-wrap:%20break-word%3B%22%3E%0ABad%20timestamp%0A%3C%2Fspan%3E%0A%3Cspan%20style=%22color:%20%23808892%3B%22%3E420%20likes\u0026nbsp%3B%E2%80%A2\u0026nbsp%3B69%20retweets\u0026nbsp%3B%E2%80%A2\u0026nbsp%3B%3C%2Fspan%3E%0A%3C%2Fdiv%3E%0A"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:      "404",
					inputURL:   utils.MustParseURL("https://twitter.com/pajlada/status/404"),
					inputTweet: "404",
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":404,"message":"Twitter tweet not found: 404"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:      "500",
					inputURL:   utils.MustParseURL("https://twitter.com/pajlada/status/500"),
					inputTweet: "500",
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":500,"message":"Twitter tweet API error: unhandled status code: 500"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:      "bad JSON",
					inputURL:   utils.MustParseURL("https://twitter.com/pajlada/status/bad"),
					inputTweet: "bad",
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":500,"message":"Twitter tweet API error: unable to unmarshal response"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
			}

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
					pool.ExpectExec("INSERT INTO cache").
						WithArgs("twitter:tweet:"+test.inputTweet, test.expectedResponse.Payload, test.expectedResponse.StatusCode, test.expectedResponse.ContentType, pgxmock.AnyArg()).
						WillReturnResult(pgxmock.NewResult("INSERT", 1))
					outputBytes, outputError := r.Run(ctx, test.inputURL, nil)
					c.Assert(outputError, qt.Equals, test.expectedError)
					c.Assert(outputBytes, qt.DeepEquals, test.expectedResponse)
					c.Assert(pool.ExpectationsWereMet(), qt.IsNil)
				})
			}
		})

		c.Run("User", func(c *qt.C) {
			type runTest struct {
				label            string
				inputURL         *url.URL
				inputUser        string
				expectedResponse *cache.Response
				expectedError    error
				rowsReturned     int
			}

			tests := []runTest{
				{
					label:     "User: pajlada",
					inputURL:  utils.MustParseURL("https://twitter.com/pajlada"),
					inputUser: "pajlada",
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://pbs.twimg.com/profile_images/1385924241619628033/fW7givJA_400x400.jpg","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EPAJLADA%20%28@pajlada%29%3C%2Fb%3E%0A%3Cspan%20style=%22white-space:%20pre-wrap%3B%20word-wrap:%20break-word%3B%22%3E%0ACool%20memer%0A%3C%2Fspan%3E%0A%3Cspan%20style=%22color:%20%23808892%3B%22%3E69%20followers%3C%2Fspan%3E%0A%3C%2Fdiv%3E%0A"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:     "User: pajlada (uppercase)",
					inputURL:  utils.MustParseURL("https://twitter.com/PAJLADA"),
					inputUser: "pajlada",
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://pbs.twimg.com/profile_images/1385924241619628033/fW7givJA_400x400.jpg","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EPAJLADA%20%28@pajlada%29%3C%2Fb%3E%0A%3Cspan%20style=%22white-space:%20pre-wrap%3B%20word-wrap:%20break-word%3B%22%3E%0ACool%20memer%0A%3C%2Fspan%3E%0A%3Cspan%20style=%22color:%20%23808892%3B%22%3E69%20followers%3C%2Fspan%3E%0A%3C%2Fdiv%3E%0A"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:     "User: 404",
					inputURL:  utils.MustParseURL("https://twitter.com/404"),
					inputUser: "404",
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":404,"message":"Twitter user not found: 404"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:     "User: 500",
					inputURL:  utils.MustParseURL("https://twitter.com/500"),
					inputUser: "500",
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":500,"message":"Twitter user API error: unhandled status code: 500"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:     "User: bad json",
					inputURL:  utils.MustParseURL("https://twitter.com/bad"),
					inputUser: "bad",
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":500,"message":"Twitter user API error: unable to unmarshal response"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
			}

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
					pool.ExpectExec("INSERT INTO cache").
						WithArgs("twitter:user:"+test.inputUser, test.expectedResponse.Payload, test.expectedResponse.StatusCode, test.expectedResponse.ContentType, pgxmock.AnyArg()).
						WillReturnResult(pgxmock.NewResult("INSERT", 1))
					outputBytes, outputError := r.Run(ctx, test.inputURL, nil)
					c.Assert(outputError, qt.Equals, test.expectedError)
					c.Assert(outputBytes, qt.DeepEquals, test.expectedResponse)
					c.Assert(pool.ExpectationsWereMet(), qt.IsNil)
				})
			}
		})
	})
}
