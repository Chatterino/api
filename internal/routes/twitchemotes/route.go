package twitchemotes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/nicklaw5/helix"
)

var (
	errInvalidEmoteID = errors.New("invalid emote id")
	errUnableToHandle = errors.New("unable to handle twitchemotes requests")

	helixAPI           *helix.Client
	helixUsernameCache cache.Cache
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

func setHandler(ctx context.Context, helixUsernameCache, twitchemotesCache cache.Cache, w http.ResponseWriter, r *http.Request) {
	setID := chi.URLParam(r, "setID")
	w.Header().Set("Content-Type", "application/json")

	response, err := twitchemotesCache.Get(ctx, setID, nil)
	if err != nil {
		var perr *twitchEmotesError
		if errors.As(err, &perr) {
			w.WriteHeader(perr.Status)
			data, err := json.Marshal(&TwitchEmotesErrorResponse{
				Status:  perr.Status,
				Message: perr.Error(),
			})
			if err != nil {
				log.Println("Error marshalling twitch emotes error response:", err)
			}
			_, err = w.Write(data)
			if err != nil {
				log.Println("Error writing response:", err)
			}
		} else {
			fmt.Println("Unknown error in twitchemotes set handler:", err)
		}

		return
	}

	_, err = w.Write(response)
	if err != nil {
		log.Println("Error writing response:", err)
	}
}

type TwitchemotesLoader struct {
}

func (l *TwitchemotesLoader) Load(ctx context.Context, setID string, r *http.Request) ([]byte, time.Duration, error) {
	if helixAPI == nil || helixUsernameCache == nil {
		return nil, 0, &twitchEmotesError{
			UnderlyingError: errUnableToHandle,
			Status:          500,
		}
	}

	params := &helix.GetEmoteSetsParams{
		EmoteSetIDs: []string{
			setID,
		},
	}
	resp, err := helixAPI.GetEmoteSets(params)
	if err != nil {
		return nil, 0, &twitchEmotesError{
			UnderlyingError: err,
			Status:          500,
		}
	}

	if len(resp.Data.Emotes) == 0 {
		return nil, 0, &twitchEmotesError{
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
		if usernameBytes, err := helixUsernameCache.Get(ctx, emote.OwnerID, nil); err != nil {
			return nil, 0, &twitchEmotesError{
				UnderlyingError: err,
				Status:          404,
			}
		} else {
			username = string(usernameBytes)
		}
	}

	emoteSet := EmoteSet{
		ChannelName: username,
		ChannelID:   emote.OwnerID,
		Type:        emote.EmoteType,
	}

	return utils.MarshalNoDur(&emoteSet)
}

// Initialize servers the /twitchemotes/set/{setID} route
// In newer versions of Chatterino this data is fetched client-side instead.
// To support older versions of Chattterino that relied on this API we will keep this API functional for some time longer.
func Initialize(ctx context.Context, cfg config.APIConfig, pool db.Pool, router *chi.Mux, helixClient *helix.Client, helisUsernameCache cache.Cache) error {
	loader := &TwitchemotesLoader{}
	helixAPI = helixClient
	twitchemotesCache := cache.NewPostgreSQLCache(ctx, cfg, pool, "twitchemotes", loader, time.Duration(30)*time.Minute)

	router.Get("/twitchemotes/set/{setID}", func(w http.ResponseWriter, r *http.Request) {
		setHandler(ctx, helixUsernameCache, twitchemotesCache, w, r)
	})

	return nil
}
