package imgur

import (
	"context"
	"html/template"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/koffeinsource/go-imgur"
)

var (
	// max size of an image before we use a small thumbnail of it
	maxRawImageSize = 50 * 1024

	imageTooltipTemplate = template.Must(template.New("imageTooltipTemplate").Parse(imageTooltip))
)

func Initialize(ctx context.Context, cfg config.APIConfig, pool db.Pool, resolvers *[]resolver.Resolver) {
	log := logger.FromContext(ctx)
	if cfg.ImgurClientID == "" {
		log.Warnw("[Config] imgur-client-id is missing, won't do special responses for imgur")
		return
	}

	imgurClient := &imgur.Client{
		HTTPClient:    resolver.HTTPClient(),
		Log:           &NullLogger{},
		ImgurClientID: cfg.ImgurClientID,
		RapidAPIKEY:   "",
	}

	*resolvers = append(*resolvers, NewResolver(ctx, cfg, pool, imgurClient))
}
