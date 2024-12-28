package twitch

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/nicklaw5/helix"
)

type twitchUserTooltipData struct {
	Name        string
	CreatedAt   string
	Description string
	URL         string
}

type twitchUserLiveTooltipData struct {
	Name        string
	CreatedAt   string
	Description string
	URL         string
	Title       string
	Game        string
	Viewers     string
	Uptime      string
}

type UserLoader struct {
	helixAPI TwitchAPIClient
}

func (l *UserLoader) Load(ctx context.Context, login string, r *http.Request) (*resolver.Response, time.Duration, error) {
	log := logger.FromContext(ctx)

	log.Debugw("[Twitch] Get user",
		"login", login,
	)

	response, err := l.helixAPI.GetUsers(&helix.UsersParams{Logins: []string{login}})
	if err != nil {
		log.Errorw("[Twitch] Error getting user",
			"login", login,
			"error", err,
		)

		return resolver.Errorf("Twitch user load error: %s", err)
	}

	if len(response.Data.Users) != 1 {
		return nil, cache.NoSpecialDur, resolver.ErrDontHandle
	}

	var user = response.Data.Users[0]

	streamResponse, err := l.helixAPI.GetStreams(&helix.StreamsParams{UserLogins: []string{login}})
	if err != nil || len(streamResponse.Data.Streams) == 0 {
		return userResponse(login, user)
	}

	return userLiveResponse(login, user, streamResponse.Data.Streams[0])
}

func buildName(login string, user helix.User) string {
	if strings.ToLower(user.DisplayName) == login {
		return user.DisplayName
	} else {
		return fmt.Sprintf("%s (%s)", user.DisplayName, user.Login)
	}
}

func userResponse(login string, user helix.User) (*resolver.Response, time.Duration, error) {
	data := twitchUserTooltipData{
		Name:        buildName(login, user),
		CreatedAt:   humanize.CreationDate(user.CreatedAt.Time),
		Description: user.Description,
		URL:         fmt.Sprintf("https://twitch.tv/%s", user.Login),
	}

	var tooltip bytes.Buffer
	if err := twitchUserTooltip.Execute(&tooltip, data); err != nil {
		return resolver.Errorf("Twitch user template error: %s", err)
	}

	return &resolver.Response{
		Status:    200,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: user.ProfileImageURL,
	}, cache.NoSpecialDur, nil
}

func userLiveResponse(login string, user helix.User, stream helix.Stream) (*resolver.Response, time.Duration, error) {
	data := twitchUserLiveTooltipData{
		Name:        buildName(login, user),
		CreatedAt:   humanize.CreationDate(user.CreatedAt.Time),
		Description: user.Description,
		URL:         fmt.Sprintf("https://twitch.tv/%s", user.Login),
		Title:       stream.Title,
		Game:        stream.GameName,
		Viewers:     humanize.Number(uint64(stream.ViewerCount)),
		Uptime:      humanize.Duration(time.Since(stream.StartedAt)),
	}

	var tooltip bytes.Buffer
	if err := twitchUserLiveTooltip.Execute(&tooltip, data); err != nil {
		return resolver.Errorf("Twitch user template error: %s", err)
	}

	thumbnail := strings.ReplaceAll(stream.ThumbnailURL, "{width}", "1280")
	thumbnail = strings.ReplaceAll(thumbnail, "{height}", "720")

	return &resolver.Response{
		Status:    200,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: thumbnail,
	}, cache.NoSpecialDur, nil
}
