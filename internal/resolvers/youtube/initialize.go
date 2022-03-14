package youtube

import (
	"context"
	"html/template"
	"time"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
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
)

func NewYouTubeVideoResolvers(ctx context.Context, cfg config.APIConfig, pool db.Pool, youtubeClient *youtubeAPI.Service) (resolver.Resolver, resolver.Resolver) {
	videoLoader := NewVideoLoader(youtubeClient)
	videoCache := cache.NewPostgreSQLCache(ctx, cfg, pool, "youtube:video", resolver.NewResponseMarshaller(videoLoader), 24*time.Hour)

	videoResolver := NewYouTubeVideoResolver(videoCache)
	videoShortURLResolver := NewYouTubeVideoShortURLResolver(videoCache)

	return videoResolver, videoShortURLResolver
}

func Initialize(ctx context.Context, cfg config.APIConfig, pool db.Pool, resolvers *[]resolver.Resolver) {
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
	*resolvers = append(*resolvers, NewYouTubeChannelResolver(ctx, cfg, pool, youtubeClient))

	videoResolver, videoShortURLResolver := NewYouTubeVideoResolvers(ctx, cfg, pool, youtubeClient)

	// Handle YouTube video URLs
	*resolvers = append(*resolvers, videoResolver)

	// Handle shortened YouTube video URLs
	*resolvers = append(*resolvers, videoShortURLResolver)
}
