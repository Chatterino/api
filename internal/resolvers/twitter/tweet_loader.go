package twitter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/davidbyttow/govips/v2/vips"
)

type TweetApiResponse struct {
	Data     Data     `json:"data"`
	Includes Includes `json:"includes"`
}

type PublicMetrics struct {
	RetweetCount uint64 `json:"retweet_count"`
	LikeCount    uint64 `json:"like_count"`
}

type Data struct {
	PublicMetrics PublicMetrics `json:"public_metrics"`
	ID            string        `json:"id"`
	Text          string        `json:"text"`
	CreatedAt     string        `json:"created_at"`
}
type Media struct {
	URL string `json:"url"`
}
type Users struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}
type Includes struct {
	Media []Media `json:"media"`
	Users []Users `json:"users"`
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
	baseURL               string
	bearerKey             string
	endpointURLFormat     string
	tweetCacheKeyProvider cache.KeyProvider
	collageCache          cache.DependentCache
	maxThumbnailSize      uint
}

var (
	errTweetNotFound     = errors.New("tweet not found")
	errNoMediaDownloaded = errors.New("couldn't download any of the attached media items")
)

func NewTweetLoader(
	baseURL string,
	bearerKey string,
	endpointURLFormat string,
	tweetCacheKeyProvider cache.KeyProvider,
	collageCache cache.DependentCache,
	maxThumbnailSize uint,
) *TweetLoader {
	return &TweetLoader{
		baseURL:               baseURL,
		bearerKey:             bearerKey,
		endpointURLFormat:     endpointURLFormat,
		tweetCacheKeyProvider: tweetCacheKeyProvider,
		collageCache:          collageCache,
		maxThumbnailSize:      maxThumbnailSize,
	}
}

func buildCollageKey(tweetID string) string {
	return fmt.Sprintf("twitter:collage:%s", tweetID)
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

	// deleted tweets do not return 404, but contain no data instead
	// example ID: 1616441855495016450
	if tweet.Data.ID == "" {
		return nil, errTweetNotFound
	}

	return tweet, nil
}

func (l *TweetLoader) Load(
	ctx context.Context,
	tweetID string,
	r *http.Request,
) (*resolver.Response, time.Duration, error) {
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

	tooltipData := l.buildTweetTooltip(ctx, tweetResp, r)

	var tooltip bytes.Buffer
	if err := tweetTooltipTemplate.Execute(&tooltip, tooltipData); err != nil {
		return resolver.Errorf("Twitter tweet template error: %s", err)
	}

	return &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: tooltipData.Thumbnail,
	}, cache.NoSpecialDur, nil
}

func (l *TweetLoader) buildTweetTooltip(
	ctx context.Context,
	tweet *TweetApiResponse,
	r *http.Request,
) *tweetTooltipData {
	data := &tweetTooltipData{}
	data.Text = tweet.Data.Text
	data.Name = tweet.Includes.Users[0].Name
	data.Username = tweet.Includes.Users[0].Username
	data.Likes = humanize.Number(tweet.Data.PublicMetrics.LikeCount)
	data.Retweets = humanize.Number(tweet.Data.PublicMetrics.RetweetCount)

	// TODO: what time format is this exactly? can we move to humanize a la CreationDteRFC3339?
	timestamp, err := time.Parse("Mon Jan 2 15:04:05 -0700 2006", tweet.Data.CreatedAt)
	if err != nil {
		data.Timestamp = ""
	} else {
		data.Timestamp = humanize.CreationDateTime(timestamp)
	}

	data.Thumbnail = l.buildThumbnailURL(ctx, tweet, r)

	return data
}

func (l *TweetLoader) buildThumbnailURL(
	ctx context.Context,
	tweet *TweetApiResponse,
	r *http.Request,
) string {
	log := logger.FromContext(ctx)

	numMedia := len(tweet.Includes.Media)
	if numMedia == 1 {
		// If tweet contains exactly one image, it will be used as thumbnail
		return tweet.Includes.Media[0].URL
	}

	// More than one media item, need to compose a thumbnail
	thumb, err := l.composeThumbnail(ctx, tweet.Includes.Media)
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

	parentKey := l.tweetCacheKeyProvider.CacheKey(ctx, tweet.Data.ID)
	collageKey := buildCollageKey(tweet.Data.ID)
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

func (l *TweetLoader) composeThumbnail(
	ctx context.Context,
	mediaEntities []Media,
) (*vips.ImageRef, error) {
	log := logger.FromContext(ctx)

	numMedia := len(mediaEntities)

	// First, download all images
	downloaded := make([]*vips.ImageRef, numMedia)
	wg := new(sync.WaitGroup)
	wg.Add(numMedia)

	for idx, media := range mediaEntities {
		idx := idx
		media := media

		go func() {
			defer wg.Done()

			resp, err := resolver.RequestGET(ctx, media.URL)
			if err != nil {
				log.Errorw("Couldn't download Twitter media",
					"url", media.URL,
					"err", err,
				)
				return
			}

			buf, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Errorw("Couldn't read response body",
					"url", media.URL,
					"err", err,
				)
				return
			}

			ref, err := vips.NewImageFromBuffer(buf)
			if err != nil {
				log.Errorw("Couldn't convert buffer to vips.ImageRef",
					"err", err,
				)
				return
			}

			downloaded[idx] = ref
		}()
	}

	wg.Wait()

	// Prepare downloaded media for collage
	var collageSource []*vips.ImageRef

	// Keep track of smallest dimension for proper resizing later
	smallestDimensionFound := math.MaxFloat64

	// In a first pass, check downloaded media to determine the smallest dimension
	for _, ref := range downloaded {
		if ref != nil {
			smallerDimensionCur := math.Min(float64(ref.Width()), float64(ref.Height()))
			smallestDimensionFound = math.Min(smallestDimensionFound, smallerDimensionCur)

			collageSource = append(collageSource, ref)
		}
	}

	// In the second pass, resize the images according to smallest dimension
	for _, ref := range collageSource {
		ref.ThumbnailWithSize(
			int(smallestDimensionFound), int(smallestDimensionFound), vips.InterestingCentre,
			vips.SizeDown,
		)
	}

	if len(collageSource) == 0 {
		log.Errorw("No Twitter media could be downloaded, cannot build collage")
		return nil, errNoMediaDownloaded
	}

	// Now compose the thumbnail
	stem := collageSource[0]

	err := stem.ArrayJoin(collageSource[1:], 2)
	if err != nil {
		log.Errorw("Couldn't ArrayJoin imags",
			"err", err,
		)
		return nil, err
	}

	maxThumbnailSize := int(l.maxThumbnailSize)
	err = stem.ThumbnailWithSize(
		maxThumbnailSize, maxThumbnailSize, vips.InterestingNone, vips.SizeDown,
	)
	if err != nil {
		log.Errorw("Couldn't generate thumbnail",
			"err", err,
		)
		return nil, err
	}

	return stem, nil
}
