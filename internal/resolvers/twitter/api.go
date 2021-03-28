package twitter

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
)

func getTweetIDFromURL(url *url.URL) string {
	match := tweetRegexp.FindAllStringSubmatch(url.Path, -1)
	if len(match) > 0 && len(match[0]) == 2 {
		return match[0][1]
	}
	return ""
}

func buildTweetTooltip(tweet *TweetApiResponse) *tweetTooltipData {
	data := &tweetTooltipData{}
	data.Text = tweet.Text
	data.Name = tweet.User.Name
	data.Username = tweet.User.Username
	data.Likes = humanize.Number(tweet.Likes)
	data.Retweets = humanize.Number(tweet.Retweets)

	// TODO: what time format is this exactly? can we move to humanize a la CreationDteRFC3339?
	timestamp, err := time.Parse("Mon Jan 2 15:04:05 -0700 2006", tweet.Timestamp)
	data.Timestamp = humanize.CreationDateTime(timestamp)
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

func buildTwitterUserTooltip(user *TwitterUserApiResponse) *twitterUserTooltipData {
	data := &twitterUserTooltipData{}
	data.Name = user.Name
	data.Username = user.Username
	data.Description = user.Description
	data.Followers = humanize.Number(user.Followers)
	data.Thumbnail = user.ProfileImageUrl

	return data
}

func getTweetByID(id, bearer string) (*TweetApiResponse, error) {
	endpointUrl := fmt.Sprintf("https://api.twitter.com/1.1/statuses/show.json?id=%s&tweet_mode=extended", id)
	extraHeaders := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", bearer),
	}
	resp, err := resolver.RequestGETWithHeaders(endpointUrl, extraHeaders)
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

func getUserByName(userName, bearer string) (*TwitterUserApiResponse, error) {
	endpointUrl := fmt.Sprintf("https://api.twitter.com/1.1/users/show.json?screen_name=%s", userName)
	extraHeaders := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", bearer),
	}
	resp, err := resolver.RequestGETWithHeaders(endpointUrl, extraHeaders)
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
