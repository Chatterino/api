package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"log"
	"net/url"
	"regexp"
	"text/template"
)

const (
	timestampFormat = "Jan 2 2006 • 15:04 UTC"

	tweetTooltip = `<div style="text-align: left;">
<b>{{.Name}} (@{{.Username}})</b>
<br>
{{.Text}}
<br>
<span style="color: #808892;">{{.Likes}} likes&nbsp;•&nbsp;{{.Retweets}} retweets&nbsp;•&nbsp;{{.Timestamp}}</span>
</div>
`

	twitterUserTooltip = `<div style="text-align: left;">
<b>{{.Name}} (@{{.Username}})</b>
<br>
{{.Description}}
<br>
<span style="color: #808892;">{{.Followers}} followers</span>
</div>
`
)

var (
	tweetRegexp       = regexp.MustCompile(`(?i)\/.*\/status(?:es)?\/([^\/\?]+)`)
	twitterUserRegexp = regexp.MustCompile(`(?i)twitter\.com\/([^\/\?\s]+)(\/?$|(\?).*)`)
)

type TweetApiResponse struct {
	ID        string `json:"id_str"`
	Text      string `json:"full_text"`
	Timestamp string `json:"created_at"`
	Likes     uint64 `json:"favorite_count"`
	Retweets  uint64 `json:"retweet_count"`
	User      struct {
		Name            string `json:"name"`
		Username        string `json:"screen_name"`
		ProfileImageUrl string `json:"profile_image_url_https"`
	} `json:"user"`
	Entities struct {
		Media []struct {
			Url string `json:"media_url_https"`
		} `json:"media"`
	} `json:"entities"`
}

type tweetTooltipData struct {
	Text      string
	Name      string
	Username  string
	Timestamp string
	Likes     string
	Retweets  string
	Thumbnail string
}

type TwitterUserApiResponse struct {
	Name            string `json:"name"`
	Username        string `json:"screen_name"`
	Description     string `json:"description"`
	Followers       uint64 `json:"followers_count"`
	ProfileImageUrl string `json:"profile_image_url_https"`
}

type twitterUserTooltipData struct {
	Name        string
	Username    string
	Description string
	Followers   string
	Thumbnail   string
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

func buildTweetTooltip(tweet *TweetApiResponse) *tweetTooltipData {
	data := &tweetTooltipData{}
	data.Text = tweet.Text
	data.Name = tweet.User.Name
	data.Username = tweet.User.Username
	data.Likes = insertCommas(strconv.FormatUint(tweet.Likes, 10), 3)
	data.Retweets = insertCommas(strconv.FormatUint(tweet.Retweets, 10), 3)

	timestamp, err := time.Parse("Mon Jan 2 15:04:05 -0700 2006", tweet.Timestamp)
	data.Timestamp = timestamp.Format(timestampFormat)
	if err != nil {
		log.Println(err.Error())
		data.Timestamp = ""
	}

	if len(tweet.Entities.Media) > 0 {
		// If tweet contains an image, it will be used as thumbnail
		data.Thumbnail = tweet.Entities.Media[0].Url
	}

	return data
}

func getUserNameFromUrl(url *url.URL) string {
	match := twitterUserRegexp.FindAllStringSubmatch(url.String(), -1)
	if len(match) > 0 && len(match[0]) > 0 {
		return match[0][1]
	}
	return ""
}

func getUserByName(userName, bearer string) (*TwitterUserApiResponse, error) {
	endpointUrl := fmt.Sprintf("https://api.twitter.com/1.1/users/show.json?screen_name=%s", userName)
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

	var user *TwitterUserApiResponse
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, errors.New("unable to process response")
	}

	return user, nil
}

func buildTwitterUserTooltip(user *TwitterUserApiResponse) *twitterUserTooltipData {
	data := &twitterUserTooltipData{}
	data.Name = user.Name
	data.Username = user.Username
	data.Description = user.Description
	data.Followers = insertCommas(strconv.FormatUint(user.Followers, 10), 3)
	data.Thumbnail = user.ProfileImageUrl

	return data
}

func init() {
	bearerKey, exists := os.LookupEnv("CHATTERINO_API_TWITTER_BEARER_TOKEN")
	if !exists {
		log.Println("No CHATTERINO_API_TWITTER_BEARER_TOKEN specified, won't do special responses for twitter")
		return
	}

	tweetTooltipTemplate, err := template.New("tweetTooltip").Parse(tweetTooltip)
	if err != nil {
		log.Println("Error initializing Tweet tooltip template:", err)
		return
	}

	twitterUserTooltipTemplate, err := template.New("twitterUserTooltip").Parse(twitterUserTooltip)
	if err != nil {
		log.Println("Error initializing Tweet tooltip template:", err)
		return
	}

	loadTweet := func(tweetID string, r *http.Request) (interface{}, error, time.Duration) {
		log.Println("[Twitter] GET", tweetID)

		tweetResp, err := getTweetByID(tweetID, bearerKey)
		if err != nil {
			if err.Error() == "404" {
				var response LinkResolverResponse
				unmarshalErr := json.Unmarshal(rNoLinkInfoFound, &response)
				if unmarshalErr != nil {
					log.Println("Error unmarshalling prebuilt response:", unmarshalErr.Error())
				}

				return &response, nil, 1 * time.Hour
			}

			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "Error getting Tweet: " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		tweetData := buildTweetTooltip(tweetResp)
		var tooltip bytes.Buffer
		if err := tweetTooltipTemplate.Execute(&tooltip, tweetData); err != nil {
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

	loadTwitterUser := func(userName string, r *http.Request) (interface{}, error, time.Duration) {
		log.Println("[Twitter] GET", userName)

		userResp, err := getUserByName(userName, bearerKey)
		if err != nil {
			if err.Error() == "50" {
				var response LinkResolverResponse
				unmarshalErr := json.Unmarshal(rNoLinkInfoFound, &response)
				if unmarshalErr != nil {
					log.Println("Error unmarshalling prebuilt response:", unmarshalErr.Error())
				}

				return &response, nil, 1 * time.Hour
			}

			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "Error getting Twitter user: " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		userData := buildTwitterUserTooltip(userResp)
		var tooltip bytes.Buffer
		if err := twitterUserTooltipTemplate.Execute(&tooltip, userData); err != nil {
			return &LinkResolverResponse{
				Status:  http.StatusInternalServerError,
				Message: "Twitter user template error: " + clean(err.Error()),
			}, nil, noSpecialDur
		}

		return &LinkResolverResponse{
			Status:    http.StatusOK,
			Tooltip:   tooltip.String(),
			Thumbnail: userData.Thumbnail,
		}, nil, noSpecialDur
	}

	tweetCache := newLoadingCache("tweets", loadTweet, 24*time.Hour)
	twitterUserCache := newLoadingCache("twitterUsers", loadTwitterUser, 24*time.Hour)

	customURLManagers = append(customURLManagers, customURLManager{
		check: func(url *url.URL) bool {
			isTwitter := (strings.HasSuffix(url.Host, ".twitter.com") || url.Host == "twitter.com")

			if !isTwitter {
				return false
			}

			isTweet := tweetRegexp.MatchString(url.String())
			if isTweet {
				return true
			}

			isTwitterUser := twitterUserRegexp.MatchString(url.String())
			return isTwitterUser
		},
		run: func(url *url.URL) ([]byte, error) {
			if tweetRegexp.MatchString(url.String()) {
				tweetID := getTweetIDFromURL(url)
				if tweetID == "" {
					return rNoLinkInfoFound, nil
				}

				apiResponse := tweetCache.Get(tweetID, nil)
				return json.Marshal(apiResponse)
			}

			if twitterUserRegexp.MatchString(url.String()) {
				// We always use the lowercase representation in order
				// to avoid making redundant requests.
				userName := strings.ToLower(getUserNameFromUrl(url))
				if userName == "" {
					return rNoLinkInfoFound, nil
				}

				apiResponse := twitterUserCache.Get(userName, nil)
				return json.Marshal(apiResponse)
			}

			return rNoLinkInfoFound, nil
		},
	})
}
