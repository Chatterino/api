package imgur

import (
	"bytes"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
)

func buildTooltip(miniData miniImage) (*resolver.Response, time.Duration, error) {
	var tooltip bytes.Buffer

	if err := imageTooltipTemplate.Execute(&tooltip, &miniData); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "imgur image template error: " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	return &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: miniData.Link,
	}, cache.NoSpecialDur, nil
}

func load(urlString string, r *http.Request) (*resolver.Response, time.Duration, error) {
	genericInfo, _, err := apiClient.GetInfoFromURL(urlString)
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
	miniData.Link = utils.FormatThumbnailURL(baseURL, r, miniData.Link)
	return buildTooltip(miniData)
}
