// +build !windows

package thumbnail

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/discord/lilliput"
)

var (
	encodeOptions = map[string]map[int]int{
		".jpeg": {lilliput.JpegQuality: 85},
		".png":  {lilliput.PngCompression: 7},
		".webp": {lilliput.WebpQuality: 85},
	}
)

func buildThumbnailByteArray(inputBuf []byte, resp *http.Response) ([]byte, error) {
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

	// don't transcode (use existing type)
	outputType := "." + strings.ToLower(decoder.Description())

	// We want to default to no resizing.
	resizeMethod := lilliput.ImageOpsNoResize

	// Only trigger if original image has higher values than maxThumbnailSize
	if newWidth > maxThumbnailSize || newHeight > maxThumbnailSize {
		resizeMethod = lilliput.ImageOpsResize // We want to resize

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
	}

	opts := &lilliput.ImageOptions{
		FileType:      outputType,
		Width:         newWidth,
		Height:        newHeight,
		ResizeMethod:  resizeMethod,
		EncodeOptions: encodeOptions[outputType],
	}

	// resize and transcode image
	outputImg, err = ops.Transform(decoder, opts, outputImg)
	if err != nil {
		return []byte{}, fmt.Errorf("could not transform image from url: %s", resp.Request.URL)
	}

	return outputImg, nil
}
