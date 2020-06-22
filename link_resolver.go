package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
	"github.com/samclarke/robotstxt"
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

func doRequest(urlString string, r *http.Request) (interface{}, error, time.Duration) {
	isAllowed, err := isAllowedByRobotsTxt(urlString, r)
	if err != nil {
		log.Println("Error checking robots.txt", err)
	} else if !isAllowed {
		return rForbiddenByRobotsTxt, nil, noSpecialDur
	}

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

	var tooltip bytes.Buffer
	if err := tooltipTemplate.Execute(&tooltip, data); err != nil {
		return marshalNoDur(&LinkResolverResponse{
			Status:  http.StatusInternalServerError,
			Message: "template error " + clean(err.Error()),
		})
	}

	response := &LinkResolverResponse{
		Status:  resp.StatusCode,
		Tooltip: tooltip.String(),
		Link:    resp.Request.URL.String(),
	}

	if isSupportedThumbnail(resp.Header.Get("content-type")) {
		if *baseURL == "" {
			scheme := "https://"
			if r.TLS == nil {
				scheme = "http://" // https://github.com/golang/go/issues/28940#issuecomment-441749380
			}
			response.Thumbnail = fmt.Sprintf("%s%s/thumbnail/%s", scheme, r.Host, url.QueryEscape(resp.Request.URL.String()))
		} else {
			response.Thumbnail = fmt.Sprintf("%s/thumbnail/%s", strings.TrimSuffix(*baseURL, "/"), url.QueryEscape(resp.Request.URL.String()))
		}
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

func isAllowedByRobotsTxt(urlString string, r *http.Request) (bool, error) {
	robotsTxtUrl, err := getRobotsTxtUrl(urlString)
	if err != nil {
		return false, err
	}
	robotsData := robotsTxtCache.Get(robotsTxtUrl, r)
	if robotsData == nil {
		return false, errors.New("no robots.txt data")
	}
	robots := robotsData.(*robotstxt.RobotsTxt)

	isRouteAllowed, err := robots.IsAllowed("chatterino-api-cache/1.0 link-resolver", urlString)
	if err != nil {
		return false, err
	}
	return isRouteAllowed, nil
}

func getRobotsTxtUrl(urlString string) (string, error) {
	requestUrl, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}

	requestUrl.Path = "robots.txt"
	requestUrl.Fragment = ""
	requestUrl.RawQuery = ""

	return requestUrl.String(), nil
}

func doRobotsTxtRequest(urlString string, r *http.Request) (interface{}, error, time.Duration) {
	resp, err := makeRequest(urlString)
	if err != nil {
		return nil, err, noSpecialDur
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK && resp.StatusCode >= http.StatusMultipleChoices {
		return nil, errors.New("failed to get robots.txt data"), noSpecialDur
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, noSpecialDur
	}

	robots, err := robotstxt.Parse(string(body), urlString)
	if err != nil {
		return nil, nil, noSpecialDur
	}

	return robots, nil, noSpecialDur
}

var linkResolverCache *loadingCache
var robotsTxtCache *loadingCache

func init() {
	linkResolverCache = newLoadingCache("linkResolver", doRequest, 10*time.Minute)
	robotsTxtCache = newLoadingCache("robotsTxt", doRobotsTxtRequest, 10*time.Minute)
}

func handleLinkResolver(router *mux.Router) {
	router.HandleFunc("/link_resolver/{url:.*}", linkResolver).Methods("GET")
}
