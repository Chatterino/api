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
		var clip helix.Clip
		clip.Title = "Clipped it LUL"
		clip.BroadcasterName = "pajlada"
		clip.CreatorName = "supinic"
		//clipResponse.Duration = 30
		clip.CreatedAt = "2019-11-14T04:20:06.09Z"
		clip.ViewCount = 69

		response := &helix.ClipsResponse{}
		//response.Data.Clips = (*helix.ManyClips).Clips

		//response.Data.Clips[0] = clip

		//clips := &ClipsResponse{}
		//resp.HydrateResponseCommon(&clips.ResponseCommon)
		//clips.Data.Clips = resp.Data.(*ManyClips).Clips
		//clips.Data.Pagination = resp.Data.(*ManyClips).Pagination

		m.
			EXPECT().
			GetClips(gomock.Eq(&helix.ClipsParams{IDs: []string{slug}})).
			Return(response, nil)

		const expectedTooltip = `<div style="text-align: left;"><b>Clipped it LUL</b><hr><b>Clipped by:</b> supinic<br><b>Channel:</b> pajlada<br><b>Duration:</b> 30s<br><b>Created:</b> 14 Nov 2019<br><b>Views:</b> 69</div>`

		cleanTooltip := testLoadAndUnescape(c, slug)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("Normal clip (Number formatting)", func(c *qt.C) {
		const slug = "KKona"
		var clip helix.Clip
		clip.Title = "Clipped it LUL"
		clip.BroadcasterName = "pajlada"
		clip.CreatorName = "supinic"
		//clipResponse.Duration = 30
		clip.CreatedAt = "2019-11-14T04:20:06.09Z"
		clip.ViewCount = 6969

		m.
			EXPECT().
			GetClips(gomock.Eq(&helix.ClipsParams{IDs: []string{slug}})).
			Return(clip, nil)

		const expectedTooltip = `<div style="text-align: left;"><b>Clipped it LUL</b><hr><b>Clipped by:</b> supinic<br><b>Channel:</b> pajlada<br><b>Duration:</b> 30s<br><b>Created:</b> 14 Nov 2019<br><b>Views:</b> 6,969</div>`

		cleanTooltip := testLoadAndUnescape(c, slug)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})

	c.Run("Normal clip (HTML)", func(c *qt.C) {
		const slug = "KKona"
		var clip helix.Clip
		clip.Title = "Clipped it <b>LUL</b>"
		clip.BroadcasterName = "pajlada"
		clip.CreatorName = "supinic"
		//clipResponse.Duration = 30
		clip.CreatedAt = "2019-11-14T04:20:06.09Z"
		clip.ViewCount = 69

		m.
			EXPECT().
			GetClips(gomock.Eq(&helix.ClipsParams{IDs: []string{slug}})).
			Return(clip, nil)

		const expectedTooltip = `<div style="text-align: left;"><b>Clipped it &lt;b&gt;LUL&lt;/b&gt;</b><hr><b>Clipped by:</b> &lt;b&gt;supinic&lt;/b&gt;<br><b>Channel:</b> &lt;b&gt;pajlada&lt;/b&gt;<br><b>Duration:</b> 30s<br><b>Created:</b> 14 Nov 2019<br><b>Views:</b> 69</div>`

		cleanTooltip := testLoadAndUnescape(c, slug)

		c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	})
}
