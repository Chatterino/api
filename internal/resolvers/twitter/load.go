package twitter

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
)

func loadTweet(tweetID string, r *http.Request) (interface{}, time.Duration, error) {
	log.Println("[Twitter] GET", tweetID)

	tweetResp, err := getTweetByID(tweetID, bearerKey)
	if err != nil {
		if err.Error() == "404" {
			var response resolver.Response
			unmarshalErr := json.Unmarshal(resolver.NoLinkInfoFound, &response)
			if unmarshalErr != nil {
				log.Println("Error unmarshalling prebuilt response:", unmarshalErr.Error())
			}

			return &response, 1 * time.Hour, nil
		}

		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "Error getting Tweet: " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	tweetData := buildTweetTooltip(tweetResp)
	var tooltip bytes.Buffer
	if err := tweetTooltipTemplate.Execute(&tooltip, tweetData); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "Tweet template error: " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	return &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: tweetData.Thumbnail,
	}, cache.NoSpecialDur, nil
}

func loadTwitterUser(userName string, r *http.Request) (interface{}, time.Duration, error) {
	log.Println("[Twitter] GET", userName)

	userResp, err := getUserByName(userName, bearerKey)
	if err != nil {
		// Error code for "User not found.", as described here:
		// https://developer.twitter.com/en/support/twitter-api/error-troubleshooting#error-codes
		if err.Error() == "50" {
			return &resolver.Response{
				Status:  http.StatusNotFound,
				Message: "Error: Twitter user not found.",
			}, cache.NoSpecialDur, nil
		}

		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "Error getting Twitter user: " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	userData := buildTwitterUserTooltip(userResp)
	var tooltip bytes.Buffer
	if err := twitterUserTooltipTemplate.Execute(&tooltip, userData); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "Twitter user template error: " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	return &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: userData.Thumbnail,
	}, cache.NoSpecialDur, nil
}
