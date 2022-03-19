package youtube

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
	qt "github.com/frankban/quicktest"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
	"google.golang.org/api/option"
	youtubeAPI "google.golang.org/api/youtube/v3"
)

func TestVideoShortURLResolver(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := qt.New(t)

	pool, _ := pgxmock.NewPool()
	cfg := config.APIConfig{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		videoID := r.URL.Query().Get("id")
		w.Header().Set("Content-Type", "application/json")
		if videoID == "badresp" {
			w.Write([]byte("xd"))
			return
		}
		if resp, ok := videos[videoID]; ok {
			b, err := json.Marshal(resp)
			if err != nil {
				http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
				return
			}
			w.Write(b)

			return
		}

		http.Error(w, http.StatusText(404), 404)
	}))
	defer ts.Close()
	youtubeClient, err := youtubeAPI.NewService(ctx, option.WithAPIKey("test"), option.WithEndpoint(ts.URL))
	c.Assert(err, qt.IsNil)

	loader := NewVideoLoader(youtubeClient)
	videoCache := cache.NewPostgreSQLCache(ctx, cfg, pool, "youtube:video", resolver.NewResponseMarshaller(loader), 24*time.Hour)

	resolver := NewYouTubeVideoShortURLResolver(videoCache)

	c.Assert(resolver, qt.IsNotNil)

	c.Run("Name", func(c *qt.C) {
		c.Assert(resolver.Name(), qt.Equals, "youtube:video:shorturl")
	})

	c.Run("Check", func(c *qt.C) {
		type checkTest struct {
			label    string
			input    *url.URL
			expected bool
		}

		tests := []checkTest{
			{
				label:    "Correct domain, correct path",
				input:    utils.MustParseURL("https://youtu.be/foobar"),
				expected: true,
			},
			{
				label:    "Correct domain, no path",
				input:    utils.MustParseURL("https://youtube.com"),
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
					inputURL:      utils.MustParseURL("https://clips.twitch.tv"),
					expectedError: errInvalidVideoLink,
				},
			}

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					outputBytes, outputError := resolver.Run(ctx, test.inputURL, test.inputReq)
					c.Assert(outputError, qt.Equals, test.expectedError)
					c.Assert(outputBytes, qt.IsNil)
				})
			}
		})

		c.Run("Not cached", func(c *qt.C) {
			type runTest struct {
				label         string
				inputURL      *url.URL
				inputVideoID  string
				inputReq      *http.Request
				expectedBytes []byte
				expectedError error
				rowsReturned  int
			}

			tests := []runTest{
				{
					label:         "Video",
					inputURL:      utils.MustParseURL("https://youtu.be/foobar"),
					inputVideoID:  "foobar",
					inputReq:      nil,
					expectedBytes: []byte(`{"status":200,"thumbnail":"https://example.com/thumbnail.png","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EVideo%20Title%3C%2Fb%3E%0A%3Cbr%3E%3Cb%3EChannel:%3C%2Fb%3E%20Channel%20Title%0A%3Cbr%3E%3Cb%3EDuration:%3C%2Fb%3E%2000:00:00%0A%3Cbr%3E%3Cb%3EPublished:%3C%2Fb%3E%2012%20Oct%202019%0A%3Cbr%3E%3Cb%3EViews:%3C%2Fb%3E%2050%0A%3Cbr%3E%3Cb%3E%3Cspan%20style=%22color:%20red%3B%22%3EAGE%20RESTRICTED%3C%2Fspan%3E%3C%2Fb%3E%0A%3Cbr%3E%3Cspan%20style=%22color:%20%232ecc71%3B%22%3E10%20likes%3C%2Fspan%3E\u0026nbsp%3B%E2%80%A2\u0026nbsp%3B%3Cspan%20style=%22color:%20%23808892%3B%22%3E5%20comments%3C%2Fspan%3E%0A%3C%2Fdiv%3E%0A"}`),
				},
				{
					label:         "404",
					inputURL:      utils.MustParseURL("https://youtu.be/404"),
					inputVideoID:  "404",
					inputReq:      nil,
					expectedBytes: []byte(`{"status":404,"message":"No YouTube video with the ID 404 found"}`),
				},
				{
					label:         "Too many videos",
					inputURL:      utils.MustParseURL("https://youtu.be/toomany"),
					inputVideoID:  "toomany",
					inputReq:      nil,
					expectedBytes: []byte(`{"status":500,"message":"YouTube API returned more than 2 videos"}`),
				},
				{
					label:         "Unavailable",
					inputURL:      utils.MustParseURL("https://youtu.be/unavailable"),
					inputVideoID:  "unavailable",
					inputReq:      nil,
					expectedBytes: []byte(`{"status":500,"message":"YouTube video unavailable"}`),
				},
				{
					label:         "Medium thumbnail",
					inputURL:      utils.MustParseURL("https://youtu.be/mediumtn"),
					inputVideoID:  "mediumtn",
					inputReq:      nil,
					expectedBytes: []byte(`{"status":200,"thumbnail":"https://example.com/medium.png","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EVideo%20Title%3C%2Fb%3E%0A%3Cbr%3E%3Cb%3EChannel:%3C%2Fb%3E%20Channel%20Title%0A%3Cbr%3E%3Cb%3EDuration:%3C%2Fb%3E%2000:00:00%0A%3Cbr%3E%3Cb%3EPublished:%3C%2Fb%3E%2012%20Oct%202019%0A%3Cbr%3E%3Cb%3EViews:%3C%2Fb%3E%2050%0A%3Cbr%3E%3Cb%3E%3Cspan%20style=%22color:%20red%3B%22%3EAGE%20RESTRICTED%3C%2Fspan%3E%3C%2Fb%3E%0A%3Cbr%3E%3Cspan%20style=%22color:%20%232ecc71%3B%22%3E10%20likes%3C%2Fspan%3E\u0026nbsp%3B%E2%80%A2\u0026nbsp%3B%3Cspan%20style=%22color:%20%23808892%3B%22%3E5%20comments%3C%2Fspan%3E%0A%3C%2Fdiv%3E%0A"}`),
				},
				{
					label:         "Bad response",
					inputURL:      utils.MustParseURL("https://youtu.be/badresp"),
					inputVideoID:  "badresp",
					inputReq:      nil,
					expectedBytes: []byte(`{"status":500,"message":"YouTube API error: invalid character \u0026#39;x\u0026#39; looking for beginning of value"}`),
				},
			}

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
					pool.ExpectExec("INSERT INTO cache").
						WithArgs("youtube:video:"+test.inputVideoID, test.expectedBytes, pgxmock.AnyArg()).
						WillReturnResult(pgxmock.NewResult("INSERT", 1))
					outputBytes, outputError := resolver.Run(ctx, test.inputURL, test.inputReq)
					c.Assert(outputError, qt.Equals, test.expectedError)
					c.Assert(outputBytes, qt.DeepEquals, test.expectedBytes)
				})
			}
		})
	})
}
