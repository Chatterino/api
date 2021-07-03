package twitchapiclient

import (
	"errors"
	"net/http"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/nicklaw5/helix"
)

func loadUsername(helixClient *helix.Client) func(key string, r *http.Request) (interface{}, time.Duration, error) {
	return func(twitchUserID string, r *http.Request) (interface{}, time.Duration, error) {
		params := &helix.UsersParams{
			IDs: []string{
				twitchUserID,
			},
		}

		response, err := helixClient.GetUsers(params)
		if err != nil {
			return nil, cache.NoSpecialDur, err
		}

		if len(response.Data.Users) != 1 {
			return nil, cache.NoSpecialDur, errors.New("no user with this ID found")
		}

		user := response.Data.Users[0]

		return user.Login, cache.NoSpecialDur, nil
	}
}
