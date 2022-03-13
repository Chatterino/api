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
