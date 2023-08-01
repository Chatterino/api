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
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

type EmbedApiResponse struct {
	Text              string             `json:"text"`
	ID                string             `json:"id_str"`
	CreatedAt         time.Time          `json:"created_at"`
	User              EmbedUser          `json:"user"`
	FavoriteCount     uint64             `json:"favorite_count"`
	ConversationCount uint64             `json:"conversation_count"`
	MediaDetails      []EmbedMediaDetail `json:"mediaDetails"`
}

type EmbedUser struct {
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
}

type EmbedMediaDetail struct {
	MediaUrl string `json:"media_url_https"`
}

func (e EmbedMediaDetail) Url() string {
	return e.MediaUrl
}

type EmbedLoader struct {
	baseURL               string
	endpointURLFormat     string
	tweetCacheKeyProvider cache.KeyProvider
	collageCache          cache.DependentCache
	maxThumbnailSize      uint
}

type embedTweetTooltipData struct {
	Text      string
	Name      string
	Username  string
	Timestamp string
	Likes     string
	Replies   string
	Thumbnail string
}

func NewEmbedLoader(
	baseURL string,
	endpointURLFormat string,
	tweetCacheKeyProvider cache.KeyProvider,
	collageCache cache.DependentCache,
	maxThumbnailSize uint,
) *EmbedLoader {
	return &EmbedLoader{
		baseURL:               baseURL,
		endpointURLFormat:     endpointURLFormat,
		tweetCacheKeyProvider: tweetCacheKeyProvider,
		collageCache:          collageCache,
		maxThumbnailSize:      maxThumbnailSize,
	}
}

func (l *EmbedLoader) getTweetByID(id string) (*EmbedApiResponse, error) {
	endpointUrl := fmt.Sprintf(l.endpointURLFormat, id)
	resp, err := resolver.RequestGETWithHeaders(endpointUrl, map[string]string{})
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

	var tweet *EmbedApiResponse
	err = json.NewDecoder(resp.Body).Decode(&tweet)
	if err != nil {
		return nil, errors.New("unable to unmarshal response")
	}

	// deleted tweets do not return 404, but contain no data instead
	// example ID: 1616441855495016450
	if tweet.ID == "" {
		return nil, errTweetNotFound
	}

	return tweet, nil
}

func (l *EmbedLoader) Load(
	ctx context.Context,
	tweetID string,
	r *http.Request,
) (*resolver.Response, time.Duration, error) {
	log := logger.FromContext(ctx)

	log.Debugw("[Twitter Embed] Get tweet",
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

	tooltipData := l.buildTweetTooltip(ctx, tweetResp, r)

	var tooltip bytes.Buffer
	if err := embedTweetTooltipTemplate.Execute(&tooltip, tooltipData); err != nil {
		return resolver.Errorf("Twitter tweet template error: %s", err)
	}

	return &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: tooltipData.Thumbnail,
	}, cache.NoSpecialDur, nil
}

func (l *EmbedLoader) buildTweetTooltip(
	ctx context.Context,
	tweet *EmbedApiResponse,
	r *http.Request,
) *embedTweetTooltipData {
	data := &embedTweetTooltipData{}
	data.Text = tweet.Text
	data.Name = tweet.User.Name
	data.Username = tweet.User.ScreenName
	data.Likes = humanize.Number(tweet.FavoriteCount)
	data.Replies = humanize.Number(tweet.ConversationCount)
	data.Timestamp = humanize.CreationDateTime(tweet.CreatedAt)
	data.Thumbnail = l.buildThumbnailURL(ctx, tweet, r)

	return data
}

func (l *EmbedLoader) buildThumbnailURL(
	ctx context.Context,
	tweet *EmbedApiResponse,
	r *http.Request,
) string {
	log := logger.FromContext(ctx)

	numMedia := len(tweet.MediaDetails)
	if numMedia == 0 {
		return ""
	}

	// If tweet contains exactly one image, it will be used as thumbnail
	if numMedia == 1 {
		return tweet.MediaDetails[0].MediaUrl
	}

	// More than one media item, need to compose a thumbnail
	thumb, err := composeThumbnail(ctx, tweet.MediaDetails, int(l.maxThumbnailSize))
	if err != nil {
		log.Errorw("Couldn't compose Twitter collage",
			"err", err,
		)
		return ""
	}

	outputBuf, metaData, err := thumb.ExportNative()
	if err != nil {
		log.Errorw("Couldn't export Twitter collage thumbnail",
			"err", err,
		)
		return ""
	}

	parentKey := l.tweetCacheKeyProvider.CacheKey(ctx, tweet.ID)
	collageKey := buildCollageKey(tweet.ID)
	contentType := utils.MimeType(metaData.Format)

	err = l.collageCache.Insert(ctx, collageKey, parentKey, outputBuf, contentType)
	if err != nil {
		log.Errorw("Couldn't insert Twitter collage thumbnail into cache",
			"err", err,
		)
		return ""
	}

	return utils.FormatGeneratedThumbnailURL(l.baseURL, r, collageKey)
}
