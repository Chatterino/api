package youtube

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/Chatterino/api/pkg/cache"
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
)

var (
	videoCache = cache.New("youtube_videos", load, 24*time.Hour)

	youtubeClient *youtubeAPI.Service

	youtubeVideoTooltipTemplate = template.Must(template.New("youtubeVideoTooltip").Parse(youtubeVideoTooltip))
)

func New() (resolvers []resolver.CustomURLManager) {
	apiKey, exists := os.LookupEnv("YOUTUBE_API_KEY")
	if !exists {
		log.Println("No YOUTUBE_API_KEY specified, won't do special responses for youtube")
		return
	}

	ctx := context.Background()
	var err error
	if youtubeClient, err = youtubeAPI.NewService(ctx, option.WithAPIKey(apiKey)); err != nil {
		log.Println("Error initialization youtube api client:", err)
		return
	}

	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: func(url *url.URL) bool {
			return utils.IsSubdomainOf(url, "youtube.com")
		},
		Run: func(url *url.URL) ([]byte, error) {
			videoID := getYoutubeVideoIDFromURL(url)

			if videoID == "" {
				return resolver.NoLinkInfoFound, nil
			}

			apiResponse := videoCache.Get(videoID, nil)
			return json.Marshal(apiResponse)
		},
	})

	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: func(url *url.URL) bool {
			return url.Host == "youtu.be"
		},
		Run: func(url *url.URL) ([]byte, error) {
			videoID := getYoutubeVideoIDFromURL2(url)

			if videoID == "" {
				return resolver.NoLinkInfoFound, nil
			}

			apiResponse := videoCache.Get(videoID, nil)
			return json.Marshal(apiResponse)
		},
	})

	return
}
