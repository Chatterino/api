package twitch

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

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

func TestUserResolver(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := qt.New(t)

	pool, _ := pgxmock.NewPool()
	cfg := config.APIConfig{}
	helixClient := mocks.NewMockTwitchAPIClient(ctrl)

	resolver := NewUserResolver(ctx, cfg, pool, helixClient)

	c.Assert(resolver, qt.IsNotNil)

	c.Run("Name", func(c *qt.C) {
		c.Assert(resolver.Name(), qt.Equals, "twitch:user")
	})

	c.Run("Check", func(c *qt.C) {
		type checkTest struct {
			label    string
			input    *url.URL
			expected bool
		}

		tests := []checkTest{}

		for _, b := range validUsers {
			tests = append(tests, checkTest{
				label:    "valid",
				input:    utils.MustParseURL(b),
				expected: true,
			})
		}

		for _, b := range invalidUsers {
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
		c.Run("Not cached", func(c *qt.C) {
			type runTest struct {
				label                 string
				inputURL              *url.URL
				login                 string
				inputReq              *http.Request
				expectedUsersResponse *helix.UsersResponse
				expectedUserError     error
				expectedResponse      *cache.Response
				expectedError         error
				rowsReturned          int
			}

			tests := []runTest{
				{
					label:    "twitch",
					inputURL: utils.MustParseURL("https://twitch.tv/twitch"),
					login:    "twitch",
					inputReq: nil,
					expectedUsersResponse: &helix.UsersResponse{
						Data: helix.ManyUsers{
							Users: []helix.User{
								helix.User{
									Login:       "twitch",
									DisplayName: "Twitch",
									CreatedAt: helix.Time{
										Time: time.Date(2007, 5, 22, 0, 0, 0, 0, time.UTC),
									},
									Description:     "Twitch is where thousands of communities come together for whatever, every day. ",
									ProfileImageURL: "https://example.com/thumbnail.png",
								},
							},
						},
					},
					expectedUserError: nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://example.com/thumbnail.png","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%3Cb%3ETwitch%20-%20Twitch%3C%2Fb%3E%3Cbr%3ETwitch%20is%20where%20thousands%20of%20communities%20come%20together%20for%20whatever%2C%20every%20day.%20%3Cbr%3E%3Cb%3ECreated:%3C%2Fb%3E%2022%20May%202007%3Cbr%3E%3Cb%3EURL:%3C%2Fb%3E%20https:%2F%2Ftwitch.tv%2Ftwitch%3C%2Fdiv%3E"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
			}

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					helixClient.EXPECT().GetUsers(&helix.UsersParams{Logins: []string{test.login}}).Times(1).Return(test.expectedUsersResponse, test.expectedUserError)
					pool.ExpectQuery("SELECT").WillReturnError(pgx.ErrNoRows)
					pool.ExpectExec("INSERT INTO cache").
						WithArgs("twitch:user:"+test.login, test.expectedResponse.Payload, test.expectedResponse.StatusCode, test.expectedResponse.ContentType, pgxmock.AnyArg()).
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
