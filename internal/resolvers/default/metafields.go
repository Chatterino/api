package defaultresolver

import (
	"net/http"

	"github.com/Chatterino/api/pkg/utils"
	"github.com/PuerkitoBio/goquery"
)

func tooltipMetaFields(baseURL string, doc *goquery.Document, r *http.Request, resp *http.Response, data tooltipData) tooltipData {
	fields := doc.Find("meta[property][content]")

	if fields.Size() > 0 {
		fields.Each(func(i int, s *goquery.Selection) {
			prop, _ := s.Attr("property")
			cont, _ := s.Attr("content")

			switch {
			/* Support for HTML Open Graph & Twitter meta tags.
			 * Will show Open Graph & Twitter "Title", "Description", "Image" information of webpages.
			 * More OG fields & information: https://ogp.me/
			 * More Twitter fields & information: https://developer.twitter.com/en/docs/twitter-for-websites/cards/overview/markup
			 */
			case (prop == "og:title" || prop == "twitter:title") && data.Title == "":
				data.Title = cont
			case (prop == "og:description" || prop == "twitter:description") && data.Description == "":
				data.Description = cont
			case (prop == "og:image" || prop == "twitter:image") && data.ImageSrc == "":
				data.ImageSrc = utils.FormatThumbnailURL(baseURL, r, cont)
			}
		})
	}

	return data
}
