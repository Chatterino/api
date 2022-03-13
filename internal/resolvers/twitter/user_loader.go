package twitter

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/resolver"
)

type TwitterUserApiResponse struct {
	Name            string `json:"name"`
	Username        string `json:"screen_name"`
	Description     string `json:"description"`
	Followers       uint64 `json:"followers_count"`
	ProfileImageUrl string `json:"profile_image_url_https"`
}

type twitterUserTooltipData struct {
	Name        string
	Username    string
	Description string
	Followers   string
	Thumbnail   string
}

type UserLoader struct {
	bearerKey string
}

func (l *UserLoader) Load(ctx context.Context, userName string, r *http.Request) (*resolver.Response, time.Duration, error) {
	log := logger.FromContext(ctx)

	log.Debugw("[Twitter] Get user",
		"userName", userName,
	)

	userResp, err := getUserByName(userName, l.bearerKey)
	if err != nil {
		// Error code for "User not found.", as described here:
		// https://developer.twitter.com/en/support/twitter-api/error-troubleshooting#error-codes
		if err.Error() == "50" {
			return &resolver.Response{
				Status:  http.StatusNotFound,
				Message: "Error: Twitter user not found.",
			}, cache.NoSpecialDur, nil
		}

		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "Error getting Twitter user: " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	userData := buildTwitterUserTooltip(userResp)
	var tooltip bytes.Buffer
	if err := twitterUserTooltipTemplate.Execute(&tooltip, userData); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "Twitter user template error: " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	return &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: userData.Thumbnail,
	}, cache.NoSpecialDur, nil
}
