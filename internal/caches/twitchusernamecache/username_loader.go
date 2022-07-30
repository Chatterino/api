package twitchusernamecache

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/nicklaw5/helix"
)

type UsernameLoader struct {
	helixClient *helix.Client
}

func (l *UsernameLoader) Load(ctx context.Context, twitchUserID string, req *http.Request) ([]byte, *int, *string, time.Duration, error) {
	params := &helix.UsersParams{
		IDs: []string{
			twitchUserID,
		},
	}

	response, err := l.helixClient.GetUsers(params)
	if err != nil {
		return nil, nil, nil, cache.NoSpecialDur, err
	}

	if len(response.Data.Users) != 1 {
		return nil, nil, nil, cache.NoSpecialDur, errors.New("no user with this ID found")
	}

	user := response.Data.Users[0]

	return []byte(user.Login), nil, nil, cache.NoSpecialDur, nil
}
