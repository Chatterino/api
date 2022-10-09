package thumbnail

import (
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/utils"
)

var (
	supportedThumbnails = []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
	animatedThumbnails  = []string{"image/gif", "image/webp"}

	cfg config.APIConfig
)

func IsSupportedThumbnailType(contentType string) bool {
	return utils.Contains(supportedThumbnails, contentType)
}

func IsAnimatedThumbnailType(contentType string) bool {
	return utils.Contains(animatedThumbnails, contentType)
}
