package youtube

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/utils"
	qt "github.com/frankban/quicktest"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
	"google.golang.org/api/option"
	youtubeAPI "google.golang.org/api/youtube/v3"
)

func TestChannelResolver(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := qt.New(t)

	pool, _ := pgxmock.NewPool()
	cfg := config.APIConfig{}

	r := chi.NewRouter()
	r.Get("/youtube/v3/search", func(w http.ResponseWriter, r *http.Request) {
		channelID := r.URL.Query().Get("q")
		w.Header().Set("Content-Type", "application/json")
		if channelID == "badresp" {
			w.Write([]byte("xd"))
			return
		}
		if resp, ok := channelSearches[channelID]; ok {
			b, err := json.Marshal(resp)
			if err != nil {
				http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
				return
			}
			w.Write(b)

			return
		}

		http.Error(w, http.StatusText(404), 404)
	})
	r.Get("/youtube/v3/channels", func(w http.ResponseWriter, r *http.Request) {
		channelID := r.URL.Query().Get("id")
		channelUsername := r.URL.Query().Get("forUsername")
		w.Header().Set("Content-Type", "application/json")
		if channelID == "badresp" {
			w.Write([]byte("xd"))
			return
		}

		if channelUsername != "" {
			if resp, ok := channels["user:"+channelUsername]; ok {
				b, err := json.Marshal(resp)
				if err != nil {
					http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
					return
				}
				w.Write(b)

				return
			}
		} else if resp, ok := channels[channelID]; ok {
			b, err := json.Marshal(resp)
			if err != nil {
				http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
				return
			}
			w.Write(b)

			return
		}

		http.Error(w, http.StatusText(404), 404)
	})
	ts := httptest.NewServer(r)
	defer ts.Close()
	youtubeClient, err := youtubeAPI.NewService(ctx, option.WithAPIKey("test"), option.WithEndpoint(ts.URL))
	c.Assert(err, qt.IsNil)

	resolver := NewYouTubeChannelResolver(ctx, cfg, pool, youtubeClient)

	c.Assert(resolver, qt.IsNotNil)

	c.Run("Name", func(c *qt.C) {
		c.Assert(resolver.Name(), qt.Equals, "youtube:channel")
	})

	c.Run("Check", func(c *qt.C) {
		type checkTest struct {
			label    string
			input    *url.URL
			expected bool
		}

		tests := []checkTest{
			{
				label:    "Correct domain, bad (video) path",
				input:    utils.MustParseURL("https://youtube.com/watch?v=foobar"),
				expected: false,
			},
			{
				label:    "Correct domain, bad (embed) path",
				input:    utils.MustParseURL("https://youtube.com/embed/foobar"),
				expected: false,
			},
			{
				label:    "Correct domain, bad (video) path",
				input:    utils.MustParseURL("https://www.youtube.com/watch?v=foobar"),
				expected: false,
			},
			{
				label:    "Correct domain, bad (embed) path",
				input:    utils.MustParseURL("https://www.youtube.com/embed/foobar"),
				expected: false,
			},
			{
				label:    "Correct domain, correct path",
				input:    utils.MustParseURL("https://youtube.com/user/aragusea"),
				expected: true,
			},
			{
				label:    "Correct (sub)domain, correct path",
				input:    utils.MustParseURL("https://www.youtube.com/user/aragusea"),
				expected: true,
			},
			{
				label:    "Correct domain, no path",
				input:    utils.MustParseURL("https://youtube.com"),
				expected: false,
			},
			{
				label:    "Correct domain, results path",
				input:    utils.MustParseURL("https://youtube.com/results?search_query=test"),
				expected: false,
			},
			{
				label:    "Incorrect domain",
				input:    utils.MustParseURL("https://example.com/watch?v=foobar"),
				expected: false,
			},
		}

		for _, test := range tests {
			c.Run(test.label, func(c *qt.C) {
				_, output := resolver.Check(ctx, test.input)
				c.Assert(output, qt.Equals, test.expected, qt.Commentf("%s must %v", test.input, test.expected))
			})
		}
	})

	c.Run("Run", func(c *qt.C) {
		c.Run("Error", func(c *qt.C) {
			type runTest struct {
				label            string
				inputURL         *url.URL
				inputEmoteHash   string
				inputReq         *http.Request
				expectedResponse *cache.Response
				expectedError    error
			}

			tests := []runTest{
				{
					label:    "Non-matching link",
					inputURL: utils.MustParseURL("https://clips.twitch.tv/"),
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":404,"message":"Could not fetch link info: No link info found"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
			}

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					outputResponse, outputError := resolver.Run(ctx, test.inputURL, test.inputReq)
					c.Assert(outputError, qt.Equals, test.expectedError)
					c.Assert(outputResponse, qt.DeepEquals, test.expectedResponse)
				})
			}
		})

		c.Run("Not cached", func(c *qt.C) {
			type runTest struct {
				label            string
				inputURL         *url.URL
				inputVideoID     string
				inputReq         *http.Request
				expectedResponse *cache.Response
				expectedError    error
				rowsReturned     int
			}

			tests := []runTest{
				// {
				// 	label:         "Video",
				// 	inputURL:      utils.MustParseURL("https://youtube.com/watch?v=foobar"),
				// 	inputVideoID:  "foobar",
				// 	inputReq:      nil,
				// 	expectedBytes: []byte(`{"status":200,"thumbnail":"https://example.com/thumbnail.png","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EVideo%20Title%3C%2Fb%3E%0A%3Cbr%3E%3Cb%3EChannel:%3C%2Fb%3E%20Channel%20Title%0A%3Cbr%3E%3Cb%3EDuration:%3C%2Fb%3E%2000:00:00%0A%3Cbr%3E%3Cb%3EPublished:%3C%2Fb%3E%2012%20Oct%202019%0A%3Cbr%3E%3Cb%3EViews:%3C%2Fb%3E%2050%0A%3Cbr%3E%3Cb%3E%3Cspan%20style=%22color:%20red%3B%22%3EAGE%20RESTRICTED%3C%2Fspan%3E%3C%2Fb%3E%0A%3Cbr%3E%3Cspan%20style=%22color:%20%232ecc71%3B%22%3E10%20likes%3C%2Fspan%3E\u0026nbsp%3B%E2%80%A2\u0026nbsp%3B%3Cspan%20style=%22color:%20%23808892%3B%22%3E5%20comments%3C%2Fspan%3E%0A%3C%2Fdiv%3E%0A"}`),
				// },
				{
					label:        "Channel:404",
					inputURL:     utils.MustParseURL("https://youtube.com/channel/404"),
					inputVideoID: "channel:404",
					inputReq:     nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":404,"message":"No YouTube channel with the ID 404 found"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
				},
				{
					label:        "Channel:Too many videos",
					inputURL:     utils.MustParseURL("https://youtube.com/channel/toomany"),
					inputVideoID: "channel:toomany",
					inputReq:     nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":500,"message":"YouTube channel response contained 2 items"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
				},
				{
					label:        "C:404",
					inputURL:     utils.MustParseURL("https://youtube.com/c/404"),
					inputVideoID: "c:404",
					inputReq:     nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":404,"message":"No YouTube channel with the ID 404 found"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
				},
				{
					label:        "C:Too many videos",
					inputURL:     utils.MustParseURL("https://youtube.com/c/toomany"),
					inputVideoID: "c:toomany",
					inputReq:     nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":500,"message":"YouTube search response contained 2 items"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
				},
				{
					label:        "C:Custom",
					inputURL:     utils.MustParseURL("https://youtube.com/c/custom"),
					inputVideoID: "c:custom",
					inputReq:     nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://example.com/thumbnail.png","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3ECool%20YouTube%20Channel%3C%2Fb%3E%0A%3Cbr%3E%3Cb%3EJoined%20Date:%3C%2Fb%3E%2012%20Oct%202019%0A%3Cbr%3E%3Cb%3ESubscribers:%3C%2Fb%3E%2069%0A%3Cbr%3E%3Cb%3EViews:%3C%2Fb%3E%20420%0A%3C%2Fdiv%3E%0A"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
				},
				{
					label:        "User:zneix",
					inputURL:     utils.MustParseURL("https://youtube.com/user/zneix"),
					inputVideoID: "user:zneix",
					inputReq:     nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://example.com/thumbnail.png","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3ECool%20YouTube%20Channel%3C%2Fb%3E%0A%3Cbr%3E%3Cb%3EJoined%20Date:%3C%2Fb%3E%2012%20Oct%202019%0A%3Cbr%3E%3Cb%3ESubscribers:%3C%2Fb%3E%2069%0A%3Cbr%3E%3Cb%3EViews:%3C%2Fb%3E%20420%0A%3C%2Fdiv%3E%0A"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
				},
				// {
				// 	label:         "Unavailable",
				// 	inputURL:      utils.MustParseURL("https://youtube.com/watch?v=unavailable"),
				// 	inputVideoID:  "unavailable",
				// 	inputReq:      nil,
				// 	expectedBytes: []byte(`{"status":500,"message":"YouTube video unavailable"}`),
				// },
				{
					label:        "Channel:Medium thumbnail",
					inputURL:     utils.MustParseURL("https://youtube.com/channel/mediumtn"),
					inputVideoID: "channel:mediumtn",
					inputReq:     nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://example.com/medium.png","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3ECool%20YouTube%20Channel%3C%2Fb%3E%0A%3Cbr%3E%3Cb%3EJoined%20Date:%3C%2Fb%3E%2012%20Oct%202019%0A%3Cbr%3E%3Cb%3ESubscribers:%3C%2Fb%3E%2069%0A%3Cbr%3E%3Cb%3EViews:%3C%2Fb%3E%20420%0A%3C%2Fdiv%3E%0A"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
				},
				{
					label:        "C:Bad response",
					inputURL:     utils.MustParseURL("https://youtube.com/c/badresp"),
					inputVideoID: "c:badresp",
					inputReq:     nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":500,"message":"YouTube search API error: invalid character \u0026#39;x\u0026#39; looking for beginning of value"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
				},
				{
					label:        "Channel:Bad response",
					inputURL:     utils.MustParseURL("https://youtube.com/channel/badresp"),
					inputVideoID: "channel:badresp",
					inputReq:     nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":500,"message":"YouTube API error: invalid character \u0026#39;x\u0026#39; looking for beginning of value"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
				},
			}

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
					pool.ExpectExec("INSERT INTO cache").
						WithArgs("youtube:channel:"+test.inputVideoID, test.expectedResponse.Payload, test.expectedResponse.StatusCode, test.expectedResponse.ContentType, pgxmock.AnyArg()).
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
