package twitch

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/nicklaw5/helix"
)

type twitchUserTooltipData struct {
	Login       string
	DisplayName string
	CreatedAt   string
	Description string
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

	data := twitchUserTooltipData{
		Login:       user.Login,
		DisplayName: user.DisplayName,
		CreatedAt:   humanize.CreationDate(user.CreatedAt.Time),
		Description: user.Description,
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
