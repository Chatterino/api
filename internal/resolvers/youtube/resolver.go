package youtube

import (
	"context"
	"encoding/json"
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
<br><span style="color: #2ecc71;">{{.LikeCount}} likes</span>&nbsp;•&nbsp;<span style="color: #e74c3c;">{{.DislikeCount}} dislikes</span>
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
	// YouTube videos are cahced by video ID
	videoCache = cache.New("youtube_videos", loadVideos, 24*time.Hour)
	// Channels are cached by <channel_type>:<channel_id>
	// See channelCacheKey.go for more information
	channelCache = cache.New("youtube_channels", loadChannels, 24*time.Hour)

	youtubeClient *youtubeAPI.Service

	youtubeVideoTooltipTemplate = template.Must(template.New("youtubeVideoTooltip").Parse(youtubeVideoTooltip))
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

	// Handle YouTube channels (youtube.com/c/chan, youtube.com/chan, youtube.com/user/chan)
	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: func(url *url.URL) bool {
			matches := youtubeChannelRegex.MatchString(url.Path)
			return utils.IsSubdomainOf(url, "youtube.com") && matches
		},
		Run: func(url *url.URL, r *http.Request) ([]byte, error) {
			channelID := getYoutubeChannelIdFromURL(url)

			if channelID.chanType == InvalidChannel {
				return resolver.NoLinkInfoFound, nil
			}

			channelCacheKey := constructCacheKeyFromChannelID(channelID)
			apiResponse := channelCache.Get(channelCacheKey, r)
			return json.Marshal(apiResponse)
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

			apiResponse := videoCache.Get(videoID, r)
			return json.Marshal(apiResponse)
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

			apiResponse := videoCache.Get(videoID, r)
			return json.Marshal(apiResponse)
		},
	})

	return
}
