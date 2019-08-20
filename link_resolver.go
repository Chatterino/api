package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
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
	check func(resp *http.Response) bool
	run   func(resp *http.Response) ([]byte, error)
}

var (
	customURLManagers []customURLManager
)

func makeRequest(url string) (response *http.Response, err error) {
	resp, err := httpClient.Head(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		contentLengthBytes, err := strconv.Atoi(contentLength)
		if err != nil {
			return nil, err
		}
		if contentLengthBytes > maxContentLength {
			return nil, errors.New("too big")
		}
	}

	if resp.Request == nil {
		return nil, errors.New("bad response, no request")
	}

	req, err := http.NewRequest("GET", resp.Request.URL.String(), nil)
	if err != nil {
		return nil, err
	}

	// ensures websites return pages in english (e.g. twitter would return french preview
	// when the request came from a french IP.)
	req.Header.Add("Accept-Language", "en-US, en;q=0.9, *;q=0.5")
	req.Header.Set("User-Agent", "chatterino-api-cache/1.0 link-resolver")

	return getClient.Do(req)
}

func doRequest(url string) (interface{}, error) {
	resp, err := makeRequest(url)
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such host") {
			return rNoLinkInfoFound, nil
		}

		return json.Marshal(&LinkResolverResponse{
			Status:  http.StatusInternalServerError,
			Message: "client.Get " + err.Error(),
		})
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
		return rNoLinkInfoFound, nil
	}

	limiter := &WriteLimiter{Limit: maxContentLength}

	doc, err := goquery.NewDocumentFromReader(io.TeeReader(resp.Body, limiter))
	if err != nil {
		return json.Marshal(&LinkResolverResponse{
			Status:  http.StatusInternalServerError,
			Message: "html parser error (or download) " + err.Error(),
		})
	}

	for _, m := range customURLManagers {
		if m.check(resp) {
			return m.run(resp)
		}
	}

	escapedTitle := doc.Find("title").First().Text()
	if escapedTitle != "" {
		escapedTitle = fmt.Sprintf("<b>%s</b><hr>", html.EscapeString(escapedTitle))
	}
	return json.Marshal(&LinkResolverResponse{
		Status:  resp.StatusCode,
		Tooltip: fmt.Sprintf("<div style=\"text-align: left;\">%s<b>URL:</b> %s</div>", escapedTitle, html.EscapeString(resp.Request.URL.String())),
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
