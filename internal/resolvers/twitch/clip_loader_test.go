package twitch

import (
	"context"
	"net/url"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/internal/mocks"
	qt "github.com/frankban/quicktest"
	"github.com/nicklaw5/helix"
	"go.uber.org/mock/gomock"
)

func testLoadAndUnescape(ctx context.Context, loader *ClipLoader, c *qt.C, clipSlug string) (string, int, string) {
	response, _, err := loader.Load(ctx, clipSlug, nil)

	c.Assert(err, qt.IsNil)
	c.Assert(response, qt.Not(qt.IsNil))

	cleanTooltip, unescapeErr := url.PathUnescape(response.Tooltip)
	c.Assert(unescapeErr, qt.IsNil)

	return cleanTooltip, response.Status, response.Message
}

func TestLoad(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)
	mockCtrl := gomock.NewController(c)
	m := mocks.NewMockTwitchAPIClient(mockCtrl)

	loader := &ClipLoader{
		helixAPI: m,
	}

	tests := []struct {
		label           string
		slug            string
		clip            []helix.Clip
		expectedTooltip string
		status          int
		errorMessage    string
	}{
		{
			"Normal clip",
			"KKona",
			[]helix.Clip{{
				Title:           "Clipped it LUL",
				BroadcasterName: "pajlada",
				CreatorName:     "supinic",
				Duration:        30,
				CreatedAt:       "2019-11-14T04:20:06.09Z",
				ViewCount:       69,
			}},
			`<div style="text-align: left;"><b>Clipped it LUL</b><hr><b>Clipped by:</b> supinic<br><b>Channel:</b> pajlada<br><b>Duration:</b> 30s<br><b>Created:</b> 14 Nov 2019<br><b>Views:</b> 69</div>`,
			200,
			"",
		},
		{
			"Normal clip (Number formatting)",
			"KKaper",
			[]helix.Clip{{
				Title:           "Clipped it LUL",
				BroadcasterName: "pajlada",
				CreatorName:     "suspinic",
				Duration:        30.1,
				CreatedAt:       "2019-11-14T04:20:06.09Z",
				ViewCount:       6969,
			}},
			`<div style="text-align: left;"><b>Clipped it LUL</b><hr><b>Clipped by:</b> suspinic<br><b>Channel:</b> pajlada<br><b>Duration:</b> 30s<br><b>Created:</b> 14 Nov 2019<br><b>Views:</b> 6,969</div>`,
			200,
			"",
		},
		{
			"Normal clip (HTML)",
			"KKool",
			[]helix.Clip{{
				Title:           "Clipped it <b>LUL</b>",
				BroadcasterName: "<b>pajlada</b>",
				CreatorName:     "<b>supinic</b>",
				Duration:        30,
				CreatedAt:       "2019-11-14T04:20:06.09Z",
				ViewCount:       69,
			}},
			`<div style="text-align: left;"><b>Clipped it &lt;b&gt;LUL&lt;/b&gt;</b><hr><b>Clipped by:</b> &lt;b&gt;supinic&lt;/b&gt;<br><b>Channel:</b> &lt;b&gt;pajlada&lt;/b&gt;<br><b>Duration:</b> 30s<br><b>Created:</b> 14 Nov 2019<br><b>Views:</b> 69</div>`,
			200,
			"",
		},
		{
			"No clip",
			"KKorner",
			[]helix.Clip{},
			"",
			404,
			"No Twitch Clip with this ID found",
		},
	}

	for _, test := range tests {
		c.Run(test.label, func(c *qt.C) {
			response := &helix.ClipsResponse{}
			response.Data.Clips = test.clip

			m.
				EXPECT().
				GetClips(gomock.Eq(&helix.ClipsParams{IDs: []string{test.slug}})).
				Return(response, nil)

			cleanTooltip, status, errorMessage := testLoadAndUnescape(ctx, loader, c, test.slug)

			c.Assert(cleanTooltip, qt.Equals, test.expectedTooltip)
			c.Assert(status, qt.Equals, test.status)
			c.Assert(errorMessage, qt.Equals, test.errorMessage)
		})
	}
}
