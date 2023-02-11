package defaultresolver_test

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	defaultresolver "github.com/Chatterino/api/internal/resolvers/default"
	"github.com/Chatterino/api/pkg/humanize"
)

func TestMediaResolver(t *testing.T) {
	runMRTest(t, "video/mp4", 12345, "Video (MP4)")
	runMRTest(t, "video/mpeg", 12345, "Video (MPEG)")
	runMRTest(t, "video/ogg", 12345, "Video (OGG)")
	runMRTest(t, "video/webm", 12345, "Video (WEBM)")
	runMRTest(t, "video/x-msvideo", 12345, "Video (AVI)")
	runMRTest(t, "video/nam", 12345, "Video (UNKNOWN)")

	runMRTest(t, "audio/mpeg", 12345, "Audio (MP3)")
	runMRTest(t, "audio/mp4", 12345, "Audio (MP4)")
	runMRTest(t, "audio/ogg", 12345, "Audio (OGG)")
	runMRTest(t, "audio/wav", 12345, "Audio (WAV)")

	runMRTest(t, "audio/wav", 12345, "Audio (WAV)")
}

func runMRTest(t *testing.T, contentType string, size int64, expectedType string) {
	mr := &defaultresolver.MediaResolver{}
	httpRes := &http.Response{
		Header: http.Header{
			"Content-Type": []string{contentType},
		},
		ContentLength: size,
		Request: &http.Request{
			URL: &url.URL{},
		},
	}
	if !mr.Check(context.Background(), contentType) {
		if expectedType == "" {
			return
		}
		t.Errorf("Expected MediaResolver to handle content type: %s", contentType)
		return
	}
	res, err := mr.Run(context.Background(), nil, httpRes)
	if err != nil {
		t.Errorf("MediaResolver should never return an error: %v", err)
		return
	}

	resUnescaped, err := url.PathUnescape(res.Tooltip)

	if !strings.Contains(resUnescaped, expectedType) {
		t.Errorf("Expected: %s, Got: %s", expectedType, res.Tooltip)
	}

	expectedSize := humanize.Bytes(uint64(size))
	if size > 0 && !strings.Contains(resUnescaped, expectedSize) {
		t.Errorf("Expected: %s, Got: %s", expectedSize, res.Tooltip)
	}
}
