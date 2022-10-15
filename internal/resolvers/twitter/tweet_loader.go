package twitter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
)

type APIUser struct {
	Name            string `json:"name"`
	Username        string `json:"screen_name"`
	ProfileImageUrl string `json:"profile_image_url_https"`
}

type APIEntitiesMedia struct {
	Url string `json:"media_url_https"`
}

type APIEntities struct {
	Media []APIEntitiesMedia `json:"media"`
}

type TweetApiResponse struct {
	ID        string      `json:"id_str"`
	Text      string      `json:"full_text"`
	Timestamp string      `json:"created_at"`
	Likes     uint64      `json:"favorite_count"`
	Retweets  uint64      `json:"retweet_count"`
	User      APIUser     `json:"user"`
	Entities  APIEntities `json:"entities"`
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
	bearerKey             string
	endpointURLFormat     string
	tweetCacheKeyProvider cache.KeyProvider
}

var (
	errTweetNotFound = errors.New("tweet not found")
)

func NewTweetLoader(
	bearerKey string,
	endpointURLFormat string,
	tweetCacheKeyProvider cache.KeyProvider,
) *TweetLoader {
	return &TweetLoader{
		bearerKey:             bearerKey,
		endpointURLFormat:     endpointURLFormat,
		tweetCacheKeyProvider: tweetCacheKeyProvider,
	}
}

func (l *TweetLoader) getTweetByID(id string) (*TweetApiResponse, error) {
	endpointUrl := fmt.Sprintf(l.endpointURLFormat, id)
	extraHeaders := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", l.bearerKey),
	}
	resp, err := resolver.RequestGETWithHeaders(endpointUrl, extraHeaders)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errTweetNotFound
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unhandled status code: %d", resp.StatusCode)
	}

	var tweet *TweetApiResponse
	err = json.NewDecoder(resp.Body).Decode(&tweet)
	if err != nil {
		return nil, errors.New("unable to unmarshal response")
	}

	return tweet, nil
}

func (l *TweetLoader) Load(ctx context.Context, tweetID string, r *http.Request) (*resolver.Response, time.Duration, error) {
	log := logger.FromContext(ctx)

	log.Debugw("[Twitter] Get tweet",
		"tweetID", tweetID,
	)

	tweetResp, err := l.getTweetByID(tweetID)
	if err != nil {
		if err == errTweetNotFound {
			return &resolver.Response{
				Status:  http.StatusNotFound,
				Message: fmt.Sprintf("Twitter tweet not found: %s", resolver.CleanResponse(tweetID)),
			}, cache.NoSpecialDur, nil
		}

		return resolver.Errorf("Twitter tweet API error: %s", err)
	}

	tweetData := buildTweetTooltip(tweetResp)
	var tooltip bytes.Buffer
	if err := tweetTooltipTemplate.Execute(&tooltip, tweetData); err != nil {
		return resolver.Errorf("Twitter tweet template error: %s", err)
	}

	return &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: tweetData.Thumbnail,
	}, cache.NoSpecialDur, nil
}
