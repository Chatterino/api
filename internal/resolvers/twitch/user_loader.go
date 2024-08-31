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

	var name string
	if strings.ToLower(user.DisplayName) == login {
		name = user.DisplayName
	} else {
		name = fmt.Sprintf("%s (%s)", user.DisplayName, user.Login)
	}

	data := twitchUserTooltipData{
		Name:        name,
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
