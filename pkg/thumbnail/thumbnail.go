package thumbnail

import (
	"github.com/Chatterino/api/pkg/config"
)

var (
	supportedThumbnails = []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
	animatedThumbnails  = []string{"image/gif", "image/webp"}

	cfg config.APIConfig
)

func IsSupportedThumbnail(contentType string) bool {
	for _, supportedType := range supportedThumbnails {
		if contentType == supportedType {
			return true
		}
	}

	return false
}

func IsAnimatedThumbnailType(contentType string) bool {
	return utils.Contains(animatedThumbnails, contentType)
}
