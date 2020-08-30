package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"log"
	"net/url"
	"regexp"
	"text/template"
)

const (
	timestampFormat = "Jan 2 2006 • 15:04 UTC"

	tweeterTooltip = `<div style="text-align: left;">
<b>{{.Name}} (@{{.Username}})</b>
<br>
{{.Text}}
<br>
<span style="color: #808892;">{{.Timestamp}}</span>
</div>
`
)

var (
	tweetRegexp = regexp.MustCompile(`(?i)\/.*\/status(?:es)?\/([^\/\?]+)`)
)

type TweetApiResponse struct {
	ID        string `json:"id_str"`
	Text      string `json:"full_text"`
	Timestamp string `json:"created_at"`
	User      struct {
		Name     string `json:"name"`
		Username string `json:"screen_name"`
	} `json:"user"`
	Entities struct {
		Media []struct {
			URL string `json:"media_url_https"`
		} `json:"media"`
	} `json:"entities"`
}

type tweetTooltipData struct {
	Text      string
	Name      string
	Username  string
	Timestamp string
	Thumbnail string
}

func getTweetIDFromURL(url *url.URL) string {
	match := tweetRegexp.FindAllStringSubmatch(url.Path, -1)
	if len(match) > 0 && len(match[0]) == 2 {
		return match[0][1]
	}
	return ""
}

func getTweetByID(id, bearer string) (*TweetApiResponse, error) {
	endpointUrl := fmt.Sprintf("https://api.twitter.com/1.1/statuses/show.json?id=%s&tweet_mode=extended", id)
	req, err := http.NewRequest("GET", endpointUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearer))
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%d", resp.StatusCode)
	}

	var tweet *TweetApiResponse
	err = json.NewDecoder(resp.Body).Decode(&tweet)
	if err != nil {
		return nil, errors.New("unable to process response")
	}

	return tweet, nil
}

func buildTooltip(tweet *TweetApiResponse) *tweetTooltipData {
	data := &tweetTooltipData{}
	data.Text = tweet.Text
	data.Name = tweet.User.Name
	data.Username = tweet.User.Username

	timestamp, err := time.Parse("Mon Jan 2 15:04:05 -0700 2006", tweet.Timestamp)
	data.Timestamp = timestamp.Format(timestampFormat)
	if err != nil {
		log.Println(err.Error())
		data.Timestamp = ""
	}

	if len(tweet.Entities.Media) > 0 {
		data.Thumbnail = tweet.Entities.Media[0].URL
	}

	return data
}

func init() {
	bearerKey, exists := os.LookupEnv("CHATTERINO_API_TWITTER_BEARER_TOKEN")
	if !exists {
		log.Println("No CHATTERINO_API_TWITTER_BEARER_TOKEN specified, won't do special responses for twitter")
		return
	}

	tooltipTemplate, err := template.New("tweetTooltip").Parse(tweeterTooltip)
	if err != nil {
		log.Println("Error initializing Tweet tooltip template:", err)
		return
	}

	load := func(tweetID string, r *http.Request) (interface{}, error, time.Duration) {
		log.Println("[Twitter] GET", tweetID)

		tweetResp, err := getTweetByID(tweetID, bearerKey)
		if err != nil {
			if err.Error() == "404" {
				var response LinkResolverResponse
				json.Unmarshal(rNoLinkInfoFound, &response)

				return &response, nil, 1 * time.Hour
			}

			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "Error getting Tweet: " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		tweetData := buildTooltip(tweetResp)
		var tooltip bytes.Buffer
		if err := tooltipTemplate.Execute(&tooltip, tweetData); err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "Tweet template error: " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		return &LinkResolverResponse{
			Status:    http.StatusOK,
			Tooltip:   tooltip.String(),
			Thumbnail: tweetData.Thumbnail,
		}, nil, noSpecialDur
	}

	cache := newLoadingCache("twitter", load, 24*time.Hour)

	customURLManagers = append(customURLManagers, customURLManager{
		check: func(url *url.URL) bool {
			// Additionally checking the regex to provide default link resolver response on non-status links
			return (strings.HasSuffix(url.Host, ".twitter.com") || url.Host == "twitter.com") &&
				tweetRegexp.MatchString(url.String())
		},
		run: func(url *url.URL) ([]byte, error) {
			tweetID := getTweetIDFromURL(url)
			if tweetID == "" {
				return rNoLinkInfoFound, nil
			}

			apiResponse := cache.Get(tweetID, nil)
			return json.Marshal(apiResponse)
		},
	})
}
