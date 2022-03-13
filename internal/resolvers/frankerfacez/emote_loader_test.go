package frankerfacez

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/utils"

	qt "github.com/frankban/quicktest"
)

func TestFoo(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

	ts := testServer()
	defer ts.Close()
	emoteAPIURL := utils.MustParseURL(ts.URL + "/v1/emote/")
	loader := NewEmoteLoader(emoteAPIURL)

	response, _, err := loader.Load(ctx, "kkona", nil)

	c.Assert(err, qt.IsNil)
	c.Assert(response, qt.Not(qt.IsNil))

	c.Assert(response.Status, qt.Equals, 200)
	c.Assert(response.Thumbnail, qt.Equals, fmt.Sprintf(thumbnailFormat, "kkona"))

	const expectedTooltip = `<div style="text-align: left;">
<b>KKona</b><br>
<b>FrankerFaceZ Emote</b><br>
<b>By:</b> zneix</div>`

	// TODO: check error
	cleanTooltip, unescapeErr := url.PathUnescape(response.Tooltip)
	c.Assert(unescapeErr, qt.IsNil)

	c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
}
