package youtube

import (
	"context"
	"html/template"

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

	youtubePlaylistTooltip = `<div style="text-align: left;">
<b>{{.Title}}</b>
<br><b>Description:</b> {{.Description}}
<br><b>Channel:</b> {{.Channel}}
<br><b>Videos:</b> {{.VideoCount}}
<br><b>Published:</b> {{.PublishedAt}}
</div>
`

	youtubeStreamTooltip = `<div style="text-align: left;">
<b>{{.Title}}</b>
<br><b>Channel:</b> {{.ChannelTitle}}
<br><b>Uptime:</b> {{.Uptime}}
<br><b>Viewers:</b> {{.Viewers}}
<br><b><span style="color: #ff0000;">Live</span></b>&nbsp;•&nbsp;<span style="color: #2ecc71;">{{.LikeCount}} likes</span>
</div>
`
)

var (
	youtubeVideoTooltipTemplate    = template.Must(template.New("youtubeVideoTooltip").Parse(youtubeVideoTooltip))
	youtubeChannelTooltipTemplate  = template.Must(template.New("youtubeChannelTooltip").Parse(youtubeChannelTooltip))
	youtubePlaylistTooltipTemplate = template.Must(template.New("youtubePlaylistTooltip").Parse(youtubePlaylistTooltip))
	youtubeStreamTooltipTemplate   = template.Must(template.New("youtubeStreamTooltip").Parse(youtubeStreamTooltip))
)

func NewYouTubeVideoResolvers(ctx context.Context, cfg config.APIConfig, pool db.Pool, youtubeClient *youtubeAPI.Service) (resolver.Resolver, resolver.Resolver) {
	videoLoader := NewVideoLoader(youtubeClient)
	videoCache := cache.NewPostgreSQLCache(
		ctx, cfg, pool, cache.NewPrefixKeyProvider("youtube:video"), videoLoader, cfg.YoutubeVideoCacheDuration,
	)

	videoResolver := NewYouTubeVideoResolver(videoCache)
	videoShortURLResolver := NewYouTubeVideoShortURLResolver(videoCache)

	return videoResolver, videoShortURLResolver
}

func Initialize(ctx context.Context, cfg config.APIConfig, pool db.Pool, resolvers *[]resolver.Resolver) {
	log := logger.FromContext(ctx)

	if cfg.YoutubeApiKey == "" {
		log.Warnw("[Config] youtube-api-key missing, won't do special responses for YouTube")
		return

	}

	youtubeClient, err := youtubeAPI.NewService(ctx, option.WithAPIKey(cfg.YoutubeApiKey))
	if err != nil {
		log.Warnw("[Config] Failed to create youtube client, won't do special responses for YouTube",
			"error", err,
		)
		return
	}

	playlistResolver := NewYouTubePlaylistResolver(ctx, cfg, pool, youtubeClient)

	// Handle YouTube playlists
	*resolvers = append(*resolvers, playlistResolver)

	// Handle YouTube channels (youtube.com/c/chan, youtube.com/chan, youtube.com/user/chan)
	*resolvers = append(*resolvers, NewYouTubeChannelResolver(ctx, cfg, pool, youtubeClient))

	videoResolver, videoShortURLResolver := NewYouTubeVideoResolvers(ctx, cfg, pool, youtubeClient)

	// Handle YouTube video URLs
	*resolvers = append(*resolvers, videoResolver)

	// Handle shortened YouTube video URLs
	*resolvers = append(*resolvers, videoShortURLResolver)
}
