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

	Tooltip string `json:"tooltip,omitempty"`
	Link    string `json:"link,omitempty"`

	// Flag in the BTTV API to.. maybe signify that the link will download something? idk
	// Download *bool  `json:"download,omitempty"`
}

type customURLManager struct {
	check func(url *url.URL) bool
	run   func(url *url.URL, options requestOptions) ([]byte, error)
}

type requestOptions struct {
	richTooltip bool
}

const tooltip = `<div style="text-align: left;">
{{if .ImageSrc}}
<img src="{{.ImageSrc}}"><br>
{{end}}
{{if .Title}}
<b>{{.Title}}</b><hr>
{{end}}
<b>URL:</b> {{.URL}}</div>`

type tooltipData struct {
	URL      string
	Title    string
	ImageSrc string
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

func doRequest(urlString string, options requestOptions) (interface{}, error, time.Duration) {
	url, err := url.Parse(urlString)
	if err != nil {
		return rInvalidURL, nil, noSpecialDur
	}

	for _, m := range customURLManagers {
		if m.check(url) {
			data, err := m.run(url, options)
			return data, err, noSpecialDur
		}
	}

	resp, err := makeRequest(url.String())
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

	data := tooltipData{
		URL:   clean(resp.Request.URL.String()),
		Title: doc.Find("title").First().Text(),
	}

	if options.richTooltip && strings.HasPrefix(resp.Header.Get("content-type"), "image/") {
		data.ImageSrc = resp.Request.URL.String()
	}

	var tooltip bytes.Buffer
	if err := tooltipTemplate.Execute(&tooltip, data); err != nil {
		return marshalNoDur(&LinkResolverResponse{
			Status:  http.StatusInternalServerError,
			Message: "template error " + clean(err.Error()),
		})
	}

	return marshalNoDur(&LinkResolverResponse{
		Status:  resp.StatusCode,
		Tooltip: tooltip.String(),
		Link:    resp.Request.URL.String(),
	})
}

func linkResolver(w http.ResponseWriter, r *http.Request) {
	url, err := unescapeURLArgument(r, "url")
	if err != nil {
		_, err = w.Write(rInvalidURL)
		if err != nil {
			log.Println("Error writing response:", err)
		}
		return
	}

	_, rich := r.URL.Query()["rich"]

	response := linkResolverCache.Get(url, requestOptions{richTooltip: rich})

	_, err = w.Write(response.([]byte))
	if err != nil {
		log.Println("Error writing response:", err)
	}
}

var linkResolverCache *loadingCache

func init() {
	linkResolverCache = newLoadingCache("url", doRequest, 10*time.Minute)
}

func handleLinkResolver(router *mux.Router) {
	router.HandleFunc("/link_resolver/{url:.*}", linkResolver).Methods("GET")
}
