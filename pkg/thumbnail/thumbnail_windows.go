//go:build windows
// +build windows

package thumbnail

import (
	"errors"
	"net/http"

	"github.com/Chatterino/api/pkg/config"
)

func InitializeConfig(passedCfg config.APIConfig) {
	cfg = passedCfg
}

func Shutdown() {
	// Nothing to shut down on Windows
}

func BuildAnimatedThumbnail(inputBuf []byte, resp *http.Response) ([]byte, error) {
	// Since the lilliput library currently does not support Windows, we error out early and fall back to the static thumbnail generation
	return nil, errors.New("cannot build animated thumbnails on windows")
}

func BuildStaticThumbnail(inputBuf []byte, resp *http.Response) ([]byte, error) {
	// govips can run on Windows with the proper setup. If you would like to contribute Windows
	// support, see https://github.com/davidbyttow/govips#windows and open a PR at
	// https://github.com/Chatterino/api/pulls.
	return nil, errors.New("cannot build static thumbnails on windows")
}
