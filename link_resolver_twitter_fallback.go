package main

/* When there's no API key set for Twitter, this fallback should be used.
 */

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	errInvalidTweetID = errors.New("invalid tweet id")
)

func init() {
	_, exists := os.LookupEnv("TWITTER_BEARER_KEY")
	if exists {
		log.Println("TWITTER_BEARER_KEY specified, won't do fallback parsing for Twitter")
		return
	}

	const (
		mobileTwitterURL = "https://mobile.twitter.com/status/status/%s"

		tooltipTemplate = `<div style="text-align: left;">
<b>Twitter - {{.Author}}</b><hr>
{{if .Tweet}}
<span style="white-space: pre-wrap;word-wrap: break-word;">{{.Tweet}}</span><hr>
{{end}}
<b>Date:</b> {{.Date}}</div>`
	)

	var (
		tweetNotFoundResponse = &LinkResolverResponse{
			Status:  http.StatusNotFound,
			Message: "No Tweet with this id found",
		}
	)

	type TooltipData struct {
		Author string
		Tweet  string
		Date   string
	}

	tmpl, err := template.New("twitterFallbackTooltip").Parse(tooltipTemplate)
	if err != nil {
		log.Println("Error initialization twitterFallbackTooltip tooltip template:", err)
		return
	}

	load := func(twitterSnowflake string, r *http.Request) (interface{}, time.Duration, error) {
		pageURL := fmt.Sprintf(mobileTwitterURL, twitterSnowflake)

		resp, err := makeRequest(pageURL)
		if err != nil {
			if strings.HasSuffix(err.Error(), "no such host") {
				return rNoLinkInfoFound, noSpecialDur, nil
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
				return nil, noSpecialDur, err
			}
			if contentLengthBytes > maxContentLength {
				return rResponseTooLarge, noSpecialDur, nil
			}
		}

		if resp.StatusCode == http.StatusNotFound {
			return tweetNotFoundResponse, noSpecialDur, nil
		}

		if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusMultipleChoices {
			fmt.Println("Skipping twitter fallback url", resp.Request.URL, "because status code is", resp.StatusCode)
			return rNoLinkInfoFound, noSpecialDur, nil
		}

		limiter := &WriteLimiter{Limit: maxContentLength}

		doc, err := goquery.NewDocumentFromReader(io.TeeReader(resp.Body, limiter))
		if err != nil {
			return marshalNoDur(&LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "twitter fallback html parser error " + clean(err.Error()),
			})
		}

		// Build tooltip data from the API response
		data := TooltipData{
			Author: strings.TrimSpace(doc.Find(".main-tweet .user-info .fullname").Text()),
			Tweet:  strings.TrimSpace(doc.Find(".main-tweet .tweet-content .tweet-text").Text()),
			Date:   strings.TrimSpace(doc.Find(".main-tweet .tweet-content .metadata").Text()),
		}

		// Build a tooltip using the tooltip template (see tooltipTemplate) with the data we massaged above
		var tooltip bytes.Buffer
		if err := tmpl.Execute(&tooltip, data); err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "twitter fallback template error " + clean(err.Error()),
			}, noSpecialDur, nil
		}

		thumbnailURL, _ := doc.Find(".main-tweet .tweet-content .card-photo .media img").Attr("src")

		return &LinkResolverResponse{
			Status:    200,
			Tooltip:   tooltip.String(),
			Thumbnail: thumbnailURL,
		}, noSpecialDur, nil
	}

	cache := newLoadingCache("twitter_fallback", load, 1*time.Hour)
	tweetPathRegex := regexp.MustCompile(`/[\w_]{1,15}/status/([0-9]+)`)

	twitterDomains := map[string]struct{}{
		"twitter.com":        {},
		"mobile.twitter.com": {},
	}

	customURLManagers = append(customURLManagers, customURLManager{
		check: func(url *url.URL) bool {
			host := strings.ToLower(url.Host)

			if _, ok := twitterDomains[host]; !ok {
				return false
			}

			if !tweetPathRegex.MatchString(url.Path) {
				return false
			}

			return true
		},
		run: func(url *url.URL) ([]byte, error) {
			matches := tweetPathRegex.FindStringSubmatch(url.Path)
			if len(matches) != 2 {
				return nil, errInvalidTweetID
			}

			tweetSnowflake := matches[1]

			response := cache.Get(tweetSnowflake, nil)
			return json.Marshal(response)
		},
	})
}
