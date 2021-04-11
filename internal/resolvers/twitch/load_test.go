package twitch

import (
	"net/url"
	"testing"
	"time"

	"github.com/Chatterino/api/internal/mocks"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/dankeroni/gotwitch"
	qt "github.com/frankban/quicktest"
	"github.com/golang/mock/gomock"
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
	v5API = m

	c.Run("Normal clip", func(c *qt.C) {
		const slug = "KKona"
		var clipResponse gotwitch.V5GetClipResponse
		clipResponse.Title = "Clipped it LUL"
		clipResponse.Broadcaster.DisplayName = "pajlada"
		clipResponse.Curator.DisplayName = "supinic"
		clipResponse.Duration = 30
		clipResponse.CreatedAt = time.Date(2019, time.November, 14, 04, 20, 6, 9, time.UTC)
		clipResponse.Views = 69

		m.
			EXPECT().
			GetClip(gomock.Eq(slug)).
			Return(clipResponse, nil, nil)

		const expectedTooltip = `<div style="text-align: left;"><b>Clipped it LUL</b><hr><b>Clipped by:</b> supinic<br><b>Channel:</b> pajlada<br><b>Duration:</b> 30s<br><b>Created:</b> 14 Nov 2019<br><b>Views:</b> 69</div>`

		cleanTooltip := testLoadAndUnescape(c, slug)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("Normal clip (Number formatting)", func(c *qt.C) {
		const slug = "KKona"
		var clipResponse gotwitch.V5GetClipResponse
		clipResponse.Title = "Clipped it LUL"
		clipResponse.Broadcaster.DisplayName = "pajlada"
		clipResponse.Curator.DisplayName = "supinic"
		clipResponse.Duration = 30
		clipResponse.CreatedAt = time.Date(2019, time.November, 14, 04, 20, 6, 9, time.UTC)
		clipResponse.Views = 6969

		m.
			EXPECT().
			GetClip(gomock.Eq(slug)).
			Return(clipResponse, nil, nil)

		const expectedTooltip = `<div style="text-align: left;"><b>Clipped it LUL</b><hr><b>Clipped by:</b> supinic<br><b>Channel:</b> pajlada<br><b>Duration:</b> 30s<br><b>Created:</b> 14 Nov 2019<br><b>Views:</b> 6,969</div>`

		cleanTooltip := testLoadAndUnescape(c, slug)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("Normal clip (HTML)", func(c *qt.C) {
		const slug = "KKona"
		var clipResponse gotwitch.V5GetClipResponse
		clipResponse.Title = "Clipped it <b>LUL</b>"
		clipResponse.Broadcaster.DisplayName = "<b>pajlada</b>"
		clipResponse.Curator.DisplayName = "<b>supinic</b>"
		clipResponse.Duration = 30
		clipResponse.CreatedAt = time.Date(2019, time.November, 14, 04, 20, 6, 9, time.UTC)
		clipResponse.Views = 69

		m.
			EXPECT().
			GetClip(gomock.Eq(slug)).
			Return(clipResponse, nil, nil)

		const expectedTooltip = `<div style="text-align: left;"><b>Clipped it &lt;b&gt;LUL&lt;/b&gt;</b><hr><b>Clipped by:</b> &lt;b&gt;supinic&lt;/b&gt;<br><b>Channel:</b> &lt;b&gt;pajlada&lt;/b&gt;<br><b>Duration:</b> 30s<br><b>Created:</b> 14 Nov 2019<br><b>Views:</b> 69</div>`

		cleanTooltip := testLoadAndUnescape(c, slug)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})
}
