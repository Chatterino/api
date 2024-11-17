package imgur

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/internal/mocks"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
	qt "github.com/frankban/quicktest"
	"github.com/jackc/pgx/v4"
	"github.com/koffeinsource/go-imgur"
	"github.com/pashagolub/pgxmock"
	"go.uber.org/mock/gomock"
)

func testCheck(ctx context.Context, resolver resolver.Resolver, c *qt.C, urlString string) bool {
	u, err := url.Parse(urlString)
	c.Assert(u, qt.Not(qt.IsNil))
	c.Assert(err, qt.IsNil)

	_, result := resolver.Check(ctx, u)

	return result
}

func TestCheck(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

	resolver := &Resolver{}

	shouldCheck := []string{
		"https://imgur.com",
		"https://www.imgur.com",
		"https://i.imgur.com",
	}

	for _, u := range shouldCheck {
		c.Assert(testCheck(ctx, resolver, c, u), qt.IsTrue)
	}

	shouldNotCheck := []string{
		"https://imgurr.com",
		"https://www.imgur.bad.com",
		"https://iimgur.com",
		"https://google.com",
		"https://i.imgur.org",
	}

	for _, u := range shouldNotCheck {
		c.Assert(testCheck(ctx, resolver, c, u), qt.IsFalse)
	}
}

func TestResolver(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := qt.New(t)

	imgurClient := mocks.NewMockImgurClient(ctrl)

	datetime := time.Date(2019, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()

	pool, _ := pgxmock.NewPool()

	cfg := config.APIConfig{
		BaseURL: "https://example.com/",
	}

	r := NewResolver(ctx, cfg, pool, imgurClient)

	c.Assert(r, qt.IsNotNil)

	c.Run("Name", func(c *qt.C) {
		c.Assert(r.Name(), qt.Equals, "imgur")
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
				input:    utils.MustParseURL("https://imgur.com"),
				expected: true,
			},
			{
				label:    "Matching domain, WWW",
				input:    utils.MustParseURL("https://www.imgur.com"),
				expected: true,
			},
			{
				label:    "Matching subdomain",
				input:    utils.MustParseURL("https://m.imgur.com"),
				expected: true,
			},
			{
				label:    "Non-matching domain",
				input:    utils.MustParseURL("https://example.com/emotes/566ca04265dbbdab32ec054a"),
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
		c.Run("Dont handle", func(c *qt.C) {
			type runTest struct {
				label          string
				inputURL       *url.URL
				inputEmoteHash string
				inputReq       *http.Request
				err            error
				info           *imgur.GenericInfo
			}

			tests := []runTest{
				{
					label:    "client error",
					inputURL: utils.MustParseURL("https://betterttv.com/user/566ca04265dbbdab32ec054a"),
					err:      errors.New("asd"),
					info:     nil,
				},
				{
					label:    "client bad response",
					inputURL: utils.MustParseURL("https://betterttv.com/user/566ca04265dbbdab32ec054a"),
					err:      nil,
					info:     &imgur.GenericInfo{},
				},
				{
					label:    "client no error no response",
					inputURL: utils.MustParseURL("https://betterttv.com/user/566ca04265dbbdab32ec054a"),
					err:      nil,
					info:     nil,
				},
			}

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					imgurClient.
						EXPECT().
						GetInfoFromURL(test.inputURL.String()).
						Times(1).
						Return(test.info, 0, test.err)
					pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
					outputBytes, outputError := r.Run(ctx, test.inputURL, test.inputReq)
					c.Assert(outputError, qt.Equals, resolver.ErrDontHandle)
					c.Assert(outputBytes, qt.IsNil)
				})
			}
		})

		c.Run("Not cached", func(c *qt.C) {
			type runTest struct {
				label            string
				inputURL         *url.URL
				info             *imgur.GenericInfo
				expectedResponse *cache.Response
				expectedError    error
				rowsReturned     int
			}

			tests := []runTest{
				{
					label:    "A",
					inputURL: utils.MustParseURL("https://imgur.com/a"),
					info: &imgur.GenericInfo{
						Image: &imgur.ImageInfo{
							Title:       "My Cool Title",
							Description: "My Cool Description",
							Datetime:    int(datetime),
							Link:        "https://i.imgur.com/a.png",
						},
						Album:  nil,
						GImage: nil,
						GAlbum: nil,
						Limit:  &imgur.RateLimit{},
					},
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://example.com/thumbnail/https%3A%2F%2Fi.imgur.com%2Fa.png","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%3Cli%3E%3Cb%3ETitle:%3C%2Fb%3E%20My%20Cool%20Title%3C%2Fli%3E%3Cli%3E%3Cb%3EDescription:%3C%2Fb%3E%20My%20Cool%20Description%3C%2Fli%3E%3Cli%3E%3Cb%3EUploaded:%3C%2Fb%3E%2010%20Nov%202019%20%E2%80%A2%2023:00%20UTC%3C%2Fli%3E%3C%2Fdiv%3E"}`),
						StatusCode:  200,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:    "Link too big",
					inputURL: utils.MustParseURL("https://imgur.com/toobig"),
					info: &imgur.GenericInfo{
						Image: &imgur.ImageInfo{
							Title:       "My Cool Title",
							Description: "My Cool Description",
							Datetime:    int(datetime),
							Size:        maxRawImageSize + 1,
							Link:        "https://i.imgur.com/toobig.png",
						},
						Album:  nil,
						GImage: nil,
						GAlbum: nil,
						Limit:  &imgur.RateLimit{},
					},
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://example.com/thumbnail/https%3A%2F%2Fi.imgur.com%2Ftoobigl.png","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%3Cli%3E%3Cb%3ETitle:%3C%2Fb%3E%20My%20Cool%20Title%3C%2Fli%3E%3Cli%3E%3Cb%3EDescription:%3C%2Fb%3E%20My%20Cool%20Description%3C%2Fli%3E%3Cli%3E%3Cb%3EUploaded:%3C%2Fb%3E%2010%20Nov%202019%20%E2%80%A2%2023:00%20UTC%3C%2Fli%3E%3C%2Fdiv%3E"}`),
						StatusCode:  200,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:    "Link too big, malformed url",
					inputURL: utils.MustParseURL("https://imgur.com/toobigbadurl"),
					info: &imgur.GenericInfo{
						Image: &imgur.ImageInfo{
							Title:       "My Cool Title",
							Description: "My Cool Description",
							Datetime:    int(datetime),
							Size:        maxRawImageSize + 1,
							Link:        ":",
						},
						Album:  nil,
						GImage: nil,
						GAlbum: nil,
						Limit:  &imgur.RateLimit{},
					},
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%3Cli%3E%3Cb%3ETitle:%3C%2Fb%3E%20My%20Cool%20Title%3C%2Fli%3E%3Cli%3E%3Cb%3EDescription:%3C%2Fb%3E%20My%20Cool%20Description%3C%2Fli%3E%3Cli%3E%3Cb%3EUploaded:%3C%2Fb%3E%2010%20Nov%202019%20%E2%80%A2%2023:00%20UTC%3C%2Fli%3E%3C%2Fdiv%3E"}`),
						StatusCode:  200,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:    "Animated non-gif",
					inputURL: utils.MustParseURL("https://imgur.com/b"),
					info: &imgur.GenericInfo{
						Image: &imgur.ImageInfo{
							Title:       "My Cool Title",
							Description: "My Cool Description",
							Datetime:    int(datetime),
							Link:        "https://i.imgur.com/b.mp4",
							Animated:    true,
						},
						Album:  nil,
						GImage: nil,
						GAlbum: nil,
						Limit:  &imgur.RateLimit{},
					},
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://example.com/thumbnail/https%3A%2F%2Fi.imgur.com%2Fb.png","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%3Cli%3E%3Cb%3ETitle:%3C%2Fb%3E%20My%20Cool%20Title%3C%2Fli%3E%3Cli%3E%3Cb%3EDescription:%3C%2Fb%3E%20My%20Cool%20Description%3C%2Fli%3E%3Cli%3E%3Cb%3EUploaded:%3C%2Fb%3E%2010%20Nov%202019%20%E2%80%A2%2023:00%20UTC%3C%2Fli%3E%3Cli%3E%3Cb%3E%3Cspan%20style=%22color:%20red%3B%22%3EANIMATED%3C%2Fspan%3E%3C%2Fb%3E%3C%2Fli%3E%3C%2Fdiv%3E"}`),
						StatusCode:  200,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:    "Animated malformed link",
					inputURL: utils.MustParseURL("https://imgur.com/c"),
					info: &imgur.GenericInfo{
						Image: &imgur.ImageInfo{
							Title:       "My Cool Title",
							Description: "My Cool Description",
							Datetime:    int(datetime),
							Link:        ":",
							Animated:    true,
						},
						Album:  nil,
						GImage: nil,
						GAlbum: nil,
						Limit:  &imgur.RateLimit{},
					},
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%3Cli%3E%3Cb%3ETitle:%3C%2Fb%3E%20My%20Cool%20Title%3C%2Fli%3E%3Cli%3E%3Cb%3EDescription:%3C%2Fb%3E%20My%20Cool%20Description%3C%2Fli%3E%3Cli%3E%3Cb%3EUploaded:%3C%2Fb%3E%2010%20Nov%202019%20%E2%80%A2%2023:00%20UTC%3C%2Fli%3E%3Cli%3E%3Cb%3E%3Cspan%20style=%22color:%20red%3B%22%3EANIMATED%3C%2Fspan%3E%3C%2Fb%3E%3C%2Fli%3E%3C%2Fdiv%3E"}`),
						StatusCode:  200,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
			}

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					imgurClient.
						EXPECT().
						GetInfoFromURL(test.inputURL.String()).
						Times(1).
						Return(test.info, 0, nil)
					pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
					pool.ExpectExec("INSERT INTO cache").
						WithArgs("imgur:"+test.inputURL.String(), test.expectedResponse.Payload, test.expectedResponse.StatusCode, test.expectedResponse.ContentType, pgxmock.AnyArg()).
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
