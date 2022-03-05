package imgur

import (
	"context"
	"net/http"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

type Loader struct {
	baseURL   string
	apiClient ImgurClient
}

func (l *Loader) Load(ctx context.Context, urlString string, r *http.Request) (*resolver.Response, time.Duration, error) {
	genericInfo, _, err := l.apiClient.GetInfoFromURL(urlString)
	if err != nil {
		return &resolver.Response{
			Status:  http.StatusOK,
			Tooltip: "Error getting imgur API information for URL",
		}, cache.NoSpecialDur, resolver.ErrDontHandle
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
		return &resolver.Response{
			Status:  http.StatusOK,
			Tooltip: "Error getting imgur API information for URL",
		}, cache.NoSpecialDur, nil
	}

	// Proxy imgur thumbnails
	miniData.Link = utils.FormatThumbnailURL(l.baseURL, r, miniData.Link)

	return buildTooltip(miniData)
}
