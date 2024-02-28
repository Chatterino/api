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
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
	"google.golang.org/api/option"
	youtubeAPI "google.golang.org/api/youtube/v3"
)

func TestPlaylistResolver(t *testing.T) {
	c := qt.New(t)

	ctx := logger.OnContext(context.Background(), logger.NewTest())
	cfg := config.APIConfig{}
	pool, _ := pgxmock.NewPool()

	r := chi.NewRouter()
	r.Get("/youtube/v3/playlists", func(w http.ResponseWriter, r *http.Request) {
		playlistID := r.URL.Query().Get("id")
		w.Header().Set("Content-Type", "application/json")
		if playlistID == "badresp" {
			w.Write([]byte("xd"))
			return
		}

		if resp, ok := playlists[playlistID]; ok {
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

	resolver := NewYouTubePlaylistResolver(ctx, cfg, pool, youtubeClient)

	c.Run("Check", func(c *qt.C) {
		type checkTest struct {
			label    string
			input    *url.URL
			expected bool
		}

		tests := []checkTest{
			{
				label:    "Correct domain, bad (playlist) path",
				input:    utils.MustParseURL("https://youtube.com/watch?v=foobar"),
				expected: false,
			},
			{
				label:    "Correct domain, bad (playlist) parameter",
				input:    utils.MustParseURL("https://youtube.com/playlist?v=foobar"),
				expected: false,
			},
			{
				label:    "Correct domain, no path",
				input:    utils.MustParseURL("https://youtube.com"),
				expected: false,
			},
			{
				label:    "Correct domain, correct path",
				input:    utils.MustParseURL("https://youtube.com/playlist?list=testing"),
				expected: true,
			},
			{
				label:    "Correct (sub)domain, correct path",
				input:    utils.MustParseURL("https://www.youtube.com/playlist?list=testing"),
				expected: true,
			},
			{
				label:    "Incorrect domain",
				input:    utils.MustParseURL("https://example.com/playlist?list=foobar"),
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
				inputPlaylistID  string
				inputReq         *http.Request
				expectedResponse *cache.Response
				expectedError    error
				rowsReturned     int
			}

			tests := []runTest{
				{
					label:           "Playlist:404",
					inputURL:        utils.MustParseURL("https://youtube.com/playlist?list=404"),
					inputPlaylistID: "playlist:404",
					inputReq:        nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":404,"message":"No YouTube playlist with the ID 404 found"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
				},
				{
					label:           "Warframe playlist",
					inputURL:        utils.MustParseURL("https://youtube.com/playlist?list=warframe"),
					inputPlaylistID: "playlist:warframe",
					inputReq:        nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"maxres-url","tooltip":"\u003cdiv style=\"text-align: left;\"\u003e\n\u003cb\u003eCool Warframe playlist\u003c/b\u003e\n\u003cbr\u003e\u003cb\u003eDescription:\u003c/b\u003e Very cool videos about Warframe\n\u003cbr\u003e\u003cb\u003eChannel:\u003c/b\u003e Warframe Highlights\n\u003cbr\u003e\u003cb\u003ePublishing Date:\u003c/b\u003e 2020-10-12T07:20:50.52Z\n\u003c/div\u003e\n"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
				},
			}

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
					pool.ExpectExec("INSERT INTO cache").
						WithArgs("youtube:playlist:"+test.inputPlaylistID, test.expectedResponse.Payload, test.expectedResponse.StatusCode, test.expectedResponse.ContentType, pgxmock.AnyArg()).
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
