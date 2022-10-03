//go:build !windows
// +build !windows

package thumbnail

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Chatterino/api/pkg/config"

	vips "github.com/davidbyttow/govips/v2/vips"
	"github.com/discord/lilliput"
)

var encodeOptions = map[string]map[int]int{
	".jpeg": {lilliput.JpegQuality: 85},
	".png":  {lilliput.PngCompression: 7},
	".webp": {lilliput.WebpQuality: 85},
}

func InitializeConfig(passedCfg config.APIConfig) {
	cfg = passedCfg
	vips.Startup(nil)
}

func Shutdown() {
	vips.Shutdown()
}

func BuildAnimatedThumbnail(inputBuf []byte, resp *http.Response) ([]byte, error) {
	// decoder wants []byte, so read the whole file into a buffer
	decoder, err := lilliput.NewDecoder(inputBuf)
	// this error reflects very basic checks,
	// mostly just for the magic bytes of the file to match known image formats
	if err != nil {
		return []byte{}, fmt.Errorf("could not decode image from url: %s", resp.Request.URL)
	}
	defer decoder.Close()

	header, err := decoder.Header()
	// this error is much more comprehensive and reflects
	// format errors
	if err != nil {
		return []byte{}, fmt.Errorf("could not read image header from url: %s", resp.Request.URL)
	}

	newWidth := header.Width()
	newHeight := header.Height()

	// get ready to resize image,
	ops := lilliput.NewImageOps(8192)
	defer ops.Close()

	// create a buffer to store the output image, 2MB in this case
	// If the final image does not fit within this buffer, then we fall back to providing a static thumbnail
	outputImg := make([]byte, 2*1024*1024)

	// lilliput has the height & width values in int, which means we're converting uint to int.
	maxThumbnailSize := int(cfg.MaxThumbnailSize)

	// We don't need to resize image nor does it need to be passed through lilliput.
	// Only resize if the original image has bigger dimensions than maxThumbnailSize
	if newWidth < maxThumbnailSize && newHeight < maxThumbnailSize {
		return inputBuf, nil
	}

	// don't transcode (use existing type)
	outputType := "." + strings.ToLower(decoder.Description())

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

	opts := &lilliput.ImageOptions{
		FileType:      outputType,
		Width:         newWidth,
		Height:        newHeight,
		ResizeMethod:  lilliput.ImageOpsResize,
		EncodeOptions: encodeOptions[outputType],
	}

	// resize and transcode image
	outputImg, err = ops.Transform(decoder, opts, outputImg)
	if err != nil {
		return []byte{}, fmt.Errorf("could not transform image from url: %s", resp.Request.URL)
	}

	return outputImg, nil
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
