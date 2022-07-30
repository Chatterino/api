package defaultresolver

import (
	"context"
	"strings"
	"testing"

	"github.com/Chatterino/api/internal/logger"
	"github.com/PuerkitoBio/goquery"
	qt "github.com/frankban/quicktest"
)

func mustDoc(doc *goquery.Document, err error) *goquery.Document {
	if err != nil {
		panic(err)
	}

	return doc
}

func TestMetaFields(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

	testBaseURL := "https://pajlada.se/"
	testRequest := newLinkResolverRequest(t, ctx, "GET", "https://pajlada.se", nil)

	c.Run("Meta fields", func(c *qt.C) {
		tests := []struct {
			inputDoc        *goquery.Document
			inputTooltip    tooltipData
			expectedTooltip tooltipData
		}{
			{
				inputDoc:        mustDoc(goquery.NewDocumentFromReader(strings.NewReader(`<html><head><meta property="og:title" content="The Rock" /></head><body>xD</body></html>`))),
				inputTooltip:    tooltipData{},
				expectedTooltip: tooltipData{Title: "The Rock"},
			},
			{
				inputDoc:        mustDoc(goquery.NewDocumentFromReader(strings.NewReader(`<html><head><meta property="twitter:title" content="The Rock" /></head><body>xD</body></html>`))),
				inputTooltip:    tooltipData{},
				expectedTooltip: tooltipData{Title: "The Rock"},
			},
			{
				inputDoc:        mustDoc(goquery.NewDocumentFromReader(strings.NewReader(`<html><head><meta property="og:title" content="The Rock" /><meta property="twitter:title" content="The Rock 2" /></head><body>xD</body></html>`))),
				inputTooltip:    tooltipData{},
				expectedTooltip: tooltipData{Title: "The Rock"},
			},
			{
				inputDoc:        mustDoc(goquery.NewDocumentFromReader(strings.NewReader(`<html><head><meta property="og:description" content="The Rock" /></head><body>xD</body></html>`))),
				inputTooltip:    tooltipData{},
				expectedTooltip: tooltipData{Description: "The Rock"},
			},
			{
				inputDoc:        mustDoc(goquery.NewDocumentFromReader(strings.NewReader(`<html><head><meta property="twitter:description" content="The Rock" /></head><body>xD</body></html>`))),
				inputTooltip:    tooltipData{},
				expectedTooltip: tooltipData{Description: "The Rock"},
			},
			{
				inputDoc:        mustDoc(goquery.NewDocumentFromReader(strings.NewReader(`<html><head><meta property="og:description" content="The Rock" /><meta property="twitter:description" content="The Rock 2" /></head><body>xD</body></html>`))),
				inputTooltip:    tooltipData{},
				expectedTooltip: tooltipData{Description: "The Rock"},
			},
			{
				inputDoc:        mustDoc(goquery.NewDocumentFromReader(strings.NewReader(`<html><head><meta property="og:image" content="https://pajlada.se/thumbnail.png" /></head><body>xD</body></html>`))),
				inputTooltip:    tooltipData{},
				expectedTooltip: tooltipData{ImageSrc: "https://pajlada.se/thumbnail/https%3A%2F%2Fpajlada.se%2Fthumbnail.png"},
			},
			{
				inputDoc:        mustDoc(goquery.NewDocumentFromReader(strings.NewReader(`<html><head><meta property="twitter:image" content="https://pajlada.se/thumbnail.png" /></head><body>xD</body></html>`))),
				inputTooltip:    tooltipData{},
				expectedTooltip: tooltipData{ImageSrc: "https://pajlada.se/thumbnail/https%3A%2F%2Fpajlada.se%2Fthumbnail.png"},
			},
			{
				inputDoc:        mustDoc(goquery.NewDocumentFromReader(strings.NewReader(`<html><head><meta property="og:image" content="https://pajlada.se/thumbnail.png" /><meta property="twitter:image" content="https://pajlada.se/thumbnail2.png" /></head><body>xD</body></html>`))),
				inputTooltip:    tooltipData{},
				expectedTooltip: tooltipData{ImageSrc: "https://pajlada.se/thumbnail/https%3A%2F%2Fpajlada.se%2Fthumbnail.png"},
			},
		}

		for _, test := range tests {
			c.Run("", func(c *qt.C) {
				outputTooltip := tooltipMetaFields(testBaseURL, test.inputDoc, testRequest, nil, test.inputTooltip)
				c.Assert(outputTooltip, qt.Equals, test.expectedTooltip)
			})
		}
	})
}
