package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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

func doRequest(url string) (interface{}, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// ensures websites return pages in english (e.g. twitter would return french preview
	// when the request came from a french IP.)
	req.Header.Add("Accept-Language", "en-US, en;q=0.9, *;q=0.5")

	resp, err := httpClient.Do(req)
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

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return json.Marshal(&LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "html parser error " + err.Error(),
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

	return rNoLinkInfoFound, nil
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
