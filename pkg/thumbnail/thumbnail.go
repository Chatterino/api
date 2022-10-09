package thumbnail

import (
	"fmt"
	"net/http"

	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/utils"
	vips "github.com/davidbyttow/govips/v2/vips"
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

func BuildStaticThumbnail(inputBuf []byte, resp *http.Response) ([]byte, error) {
	image, err := vips.NewImageFromBuffer(inputBuf)

	// govips has the height & width values in int, which means we're converting uint to int.
	maxThumbnailSize := int(cfg.MaxThumbnailSize)

	// Only resize if the original image has bigger dimensions than maxThumbnailSize
	if image.Width() <= maxThumbnailSize && image.Height() <= maxThumbnailSize {
		// We don't need to resize image nor does it need to be passed through govips.
		return inputBuf, nil
	}

	importParams := vips.NewImportParams()

	if err != nil {
		return []byte{}, fmt.Errorf("could not load image from url: %s", resp.Request.URL)
	}

	image, err = vips.LoadThumbnailFromBuffer(inputBuf, maxThumbnailSize, maxThumbnailSize, vips.InterestingNone, vips.SizeDown, importParams)

	if err != nil {
		fmt.Println(err)
		return []byte{}, fmt.Errorf("could not transform image from url: %s", resp.Request.URL)
	}

	outputBuf, _, err := image.ExportNative()
	if err != nil {
		return []byte{}, fmt.Errorf("could not export image from url: %s", resp.Request.URL)
	}

	return outputBuf, nil
}
