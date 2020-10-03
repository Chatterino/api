package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

type LinkResolverResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`

	Thumbnail string `json:"thumbnail,omitempty"`
	Tooltip   string `json:"tooltip,omitempty"`
	Link      string `json:"link,omitempty"`

	// Flag in the BTTV API to.. maybe signify that the link will download something? idk
	// Download *bool  `json:"download,omitempty"`
}

type customURLManager struct {
	check func(url *url.URL) bool
	run   func(url *url.URL) ([]byte, error)
}

const tooltip = `<div style="text-align: left;">
{{if .Title}}
<b>{{.Title}}</b><hr>
{{end}}
{{if .Description}}
<span>{{.Description}}</span><hr>
{{end}}
<b>URL:</b> {{.URL}}</div>`

type tooltipData struct {
	URL         string
	Title       string
	Description string
	ImageSrc    string
}

func (d *tooltipData) Truncate() {
	d.Title = truncateString(d.Title, MaxTitleLength)
	d.Description = truncateString(d.Description, MaxDescriptionLength)
}

var (
	customURLManagers []customURLManager
)

func makeRequest(url string) (response *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// ensures websites return pages in english (e.g. twitter would return french preview
	// when the request came from a french IP.)
	req.Header.Add("Accept-Language", "en-US, en;q=0.9, *;q=0.5")
	req.Header.Set("User-Agent", "chatterino-api-cache/1.0 link-resolver")

	return httpClient.Do(req)
}

func defaultTooltipData(doc *goquery.Document, r *http.Request, resp *http.Response) tooltipData {
	data := tooltipData{
		URL: clean(resp.Request.URL.String()),
	}

	/* Support for HTML Open Graph meta tags.
	 * Will show Open Graph "Title", "Description", "Image" information of webpages.
	 * More fields are available: https://ogp.me/
	 */
	metaFields := doc.Find("meta[property][content]")
	if metaFields.Size() > 0 {
		metaFields.Each(func(i int, s *goquery.Selection) {
			prop, _ := s.Attr("property")
			cont, _ := s.Attr("content")
			if prop == "og:title" {
				data.Title = cont
			} else if prop == "og:description" {
				data.Description = cont
			} else if prop == "og:image" {
				data.ImageSrc = formatThumbnailUrl(r, cont)
			}
		})
	}

	if data.Title == "" {
		data.Title = doc.Find("title").First().Text()
	}

	return data
}

func formatThumbnailUrl(r *http.Request, urlString string) string {
	if *baseURL == "" {
		scheme := "https://"
		if r.TLS == nil {
			scheme = "http://" // https://github.com/golang/go/issues/28940#issuecomment-441749380
		}
		return fmt.Sprintf("%s%s/thumbnail/%s", scheme, r.Host, url.QueryEscape(urlString))
	}
	return fmt.Sprintf("%s/thumbnail/%s", strings.TrimSuffix(*baseURL, "/"), url.QueryEscape(urlString))
}

func doRequest(urlString string, r *http.Request) (interface{}, error, time.Duration) {
	requestUrl, err := url.Parse(urlString)
	if err != nil {
		return rInvalidURL, nil, noSpecialDur
	}

	for _, m := range customURLManagers {
		if m.check(requestUrl) {
			data, err := m.run(requestUrl)
			return data, err, noSpecialDur
		}
	}

	resp, err := makeRequest(requestUrl.String())
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such host") {
			return rNoLinkInfoFound, nil, noSpecialDur
		}

		return marshalNoDur(&LinkResolverResponse{
			Status:  http.StatusInternalServerError,
			Message: clean(err.Error()),
		})
	}

	defer resp.Body.Close()

	// If the initial request URL is different from the response's apparent request URL,
	// we likely followed a redirect. Re-check the custom URL managers to see if the
	// page we were redirected to supports rich content. If not, continue with the
	// default tooltip.
	if requestUrl.String() != resp.Request.URL.String() {
		for _, m := range customURLManagers {
			if m.check(resp.Request.URL) {
				data, err := m.run(resp.Request.URL)
				return data, err, noSpecialDur
			}
		}
	}

	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		contentLengthBytes, err := strconv.Atoi(contentLength)
		if err != nil {
			return nil, err, noSpecialDur
		}
		if contentLengthBytes > maxContentLength {
			return rResponseTooLarge, nil, noSpecialDur
		}
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		fmt.Println("Skipping url", resp.Request.URL, "because status code is", resp.StatusCode)
		return rNoLinkInfoFound, nil, noSpecialDur
	}

	limiter := &WriteLimiter{Limit: maxContentLength}

	doc, err := goquery.NewDocumentFromReader(io.TeeReader(resp.Body, limiter))
	if err != nil {
		return marshalNoDur(&LinkResolverResponse{
			Status:  http.StatusInternalServerError,
			Message: "html parser error (or download) " + clean(err.Error()),
		})
	}

	tooltipTemplate, err := template.New("tooltip").Parse(tooltip)
	if err != nil {
		log.Println("Error initialization tooltip template:", err)
		return marshalNoDur(&LinkResolverResponse{
			Status:  http.StatusInternalServerError,
			Message: "template error " + clean(err.Error()),
		})
	}

	data := defaultTooltipData(doc, r, resp)

	// Truncate title and description in case they're too long
	data.Truncate()

	var tooltip bytes.Buffer
	if err := tooltipTemplate.Execute(&tooltip, data); err != nil {
		return marshalNoDur(&LinkResolverResponse{
			Status:  http.StatusInternalServerError,
			Message: "template error " + clean(err.Error()),
		})
	}

	response := &LinkResolverResponse{
		Status:    resp.StatusCode,
		Tooltip:   url.PathEscape(tooltip.String()),
		Link:      resp.Request.URL.String(),
		Thumbnail: data.ImageSrc,
	}

	if isSupportedThumbnail(resp.Header.Get("content-type")) {
		response.Thumbnail = formatThumbnailUrl(r, resp.Request.URL.String())
	}

	return marshalNoDur(response)
}

func linkResolver(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	url, err := unescapeURLArgument(r, "url")
	if err != nil {
		_, err = w.Write(rInvalidURL)
		if err != nil {
			log.Println("Error writing response:", err)
		}
		return
	}

	response := linkResolverCache.Get(url, r)

	_, err = w.Write(response.([]byte))
	if err != nil {
		log.Println("Error writing response:", err)
	}
}

var linkResolverCache *loadingCache

func init() {
	linkResolverCache = newLoadingCache("linkResolver", doRequest, 10*time.Minute)
}

func handleLinkResolver(router *mux.Router) {
	router.HandleFunc("/link_resolver/{url:.*}", linkResolver).Methods("GET")
}
