package youtube

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"

	"google.golang.org/api/option"
	youtubeAPI "google.golang.org/api/youtube/v3"
)

const (
	youtubeTooltip = `<div style="text-align: left;">
<b>{{.Title}}</b>
<br><b>Channel:</b> {{.ChannelTitle}}
<br><b>Duration:</b> {{.Duration}}
<br><b>Published:</b> {{.PublishDate}}
<br><b>Views:</b> {{.Views}}
<br><span style="color: #2ecc71;">{{.LikeCount}} likes</span>&nbsp;•&nbsp;<span style="color: #e74c3c;">{{.DislikeCount}} dislikes</span>
</div>
`
)

var (
	videoCache = cache.New("youtube_videos", load, 24*time.Hour)

	youtubeClient *youtubeAPI.Service

	youtubeTooltipTemplate = template.Must(template.New("youtubeTooltip").Parse(youtubeTooltip))
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
