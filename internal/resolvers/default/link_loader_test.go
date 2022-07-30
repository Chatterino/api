package defaultresolver

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/Chatterino/api/pkg/utils"
	qt "github.com/frankban/quicktest"
)

func TestLinkLoader(t *testing.T) {
	ctx := logger.OnContext(context.Background(), logger.NewTest())
	c := qt.New(t)

	cfg := config.APIConfig{
		MaxContentLength: 5 * 1024 * 1024, // 5 MB
	}

	loader := &LinkLoader{
		baseURL:          cfg.BaseURL,
		maxContentLength: cfg.MaxContentLength,
	}

	c.Run("Early error", func(c *qt.C) {
		tests := []struct {
			inputReq              *http.Request
			inputURL              string
			expectedBytes         []byte
			expectedStatusCode    *int
			expectedContentType   *string
			expectedCacheDuration time.Duration
			expectedError         error
		}{
			{
				inputReq:            newLinkResolverRequest(t, ctx, "GET", " :", nil),
				inputURL:            " :",
				expectedBytes:       resolver.InvalidURLBytes,
				expectedStatusCode:  utils.IntPtr(http.StatusBadRequest),
				expectedContentType: utils.StringPtr("application/json"),
			},
		}

		for _, test := range tests {
			c.Run("", func(c *qt.C) {
				bytes, statusCode, contentType, cacheDuration, err := loader.Load(ctx, test.inputURL, test.inputReq)
				c.Assert(err, qt.IsNil)

				c.Assert(bytes, qt.DeepEquals, test.expectedBytes)
				c.Assert(statusCode, qt.DeepEquals, test.expectedStatusCode)
				c.Assert(contentType, qt.DeepEquals, test.expectedContentType)
				c.Assert(cacheDuration, qt.Equals, test.expectedCacheDuration)
				c.Assert(err, qt.Equals, test.expectedError)
			})
		}
	})
}
