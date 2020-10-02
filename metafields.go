package main

import (
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func metaFields(doc *goquery.Document, r *http.Request, resp *http.Response, data tooltipData) tooltipData {
	fields := doc.Find("meta[property][content]")

	if fields.Size() > 0 {
		fields.Each(func(i int, s *goquery.Selection) {
			prop, _ := s.Attr("property")
			cont, _ := s.Attr("content")

			/* Support for HTML Open Graph meta tags.
			 * Will show Open Graph "Title", "Description", "Image" information of webpages.
			 * More fields & information: https://ogp.me/
			 */
			if prop == "og:title" {
				data.Title = cont
			} else if prop == "og:description" {
				data.Description = cont
			} else if prop == "og:image" {
				data.ImageSrc = formatThumbnailUrl(r, cont)

				/* Support for HTML Twitter meta tags.
				 * Will show Twitter "Title", "Description", "Image", information of webpages.
				 * More fields & information: https://developer.twitter.com/en/docs/twitter-for-websites/cards/overview/markup
				 * OG meta tags should override these.
				 */
			} else if prop == "twitter:title" && data.Title == "" {
				data.Title = cont
			} else if prop == "twitter:description" && data.Description == "" {
				data.Description = cont
			} else if prop == "twitter:image" && data.ImageSrc == "" {
				data.ImageSrc = formatThumbnailUrl(r, cont)
			}
		})
	}

	return data
}
