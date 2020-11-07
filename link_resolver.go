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

	"github.com/Chatterino/api/pkg/cache"
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"

	"github.com/Chatterino/api/internal/resolvers/betterttv"
	"github.com/Chatterino/api/pkg/resolver"
)

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
	customURLManagers []resolver.CustomURLManager
)

func makeRequest(url string) (response *http.Response, err error) {
	return resolver.RequestGET(url)
}

func defaultTooltipData(doc *goquery.Document, r *http.Request, resp *http.Response) tooltipData {
	data := tooltipMetaFields(doc, r, resp, tooltipData{
		URL: resolver.CleanResponse(resp.Request.URL.String()),
	})

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
		if m.Check(requestUrl) {
			data, err := m.Run(requestUrl)
			return data, err, noSpecialDur
		}
	}

	resp, err := makeRequest(requestUrl.String())
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such host") {
			return rNoLinkInfoFound, nil, noSpecialDur
		}

		return marshalNoDur(&resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: resolver.CleanResponse(err.Error()),
		})
	}

	defer resp.Body.Close()

	// If the initial request URL is different from the response's apparent request URL,
	// we likely followed a redirect. Re-check the custom URL managers to see if the
	// page we were redirected to supports rich content. If not, continue with the
	// default tooltip.
	if requestUrl.String() != resp.Request.URL.String() {
		for _, m := range customURLManagers {
			if m.Check(resp.Request.URL) {
				data, err := m.Run(resp.Request.URL)
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
		return marshalNoDur(&resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "html parser error (or download) " + resolver.CleanResponse(err.Error()),
		})
	}

	tooltipTemplate, err := template.New("tooltip").Parse(tooltip)
	if err != nil {
		log.Println("Error initialization tooltip template:", err)
		return marshalNoDur(&resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "template error " + resolver.CleanResponse(err.Error()),
		})
	}

	data := defaultTooltipData(doc, r, resp)

	// Truncate title and description in case they're too long
	data.Truncate()

	var tooltip bytes.Buffer
	if err := tooltipTemplate.Execute(&tooltip, data); err != nil {
		return marshalNoDur(&resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "template error " + resolver.CleanResponse(err.Error()),
		})
	}

	response := &resolver.Response{
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

var linkResolverCache = cache.New("linkResolver", doRequest, time.Duration(10)*time.Minute)

func register(managers []resolver.CustomURLManager) {
	customURLManagers = append(customURLManagers, managers...)
}

func init() {
	// Register Link Resolvers from internal/resolvers/
	register(betterttv.New())
}

func handleLinkResolver(router *mux.Router) {
	router.HandleFunc("/link_resolver/{url:.*}", linkResolver).Methods("GET")
}
