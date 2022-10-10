package twitch

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/internal/mocks"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/utils"
	qt "github.com/frankban/quicktest"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v4"
	"github.com/nicklaw5/helix"
	"github.com/pashagolub/pgxmock"
)

func TestClipResolver(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := qt.New(t)

	pool, _ := pgxmock.NewPool()
	cfg := config.APIConfig{}
	helixClient := mocks.NewMockTwitchAPIClient(ctrl)

	resolver := NewClipResolver(ctx, cfg, pool, helixClient)

	c.Assert(resolver, qt.IsNotNil)

	c.Run("Name", func(c *qt.C) {
		c.Assert(resolver.Name(), qt.Equals, "twitch:clip")
	})

	c.Run("Check", func(c *qt.C) {
		type checkTest struct {
			label    string
			input    *url.URL
			expected bool
		}

		tests := []checkTest{}

		for _, b := range validClipBase {
			tests = append(tests, checkTest{
				label:    "valid",
				input:    utils.MustParseURL(b + goodSlugV1),
				expected: true,
			})
			tests = append(tests, checkTest{
				label:    "valid",
				input:    utils.MustParseURL(b + goodSlugV2),
				expected: true,
			})
		}

		for _, b := range invalidClips {
			tests = append(tests, checkTest{
				label:    "invalid",
				input:    utils.MustParseURL(b),
				expected: false,
			})
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
					inputURL:      utils.MustParseURL("https://clips.twitch.tv/user/566ca04265dbbdab32ec054a"),
					expectedError: errInvalidTwitchClip,
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

		c.Run("Not cached", func(c *qt.C) {
			type runTest struct {
				label                string
				inputURL             *url.URL
				inputSlug            string
				inputReq             *http.Request
				expectedClipResponse *helix.ClipsResponse
				expectedClipError    error
				expectedResponse     *cache.Response
				expectedError        error
				rowsReturned         int
			}

			tests := []runTest{
				{
					label:     "Emote",
					inputURL:  utils.MustParseURL("https://clips.twitch.tv/GoodSlugV1"),
					inputSlug: "GoodSlugV1",
					inputReq:  nil,
					expectedClipResponse: &helix.ClipsResponse{
						Data: helix.ManyClips{
							Clips: []helix.Clip{
								{
									Title:           "Title",
									CreatorName:     "CreatorName",
									BroadcasterName: "BroadcasterName",
									Duration:        5,
									CreatedAt:       "202", // will fail
									ViewCount:       420,
									ThumbnailURL:    "https://example.com/thumbnail.png",
								},
							},
						},
					},
					expectedClipError: nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://example.com/thumbnail.png","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%3Cb%3ETitle%3C%2Fb%3E%3Chr%3E%3Cb%3EClipped%20by:%3C%2Fb%3E%20CreatorName%3Cbr%3E%3Cb%3EChannel:%3C%2Fb%3E%20BroadcasterName%3Cbr%3E%3Cb%3EDuration:%3C%2Fb%3E%205s%3Cbr%3E%3Cb%3ECreated:%3C%2Fb%3E%20%3Cbr%3E%3Cb%3EViews:%3C%2Fb%3E%20420%3C%2Fdiv%3E"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:                "GetClipsError",
					inputURL:             utils.MustParseURL("https://clips.twitch.tv/GoodSlugV1GetClipsError"),
					inputSlug:            "GoodSlugV1GetClipsError",
					inputReq:             nil,
					expectedClipResponse: nil,
					expectedClipError:    errors.New("error"),
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":500,"message":"Twitch clip load error: error"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				// {
				// 	label:         "Bad JSON",
				// 	inputURL:      utils.MustParseURL("https://betterttv.com/emotes/bad"),
				// 	inputSlug:     "bad",
				// 	inputReq:      nil,
				// 	expectedBytes: []byte(`{"status":500,"message":"betterttv api unmarshal error: invalid character \u0026#39;x\u0026#39; looking for beginning of value"}`),
				// 	expectedError: nil,
				// },
			}

			const q = `SELECT value FROM cache WHERE key=$1`

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					helixClient.EXPECT().GetClips(&helix.ClipsParams{IDs: []string{test.inputSlug}}).Times(1).Return(test.expectedClipResponse, test.expectedClipError)
					pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
					pool.ExpectExec("INSERT INTO cache").
						WithArgs("twitch:clip:"+test.inputSlug, test.expectedResponse.Payload, test.expectedResponse.StatusCode, test.expectedResponse.ContentType, pgxmock.AnyArg()).
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
