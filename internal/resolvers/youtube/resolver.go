package youtube

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"

	"google.golang.org/api/option"
	youtubeAPI "google.golang.org/api/youtube/v3"
)

const (
	youtubeVideoTooltip = `<div style="text-align: left;">
<b>{{.Title}}</b>
<br><b>Channel:</b> {{.ChannelTitle}}
<br><b>Duration:</b> {{.Duration}}
<br><b>Published:</b> {{.PublishDate}}
<br><b>Views:</b> {{.Views}}
{{ if .AgeRestricted }}<br><b><span style="color: red;">AGE RESTRICTED</span></b>{{ end }}
<br><span style="color: #2ecc71;">{{.LikeCount}} likes</span>&nbsp;•&nbsp;<span style="color: #808892;">{{.CommentCount}} comments</span>
</div>
`

	youtubeChannelTooltip = `<div style="text-align: left;">
<b>{{.Title}}</b>
<br><b>Joined Date:</b> {{.JoinedDate}}
<br><b>Subscribers:</b> {{.Subscribers}}
<br><b>Views:</b> {{.Views}}
</div>
`
)

var (
	youtubeVideoTooltipTemplate   = template.Must(template.New("youtubeVideoTooltip").Parse(youtubeVideoTooltip))
	youtubeChannelTooltipTemplate = template.Must(template.New("youtubeChannelTooltip").Parse(youtubeChannelTooltip))

	youtubeChannelRegex = regexp.MustCompile(`/(user|c(hannel)?)/[\w._\-']+`)
)

type YouTubeChannelResolver struct {
	channelCache  cache.Cache
	youtubeClient *youtubeAPI.Service
}

func (r *YouTubeChannelResolver) Check(ctx context.Context, url *url.URL) bool {
	matches := youtubeChannelRegex.MatchString(url.Path)
	return utils.IsSubdomainOf(url, "youtube.com") && matches
}

func (r *YouTubeChannelResolver) Run(ctx context.Context, url *url.URL, req *http.Request) ([]byte, error) {
	channelID := getYoutubeChannelIDFromURL(url)

	if channelID.chanType == InvalidChannel {
		return resolver.NoLinkInfoFound, nil
	}

	channelCacheKey := constructCacheKeyFromChannelID(channelID)
	return r.channelCache.Get(ctx, channelCacheKey, req)
}

func (r *YouTubeChannelResolver) Load(ctx context.Context, channelCacheKey string, req *http.Request) (*resolver.Response, time.Duration, error) {
	youtubeChannelParts := []string{
		"statistics",
		"snippet",
	}

	log.Println("[YouTube] GET channel", channelCacheKey)
	builtRequest := r.youtubeClient.Channels.List(youtubeChannelParts)

	channelID := deconstructChannelIDFromCacheKey(channelCacheKey)
	if channelID.chanType == CustomChannel {
		// Channels with custom URLs aren't searchable with the channel/list endpoint
		// The only average way to do this at the moment is to do a YouTube search of that name
		// and filter for channels. Not ideal...

		searchRequest := r.youtubeClient.Search.List([]string{"snippet"}).Q(channelID.ID).Type("channel")
		response, err := searchRequest.MaxResults(1).Do()

		if err != nil {
			return &resolver.Response{
				Status:  500,
				Message: "youtube search api error " + resolver.CleanResponse(err.Error()),
			}, 1 * time.Hour, nil
		}

		if len(response.Items) != 1 {
			return nil, cache.NoSpecialDur, errors.New("channel search response is not size 1")
		}

		channelID.ID = response.Items[0].Snippet.ChannelId
	}

	switch channelID.chanType {
	case UserChannel:
		builtRequest = builtRequest.ForUsername(channelID.ID)
	case IdentifierChannel:
		builtRequest = builtRequest.Id(channelID.ID)
	case CustomChannel:
		builtRequest = builtRequest.Id(channelID.ID)
	case InvalidChannel:
		return &resolver.Response{
			Status:  500,
			Message: "cached channel ID is invalid",
		}, 1 * time.Hour, nil
	}

	youtubeResponse, err := builtRequest.Do()

	if err != nil {
		return &resolver.Response{
			Status:  500,
			Message: "youtube api error " + resolver.CleanResponse(err.Error()),
		}, 1 * time.Hour, nil
	}

	if len(youtubeResponse.Items) != 1 {
		return nil, cache.NoSpecialDur, errors.New("channel response is not size 1")
	}

	channel := youtubeResponse.Items[0]

	data := youtubeChannelTooltipData{
		Title:       channel.Snippet.Title,
		JoinedDate:  humanize.CreationDateRFC3339(channel.Snippet.PublishedAt),
		Subscribers: humanize.Number(channel.Statistics.SubscriberCount),
		Views:       humanize.Number(channel.Statistics.ViewCount),
	}

	var tooltip bytes.Buffer
	if err := youtubeChannelTooltipTemplate.Execute(&tooltip, data); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "youtube template error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	thumbnail := channel.Snippet.Thumbnails.Default.Url
	if channel.Snippet.Thumbnails.Medium != nil {
		thumbnail = channel.Snippet.Thumbnails.Medium.Url
	}

	return &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: thumbnail,
	}, cache.NoSpecialDur, nil
}

func NewYouTubeChannelResolver(ctx context.Context, cfg config.APIConfig, youtubeClient *youtubeAPI.Service) *YouTubeChannelResolver {
	r := &YouTubeChannelResolver{
		youtubeClient: youtubeClient,
	}

	// Use YoutubeChannelResolver's Load function, wrapped by the resolvers ResponseMarshaller
	channelCache := cache.NewPostgreSQLCache(ctx, cfg, "youtube_channels", resolver.NewResponseMarshaller(r), 24*time.Hour)

	// Set the channelCache variable for YoutubeChannelResolver so it can be used from Run
	r.channelCache = channelCache

	return r
}

func NewYouTubeVideoResolvers(ctx context.Context, cfg config.APIConfig, youtubeClient *youtubeAPI.Service) (resolver.Resolver, resolver.Resolver) {
	videoResolver := NewYouTubeVideoResolver(ctx, cfg, youtubeClient)

	// The Short URL resolver shared the cache that the videoResolver manages
	videoShortURLResolver := NewYouTubeVideoShortURLResolver(videoResolver.videoCache)

	return videoResolver, videoShortURLResolver
}

func Initialize(ctx context.Context, cfg config.APIConfig, resolvers *[]resolver.Resolver) {
	log := logger.FromContext(ctx)
	if cfg.YoutubeApiKey == "" {
		log.Warnw("[Config] youtube-api-key is missing, won't do special responses for YouTube")
		return
	}

	youtubeClient, err := youtubeAPI.NewService(ctx, option.WithAPIKey(cfg.YoutubeApiKey))
	if err != nil {
		log.Warnw("Error initialization YouTube api client",
			"error", err,
		)
		return
	}

	// Handle YouTube channels (youtube.com/c/chan, youtube.com/chan, youtube.com/user/chan)
	*resolvers = append(*resolvers, NewYouTubeChannelResolver(ctx, cfg, youtubeClient))

	videoResolver, videoShortURLResolver := NewYouTubeVideoResolvers(ctx, cfg, youtubeClient)

	// Handle YouTube video URLs
	*resolvers = append(*resolvers, videoResolver)

	// Handle shortened YouTube video URLs
	*resolvers = append(*resolvers, videoShortURLResolver)
}
