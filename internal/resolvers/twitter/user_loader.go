package twitter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
	bearerKey         string
	endpointURLFormat string
}

var (
	errUserNotFound = errors.New("user not found")
)

func (l *UserLoader) getUserByName(userName string) (*TwitterUserApiResponse, error) {
	endpointUrl := fmt.Sprintf(l.endpointURLFormat, userName)
	extraHeaders := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", l.bearerKey),
	}
	resp, err := resolver.RequestGETWithHeaders(endpointUrl, extraHeaders)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errUserNotFound
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unhandled status code: %d", resp.StatusCode)
	}

	var user *TwitterUserApiResponse
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, errors.New("unable to unmarshal response")
	}

	/* By default, Twitter returns a low resolution image.
	 * This modification removes "_normal" to get the original sized image, based on Twitter's API documentation:
	 * https://developer.twitter.com/en/docs/twitter-api/v1/accounts-and-users/user-profile-images-and-banners
	 */
	user.ProfileImageUrl = strings.Replace(user.ProfileImageUrl, "_normal", "", 1)

	return user, nil
}

func (l *UserLoader) Load(ctx context.Context, userName string, r *http.Request) (*resolver.Response, time.Duration, error) {
	log := logger.FromContext(ctx)

	log.Debugw("[Twitter] Get user",
		"userName", userName,
	)

	userResp, err := l.getUserByName(userName)
	if err != nil {
		if err == errUserNotFound {
			return &resolver.Response{
				Status:  http.StatusNotFound,
				Message: fmt.Sprintf("Twitter user not found: %s", resolver.CleanResponse(userName)),
			}, cache.NoSpecialDur, nil
		}

		return resolver.Errorf("Twitter user API error: %s", err)
	}

	userData := buildTwitterUserTooltip(userResp)
	var tooltip bytes.Buffer
	if err := twitterUserTooltipTemplate.Execute(&tooltip, userData); err != nil {
		return resolver.Errorf("Twitter user template error: %s", err)
	}

	return &resolver.Response{
		Status:    http.StatusOK,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: userData.Thumbnail,
	}, cache.NoSpecialDur, nil
}
