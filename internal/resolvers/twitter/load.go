package twitter

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
)

type TweetLoader struct {
	bearerKey string
}

func (l *TweetLoader) Load(ctx context.Context, tweetID string, r *http.Request) (*resolver.Response, time.Duration, error) {
	log := logger.FromContext(ctx)

	log.Debugw("[Twitter] Get tweet",
		"tweetID", tweetID,
	)

	tweetResp, err := getTweetByID(tweetID, l.bearerKey)
	if err != nil {
		if err.Error() == "404" {
			var response resolver.Response
			unmarshalErr := json.Unmarshal(resolver.NoLinkInfoFound, &response)
			if unmarshalErr != nil {
				log.Errorw("Error unmarshalling prebuilt response",
					"error", unmarshalErr.Error(),
				)
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

type UserLoader struct {
	bearerKey string
}

func (l *UserLoader) Load(ctx context.Context, userName string, r *http.Request) (*resolver.Response, time.Duration, error) {
	log := logger.FromContext(ctx)

	log.Debugw("[Twitter] Get user",
		"userName", userName,
	)

	userResp, err := getUserByName(userName, l.bearerKey)
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
