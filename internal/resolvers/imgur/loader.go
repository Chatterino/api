package imgur

import (
	"context"
	"net/http"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

type Loader struct {
	baseURL   string
	apiClient ImgurClient
}

func (l *Loader) Load(ctx context.Context, urlString string, r *http.Request) (*resolver.Response, time.Duration, error) {
	log := logger.FromContext(ctx)

	genericInfo, _, err := l.apiClient.GetInfoFromURL(urlString)
	if err != nil {
		log.Warnw("Error getting imgur info from URL",
			"url", urlString,
			"error", err,
		)
		return nil, cache.NoSpecialDur, resolver.ErrDontHandle
	}

	if genericInfo == nil {
		log.Warnw("Missing imgur info",
			"url", urlString,
			"error", err,
		)
		return nil, cache.NoSpecialDur, resolver.ErrDontHandle
	}

	var miniData miniImage

	if genericInfo.Image != nil {
		miniData = makeMiniImage(*genericInfo.Image)
	} else if genericInfo.GImage != nil {
		miniData = makeMiniImageFromGImage(*genericInfo.GImage)
	} else if genericInfo.Album != nil {
		ptr := genericInfo.Album
		if len(ptr.Images) == 0 {
			return &resolver.Response{
				Status:  http.StatusOK,
				Tooltip: "Empty album",
			}, cache.NoSpecialDur, nil
		}

		miniData = makeMiniImage(ptr.Images[0])

		miniData.Album = true
		miniData.Title = ptr.Title
		miniData.Description = ptr.Description
	} else if genericInfo.GAlbum != nil {
		ptr := genericInfo.GAlbum
		if len(ptr.Images) == 0 {
			return &resolver.Response{
				Status:  http.StatusOK,
				Tooltip: "Empty album",
			}, cache.NoSpecialDur, nil
		}

		miniData = makeMiniImage(ptr.Images[0])

		miniData.Album = true
		miniData.Title = ptr.Title
		miniData.Description = ptr.Description
	} else {
		log.Warnw("Missing relevant imgur response",
			"url", urlString,
		)

		return nil, resolver.NoSpecialDur, resolver.ErrDontHandle
	}

	// Proxy imgur thumbnails
	if miniData.Link != "" {
		miniData.Link = utils.FormatThumbnailURL(l.baseURL, r, miniData.Link)
	}

	return buildTooltip(miniData)
}
