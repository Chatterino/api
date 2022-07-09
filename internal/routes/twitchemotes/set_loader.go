package twitchemotes

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/nicklaw5/helix"
)

var (
	errInvalidEmoteID = errors.New("invalid emote id")
	errUnableToHandle = errors.New("unable to handle twitchemotes requests")
)

type TwitchEmotesErrorResponse struct {
	Status  int
	Message string
}

type EmoteSet struct {
	ChannelName string `json:"channel_name"`
	ChannelID   string `json:"channel_id"`
	Type        string `json:"type"`
	Tier        int    `json:"tier"`
	Custom      bool   `json:"custom"`
}

type twitchEmotesError struct {
	Status          int
	UnderlyingError error
}

func (e *twitchEmotesError) Error() string {
	return e.UnderlyingError.Error()
}

type TwitchemotesLoader struct {
	helixAPI           *helix.Client
	helixUsernameCache cache.Cache
}

func (l *TwitchemotesLoader) Load(ctx context.Context, setID string, r *http.Request) ([]byte, *int, *string, time.Duration, error) {
	if l.helixAPI == nil || l.helixUsernameCache == nil {
		return nil, nil, nil, 0, &twitchEmotesError{
			UnderlyingError: errUnableToHandle,
			Status:          500,
		}
	}

	params := &helix.GetEmoteSetsParams{
		EmoteSetIDs: []string{
			setID,
		},
	}
	resp, err := l.helixAPI.GetEmoteSets(params)
	if err != nil {
		return nil, nil, nil, 0, &twitchEmotesError{
			UnderlyingError: err,
			Status:          500,
		}
	}

	if len(resp.Data.Emotes) == 0 {
		return nil, nil, nil, 0, &twitchEmotesError{
			UnderlyingError: errInvalidEmoteID,
			Status:          404,
		}
	}

	emote := resp.Data.Emotes[0]

	var username string

	// For Emote Sets 0 (global) and 19194 (prime emotes), the Owner ID returns 0
	// 0 is not a valid Twitch User ID, so hardcode the username to Twitch
	if emote.OwnerID == "0" {
		username = "Twitch"
	} else {
		// Load username from Helix
		if usernameBytes, err := l.helixUsernameCache.Get(ctx, emote.OwnerID, nil); err != nil {
			return nil, nil, nil, 0, &twitchEmotesError{
				UnderlyingError: err,
				Status:          404,
			}
		} else {
			username = string(usernameBytes.Payload)
		}
	}

	emoteSet := EmoteSet{
		ChannelName: username,
		ChannelID:   emote.OwnerID,
		Type:        emote.EmoteType,
	}

	return utils.MarshalNoDur(&emoteSet)
}
