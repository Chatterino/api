package twitch

import (
	"net/url"
	"testing"

	"github.com/Chatterino/api/internal/mocks"
	"github.com/Chatterino/api/pkg/resolver"
	qt "github.com/frankban/quicktest"
	"github.com/golang/mock/gomock"
	"github.com/nicklaw5/helix"
)

func testLoadAndUnescape(c *qt.C, clipSlug string) (cleanTooltip string) {
	iret, _, err := load(clipSlug, nil)

	c.Assert(err, qt.IsNil)
	c.Assert(iret, qt.Not(qt.IsNil))

	response := iret.(*resolver.Response)

	c.Assert(response, qt.Not(qt.IsNil))

	cleanTooltip, unescapeErr := url.PathUnescape(response.Tooltip)
	c.Assert(unescapeErr, qt.IsNil)

	return cleanTooltip
}

func TestLoad(t *testing.T) {
	c := qt.New(t)
	mockCtrl := gomock.NewController(c)
	m := mocks.NewMockTwitchAPIClient(mockCtrl)
	helixAPI = m

	c.Run("Normal clip", func(c *qt.C) {
		const slug = "KKona"

		clip := helix.Clip{
			Title:           "Clipped it LUL",
			BroadcasterName: "pajlada",
			CreatorName:     "supinic",
			Duration:        30,
			CreatedAt:       "2019-11-14T04:20:06.09Z",
			ViewCount:       69,
		}

		response := &helix.ClipsResponse{}
		response.Data.Clips = []helix.Clip{clip}

		m.
			EXPECT().
			GetClips(gomock.Eq(&helix.ClipsParams{IDs: []string{slug}})).
			Return(response, nil)

		const expectedTooltip = `<div style="text-align: left;"><b>Clipped it LUL</b><hr><b>Clipped by:</b> supinic<br><b>Channel:</b> pajlada<br><b>Duration:</b> 30s<br><b>Created:</b> 14 Nov 2019<br><b>Views:</b> 69</div>`

		cleanTooltip := testLoadAndUnescape(c, slug)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("Normal clip (Number formatting)", func(c *qt.C) {
		const slug = "KKaper"

		clip := helix.Clip{
			Title:           "Clipped it LUL",
			BroadcasterName: "pajlada",
			CreatorName:     "suspinic",
			Duration:        30.1,
			CreatedAt:       "2019-11-14T04:20:06.09Z",
			ViewCount:       6969,
		}

		response := &helix.ClipsResponse{}
		response.Data.Clips = []helix.Clip{clip}

		m.
			EXPECT().
			GetClips(gomock.Eq(&helix.ClipsParams{IDs: []string{slug}})).
			Return(response, nil)

		const expectedTooltip = `<div style="text-align: left;"><b>Clipped it LUL</b><hr><b>Clipped by:</b> suspinic<br><b>Channel:</b> pajlada<br><b>Duration:</b> 30s<br><b>Created:</b> 14 Nov 2019<br><b>Views:</b> 6,969</div>`

		cleanTooltip := testLoadAndUnescape(c, slug)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("Normal clip (HTML)", func(c *qt.C) {
		const slug = "KKool"

		clip := helix.Clip{
			Title:           "Clipped it <b>LUL</b>",
			BroadcasterName: "<b>pajlada</b>",
			CreatorName:     "<b>supinic</b>",
			Duration:        30,
			CreatedAt:       "2019-11-14T04:20:06.09Z",
			ViewCount:       69,
		}

		response := &helix.ClipsResponse{}
		response.Data.Clips = []helix.Clip{clip}

		m.
			EXPECT().
			GetClips(gomock.Eq(&helix.ClipsParams{IDs: []string{slug}})).
			Return(response, nil)

		const expectedTooltip = `<div style="text-align: left;"><b>Clipped it &lt;b&gt;LUL&lt;/b&gt;</b><hr><b>Clipped by:</b> &lt;b&gt;supinic&lt;/b&gt;<br><b>Channel:</b> &lt;b&gt;pajlada&lt;/b&gt;<br><b>Duration:</b> 30s<br><b>Created:</b> 14 Nov 2019<br><b>Views:</b> 69</div>`

		cleanTooltip := testLoadAndUnescape(c, slug)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})
}
