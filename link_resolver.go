package main

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
	"sync"
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

var linkResolverRequestsMutex sync.Mutex
var linkResolverRequests = make(map[string][](chan interface{}))

type customURLManager struct {
	check func(resp *http.Response) bool
	run   func(resp *http.Response) ([]byte, error)
}

var (
	customURLManagers []customURLManager
)

func doRequest(url string) {
	response := cacheGetOrSet("url:"+url, 10*time.Minute, func() (interface{}, error) {
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

			return json.Marshal(&LinkResolverResponse{Status: 500, Message: "client.Get " + err.Error()})
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				return json.Marshal(&LinkResolverResponse{Status: 500, Message: "html parser error " + err.Error()})
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
	})

	linkResolverRequestsMutex.Lock()
	fmt.Println("Notify channels")
	for _, channel := range linkResolverRequests[url] {
		fmt.Printf("Notify channel %v\n", channel)
		/*
			select {
			case channel <- response:
				fmt.Println("hehe")
			default:
				fmt.Println("Unable to respond")
			}
		*/
		channel <- response
	}
	delete(linkResolverRequests, url)
	linkResolverRequestsMutex.Unlock()
}

func linkResolver(w http.ResponseWriter, r *http.Request) {
	url, err := unescapeURLArgument(r, "url")
	if err != nil {
		_, err = w.Write(rInvalidURL)
		if err != nil {
			fmt.Println("Error in w.Write:", err)
		}
		return
	}

	cacheKey := "url:" + url

	var response interface{}

	if data := cacheGet(cacheKey); data != nil {
		response = data
	} else {
		responseChannel := make(chan interface{})

		linkResolverRequestsMutex.Lock()
		linkResolverRequests[url] = append(linkResolverRequests[url], responseChannel)
		urlRequestsLength := len(linkResolverRequests[url])
		linkResolverRequestsMutex.Unlock()
		if urlRequestsLength == 1 {
			// First poll for this URL, start the request!
			go doRequest(url)
		}

		fmt.Printf("Listening to channel %v\n", responseChannel)
		response = <-responseChannel
		fmt.Println("got response!")
	}

	_, err = w.Write(response.([]byte))
	if err != nil {
		fmt.Println("Error in w.Write:", err)
	}
}
