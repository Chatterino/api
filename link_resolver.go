package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	run   func(url *url.URL) ([]byte, error)
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

func doRequest(urlString string) (interface{}, error, time.Duration) {
	url, err := url.Parse(urlString)
	if err != nil {
		return rInvalidURL, nil, noSpecialDur
	}

	for _, m := range customURLManagers {
		if m.check(url) {
			data, err := m.run(url)
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

	escapedTitle := doc.Find("title").First().Text()
	if escapedTitle != "" {
		escapedTitle = fmt.Sprintf("<b>%s</b><hr>", clean(escapedTitle))
	}
	return marshalNoDur(&LinkResolverResponse{
		Status:  resp.StatusCode,
		Tooltip: fmt.Sprintf("<div style=\"text-align: left;\">%s<b>URL:</b> %s</div>", escapedTitle, clean(resp.Request.URL.String())),
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

	response := linkResolverCache.Get(url)

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
