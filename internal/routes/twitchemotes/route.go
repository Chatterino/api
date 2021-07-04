package twitchemotes

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/nicklaw5/helix"
)

var (
	errInvalidEmoteID = errors.New("invalid emote id")
	errUnableToHandle = errors.New("unable to handle twitchemotes requests")

	twitchemotesCache = cache.New("twitchemotes", doTwitchemotesRequest, time.Duration(30)*time.Minute)

	helixAPI           *helix.Client
	helixUsernameCache *cache.Cache
)

type TwitchEmotesError struct {
	Status int
	Error  error
}

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

func doTwitchemotesRequest(setID string, r *http.Request) (interface{}, time.Duration, error) {
	if helixAPI == nil || helixUsernameCache == nil {
		return &TwitchEmotesError{
			Error:  errUnableToHandle,
			Status: 500,
		}, 0, nil
	}

	params := &helix.GetEmoteSetsParams{
		EmoteSetIDs: []string{
			setID,
		},
	}
	resp, err := helixAPI.GetEmoteSets(params)
	if err != nil {
		return &TwitchEmotesError{
			Error:  err,
			Status: 500,
		}, 0, nil
	}

	if len(resp.Data.Emotes) == 0 {
		return &TwitchEmotesError{
			Error:  errInvalidEmoteID,
			Status: 404,
		}, 0, nil
	}

	emote := resp.Data.Emotes[0]

	var ok bool
	var username string

	// For Emote Sets 0 (global) and 19194 (prime emotes), the Owner ID returns 0
	// 0 is not a valid Twitch User ID, so hardcode the username to Twitch
	if emote.OwnerID == "0" {
		username = "Twitch"
	} else {
		// Load username from Helix
		if username, ok = helixUsernameCache.Get(emote.OwnerID, nil).(string); !ok {
			return &TwitchEmotesError{
				Error:  errInvalidEmoteID,
				Status: 404,
			}, 0, nil
		}
	}

	emoteSet := EmoteSet{
		ChannelName: username,
		ChannelID:   emote.OwnerID,
		Type:        emote.EmoteType,
	}

	return utils.MarshalNoDur(&emoteSet)
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	setID := chi.URLParam(r, "setID")
	w.Header().Set("Content-Type", "application/json")

	response := twitchemotesCache.Get(setID, nil)

	switch v := response.(type) {
	case []byte:
		_, err := w.Write(v)
		if err != nil {
			log.Println("Error writing response:", err)
		}

	case *TwitchEmotesError:
		w.WriteHeader(v.Status)
		data, err := json.Marshal(&TwitchEmotesErrorResponse{
			Status:  v.Status,
			Message: v.Error.Error(),
		})
		if err != nil {
			log.Println("Error marshalling twitch emotes error response:", err)
		}
		_, err = w.Write(data)
		if err != nil {
			log.Println("Error writing response:", err)
		}
	}
}

// Initialize servers the /twitchemotes/set/{setID} route
// In newer versions of Chatterino this data is fetched client-side instead.
// To support older versions of Chattterino that relied on this API we will keep this API functional for some time longer.
func Initialize(router *chi.Mux, helixClient *helix.Client, usernameCache *cache.Cache) error {
	helixAPI = helixClient
	helixUsernameCache = usernameCache

	router.Get("/twitchemotes/set/{setID}", setHandler)

	return nil
}
