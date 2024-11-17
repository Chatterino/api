package discord

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
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
)

func TestInviteResolver(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := qt.New(t)

	// pool := mocks.NewMockPool(ctrl)
	pool, _ := pgxmock.NewPool()

	cfg := config.APIConfig{}
	ts := testServer()
	defer ts.Close()
	emoteAPIURL := utils.MustParseURL(ts.URL + "/api/v9/invites/")

	resolver := NewInviteResolver(ctx, cfg, pool, emoteAPIURL)

	c.Assert(resolver, qt.IsNotNil)

	c.Run("Name", func(c *qt.C) {
		c.Assert(resolver.Name(), qt.Equals, "discord:invite")
	})

	c.Run("Check", func(c *qt.C) {
		type checkTest struct {
			label    string
			input    *url.URL
			expected bool
		}

		tests := []checkTest{
			{
				label:    "Matching domain 1, no WWW",
				input:    utils.MustParseURL("https://discord.gg/forsen"),
				expected: true,
			},
			{
				label:    "Matching domain 1, WWW",
				input:    utils.MustParseURL("https://www.discord.gg/forsen"),
				expected: true,
			},
			{
				label:    "Matching domain 2, no WWW",
				input:    utils.MustParseURL("https://discord.com/invite/forsen"),
				expected: true,
			},
			{
				label:    "Matching domain 2, WWW",
				input:    utils.MustParseURL("https://www.discord.com/invite/forsen"),
				expected: true,
			},
			{
				label:    "Matching domain, non-matching path",
				input:    utils.MustParseURL("https://discord.com/forsen"),
				expected: false,
			},
			{
				label:    "Non-matching domain",
				input:    utils.MustParseURL("https://forsen.tv/forsen"),
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
				label           string
				inputURL        *url.URL
				inputInviteCode string
				inputReq        *http.Request
				expectedError   error
			}

			tests := []runTest{
				{
					label:         "Non-matching link",
					inputURL:      utils.MustParseURL("https://discord.gg/_xXx_forsen_xXx_"),
					expectedError: errInvalidDiscordInvite,
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
		c.Run("Cached", func(c *qt.C) {
			type runTest struct {
				label           string
				inputURL        *url.URL
				inputInviteCode string
				inputReq        *http.Request
				// expectedResponse will be returned from the cache, and expected to be returned in the same form
				expectedResponse *cache.Response
				expectedError    error
			}

			tests := []runTest{
				{
					label:           "Matching link - cached 1",
					inputURL:        utils.MustParseURL("https://discord.gg/forsen"),
					inputInviteCode: "forsen",
					inputReq:        nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://cdn.discordapp.com/icons/97034666673975296/a_ea433153b6ce120e0fb518efc084dc38","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EForsen%3C%2Fb%3E%0A%3Cbr%3E%0A%3Cbr%3E%3Cb%3EServer%20Created:%3C%2Fb%3E%2025%20Sep%202015%0A%3Cbr%3E%3Cb%3EChannel:%3C%2Fb%3E%20%23readme%0A%0A%3Cbr%3E%3Cb%3EServer%20Perks:%3C%2Fb%3E%20banner%2C%20vanity%20url%2C%20invite%20splash%2C%20animated%20icon%0A%3Cbr%3E%3Cb%3EMembers:%3C%2Fb%3E%20%3Cspan%20style=%22color:%20%2343b581%3B%22%3E13%2C465%20online%3C%2Fspan%3E\u0026nbsp%3B%E2%80%A2\u0026nbsp%3B%3Cspan%20style=%22color:%20%23808892%3B%22%3E44%2C961%20total%3C%2Fspan%3E%0A%3C%2Fdiv%3E%0A","link":"https://discord.gg/forsen"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:           "Matching link - cached 2",
					inputURL:        utils.MustParseURL("https://discord.com/invite/qbRE8WR"),
					inputInviteCode: "qbRE8WR",
					inputReq:        nil,
					expectedResponse: &cache.Response{
						Payload: []byte(`{"status":200,"thumbnail":"https://cdn.discordapp.com/icons/138009976613502976/dcbac612ccdd3ffa2fbf89647e26f929","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3Epajlada%3C%2Fb%3E%0A%3Cbr%3E%0A%3Cbr%3E%3Cb%3EServer%20Created:%3C%2Fb%3E%2016%20Jan%202016%0A%3Cbr%3E%3Cb%3EChannel:%3C%2Fb%3E%20%23general%0A%3Cbr%3E%3Cb%3EInviter:%3C%2Fb%3E%20pajlada%230%0A%3Cbr%3E%3Cb%3EServer%20Perks:%3C%2Fb%3E%20invite%20splash%2C%20animated%20icon%2C%20community%0A%3Cbr%3E%3Cb%3EMembers:%3C%2Fb%3E%20%3Cspan%20style=%22color:%20%2343b581%3B%22%3E546%20online%3C%2Fspan%3E\u0026nbsp%3B%E2%80%A2\u0026nbsp%3B%3Cspan%20style=%22color:%20%23808892%3B%22%3E1%2C515%20total%3C%2Fspan%3E%0A%3C%2Fdiv%3E%0A","link":"https://discord.gg/qbRE8WR"}
`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:           "Matching link - 404",
					inputURL:        utils.MustParseURL("https://discord.gg/404"),
					inputInviteCode: "404",
					inputReq:        nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":404,"message":"No Discord invite with this code found"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
			}

			for _, test := range tests {
				c.Run(test.label, func(c *qt.C) {
					rows := pgxmock.NewRows([]string{"value", "http_status_code", "http_content_type"}).AddRow(test.expectedResponse.Payload, http.StatusOK, test.expectedResponse.ContentType)
					pool.ExpectQuery("SELECT").
						WithArgs("discord:invite:" + test.inputInviteCode).
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
					label:          "Forsen",
					inputURL:       utils.MustParseURL("https://discord.gg/forsen"),
					inputEmoteHash: "forsen",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://cdn.discordapp.com/icons/97034666673975296/a_ea433153b6ce120e0fb518efc084dc38","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3EForsen%3C%2Fb%3E%0A%3Cbr%3E%0A%3Cbr%3E%3Cb%3EServer%20Created:%3C%2Fb%3E%2025%20Sep%202015%0A%3Cbr%3E%3Cb%3EChannel:%3C%2Fb%3E%20%23readme%0A%0A%3Cbr%3E%3Cb%3EServer%20Perks:%3C%2Fb%3E%20animated%20icon%2C%20banner%2C%20invite%20splash%2C%20vanity%20url%0A%3Cbr%3E%3Cb%3EMembers:%3C%2Fb%3E%20%3Cspan%20style=%22color:%20%2343b581%3B%22%3E13%2C730%20online%3C%2Fspan%3E\u0026nbsp%3B%E2%80%A2\u0026nbsp%3B%3Cspan%20style=%22color:%20%23808892%3B%22%3E44%2C960%20total%3C%2Fspan%3E%0A%3C%2Fdiv%3E%0A","link":"https://discord.gg/forsen"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Pajlada",
					inputURL:       utils.MustParseURL("https://discord.com/invite/qbRE8WR"),
					inputEmoteHash: "qbRE8WR",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":200,"thumbnail":"https://cdn.discordapp.com/icons/138009976613502976/dcbac612ccdd3ffa2fbf89647e26f929","tooltip":"%3Cdiv%20style=%22text-align:%20left%3B%22%3E%0A%3Cb%3Epajlada%3C%2Fb%3E%0A%3Cbr%3E%0A%3Cbr%3E%3Cb%3EServer%20Created:%3C%2Fb%3E%2016%20Jan%202016%0A%3Cbr%3E%3Cb%3EChannel:%3C%2Fb%3E%20%23general%0A%3Cbr%3E%3Cb%3EInviter:%3C%2Fb%3E%20pajlada%230%0A%3Cbr%3E%3Cb%3EServer%20Perks:%3C%2Fb%3E%20animated%20icon%2C%20community%2C%20invite%20splash%0A%3Cbr%3E%3Cb%3EMembers:%3C%2Fb%3E%20%3Cspan%20style=%22color:%20%2343b581%3B%22%3E563%20online%3C%2Fspan%3E\u0026nbsp%3B%E2%80%A2\u0026nbsp%3B%3Cspan%20style=%22color:%20%23808892%3B%22%3E1%2C515%20total%3C%2Fspan%3E%0A%3C%2Fdiv%3E%0A","link":"https://discord.gg/qbRE8WR"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "404",
					inputURL:       utils.MustParseURL("https://discord.gg/404"),
					inputEmoteHash: "404",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":404,"message":"No Discord invite with this code found"}`),
						StatusCode:  http.StatusOK,
						ContentType: "application/json",
					},
					expectedError: nil,
				},
				{
					label:          "Bad JSON",
					inputURL:       utils.MustParseURL("https://discord.gg/bad"),
					inputEmoteHash: "bad",
					inputReq:       nil,
					expectedResponse: &cache.Response{
						Payload:     []byte(`{"status":500,"message":"Discord API unmarshal error invalid character \u0026#39;x\u0026#39; looking for beginning of value"}`),
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
						WithArgs("discord:invite:"+test.inputEmoteHash, test.expectedResponse.Payload, http.StatusOK, test.expectedResponse.ContentType, pgxmock.AnyArg()).
						WillReturnResult(pgxmock.NewResult("INSERT", 1))
					outputBytes, outputError := resolver.Run(ctx, test.inputURL, test.inputReq)
					c.Assert(outputError, qt.Equals, test.expectedError)
					c.Assert(outputBytes, qt.DeepEquals, test.expectedResponse)
				})
			}
		})
	})
}
