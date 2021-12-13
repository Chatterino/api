//go:build windows
// +build windows

package thumbnail

import (
	"errors"
	"net/http"
)

func buildThumbnailByteArray(inputBuf []byte, resp *http.Response) ([]byte, error) {
	// Since the lilliput library currently does not support Windows, we error out early and fall back to the static thumbnail generation
	return nil, errors.New("cannot build animated thumbnails on windows")
}
