//go:build !windows
// +build !windows

package thumbnail

import (
	"fmt"
	"net/http"

	"github.com/davidbyttow/govips/v2/vips"
)

func BuildAnimatedThumbnail(inputBuf []byte, resp *http.Response) ([]byte, error) {
	// Only resize if the original image has bigger dimensions than maxThumbnailSize
	// if newWidth < maxThumbnailSize && newHeight < maxThumbnailSize {
	// 	// We don't need to resize image nor does it need to be passed through govips.
	// 	return inputBuf, nil
	// }

	// vips.Startup(nil)
	// defer vips.Shutdown()

	fmt.Println("yo")

	image, err := vips.NewImageFromBuffer(inputBuf)

	params := vips.NewImportParams()
	format := image.Format()

	// n=-1 is used for gifs & webps to make sure to get all frames and not just the first one.
	if format == vips.ImageTypeGIF || format == vips.ImageTypeWEBP {
		params.NumPages.Set(-1)
	}

	if err != nil {
		return []byte{}, fmt.Errorf("could not load image from url: %s", resp.Request.URL)
	}

	// govips has the height & width values in int, which means we're converting uint to int.
	maxThumbnailSize := int(cfg.MaxThumbnailSize)

	newWidth := image.Width()
	newHeight := image.Height()

	fmt.Println(newWidth, "x", newHeight)

	/* Preserve aspect ratio is from previous module, thanks nfnt/resize.
	 * (https://github.com/nfnt/resize/blob/83c6a9932646f83e3267f353373d47347b6036b2/thumbnail.go#L27)
	 */

	// Preserve aspect ratio
	if newWidth > maxThumbnailSize {
		newHeight = newHeight * maxThumbnailSize / newWidth
		if newHeight < 1 {
			newHeight = 1
		}
		newWidth = maxThumbnailSize
	}

	if newHeight > maxThumbnailSize {
		newWidth = newWidth * maxThumbnailSize / newHeight
		if newWidth < 1 {
			newWidth = 1
		}
		newHeight = maxThumbnailSize
	}

	fmt.Println(newWidth, "x", newHeight)

	image, err = vips.LoadThumbnailFromBuffer(inputBuf, newWidth, newHeight, vips.InterestingAll, vips.SizeBoth, params)

	if err != nil {
		fmt.Println(err)
		return []byte{}, fmt.Errorf("could not transform image from url: %s", resp.Request.URL)
	}

	exportParams := vips.NewWebpExportParams()
	exportParams.Quality = 10

	outputBuf, _, err := image.ExportWebp(exportParams)

	fmt.Println("size before:", len(inputBuf), "size after:", len(outputBuf))
	if err != nil {
		return []byte{}, fmt.Errorf("could not export image from url: %s", resp.Request.URL)
	}

	return outputBuf, nil
}
