package youtube

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
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
	// YouTube videos are cached by video ID
	videoCache cache.Cache
	// Channels are cached by <channel_type>:<channel_id>
	// See channelCacheKey.go for more information
	channelCache cache.Cache

	youtubeClient *youtubeAPI.Service

	youtubeVideoTooltipTemplate   = template.Must(template.New("youtubeVideoTooltip").Parse(youtubeVideoTooltip))
	youtubeChannelTooltipTemplate = template.Must(template.New("youtubeChannelTooltip").Parse(youtubeChannelTooltip))

	youtubeChannelRegex = regexp.MustCompile(`/(user|c(hannel)?)/[\w._\-']+`)
)

func New(cfg config.APIConfig) (resolvers []resolver.CustomURLManager) {
	if cfg.YoutubeApiKey == "" {
		log.Println("[Config] youtube-api-key is missing, won't do special responses for youtube")
		return
	}

	ctx := context.Background()
	var err error
	if youtubeClient, err = youtubeAPI.NewService(ctx, option.WithAPIKey(cfg.YoutubeApiKey)); err != nil {
		log.Println("Error initialization youtube api client:", err)
		return
	}

	videoCache = cache.NewPostgreSQLCache(cfg, "youtube_videos", resolver.MarshalResponse(loadVideos), 24*time.Hour)
	channelCache = cache.NewPostgreSQLCache(cfg, "youtube_channels", resolver.MarshalResponse(loadChannels), 24*time.Hour)

	// Handle YouTube channels (youtube.com/c/chan, youtube.com/chan, youtube.com/user/chan)
	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: func(url *url.URL) bool {
			matches := youtubeChannelRegex.MatchString(url.Path)
			return utils.IsSubdomainOf(url, "youtube.com") && matches
		},
		Run: func(url *url.URL, r *http.Request) ([]byte, error) {
			channelID := getYoutubeChannelIDFromURL(url)

			if channelID.chanType == InvalidChannel {
				return resolver.NoLinkInfoFound, nil
			}

			channelCacheKey := constructCacheKeyFromChannelID(channelID)
			return channelCache.Get(channelCacheKey, r)
		},
	})

	// Handle YouTube video URLs
	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: func(url *url.URL) bool {
			return utils.IsSubdomainOf(url, "youtube.com")
		},
		Run: func(url *url.URL, r *http.Request) ([]byte, error) {
			videoID := getYoutubeVideoIDFromURL(url)

			if videoID == "" {
				return resolver.NoLinkInfoFound, nil
			}

			return videoCache.Get(videoID, r)
		},
	})

	// Handle shortened YouTube video URLs
	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: func(url *url.URL) bool {
			return url.Host == "youtu.be"
		},
		Run: func(url *url.URL, r *http.Request) ([]byte, error) {
			videoID := getYoutubeVideoIDFromURL2(url)

			if videoID == "" {
				return resolver.NoLinkInfoFound, nil
			}

			return videoCache.Get(videoID, r)
		},
	})

	return
}
