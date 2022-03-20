package wikipedia

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	qt "github.com/frankban/quicktest"
)

func testLoadAndUnescape(ctx context.Context, loader *ArticleLoader, c *qt.C, locale, page string) (cleanTooltip string) {
	urlString := fmt.Sprintf("https://%s.wikipedia.org/wiki/%s", locale, page)
	response, _, err := loader.Load(ctx, urlString, nil)

	c.Assert(err, qt.IsNil)
	c.Assert(response, qt.Not(qt.IsNil))

	cleanTooltip, unescapeErr := url.PathUnescape(response.Tooltip)
	c.Assert(unescapeErr, qt.IsNil)

	return cleanTooltip
}

func TestLoad(t *testing.T) {
	// ctx := logger.OnContext(context.Background(), logger.NewTest())
	// c := qt.New(t)
	// ts := testServer()
	// defer ts.Close()

	// loader := &ArticleLoader{
	// 	apiURL: ts.URL + "/api/rest_v1/page/summary/%s/%s",
	// }

	// c.Run("Normal page", func(c *qt.C) {
	// 	const locale = "en"
	// 	const page = "test"

	// 	const expectedTooltip = `<div style="text-align: left;"><b>Test title&nbsp;•&nbsp;Test description</b><br>Test extract</div>`

	// 	cleanTooltip := testLoadAndUnescape(ctx, loader, c, locale, page)

	// 	c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	// })

	// c.Run("Normal page (HTML)", func(c *qt.C) {
	// 	const locale = "en"
	// 	const page = "test_html"

	// 	const expectedTooltip = `<div style="text-align: left;"><b>&lt;b&gt;Test title&lt;/b&gt;&nbsp;•&nbsp;&lt;b&gt;Test description&lt;/b&gt;</b><br>&lt;b&gt;Test extract&lt;/b&gt;</div>`

	// 	cleanTooltip := testLoadAndUnescape(ctx, loader, c, locale, page)

	// 	c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	// })

	// c.Run("Normal page (No description)", func(c *qt.C) {
	// 	const locale = "en"
	// 	const page = "test_no_description"

	// 	const expectedTooltip = `<div style="text-align: left;"><b>Test title</b><br>Test extract</div>`

	// 	cleanTooltip := testLoadAndUnescape(ctx, loader, c, locale, page)

	// 	c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	// })

	// c.Run("Nonexistant page", func(c *qt.C) {
	// 	const locale = "en"
	// 	const page = "404"

	// 	const expectedTooltip = `404`

	// 	cleanTooltip := testLoadAndUnescape(c, locale, page)

	// 	c.Assert(cleanTooltip, qt.Equals, expectedTooltip)
	// })
}
