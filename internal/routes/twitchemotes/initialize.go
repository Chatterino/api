package twitchemotes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Chatterino/api/internal/db"
	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/go-chi/chi/v5"
	"github.com/nicklaw5/helix"
)

func setHandler(ctx context.Context, twitchemotesCache cache.Cache, w http.ResponseWriter, r *http.Request) {
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

	// TODO: STATUS CODE
	_, err = w.Write(response.Payload)
	if err != nil {
		log.Println("Error writing response:", err)
	}
}

// Initialize servers the /twitchemotes/set/{setID} route
// In newer versions of Chatterino this data is fetched client-side instead.
// To support older versions of Chattterino that relied on this API we will keep this API functional for some time longer.
func Initialize(ctx context.Context, cfg config.APIConfig, pool db.Pool, router *chi.Mux, helixClient *helix.Client, helixUsernameCache cache.Cache) error {
	loader := &TwitchemotesLoader{
		helixAPI:           helixClient,
		helixUsernameCache: helixUsernameCache,
	}
	twitchemotesCache := cache.NewPostgreSQLCache(
		ctx, cfg, pool, cache.NewPrefixKeyProvider("twitchemotes"), loader,
		cfg.TwitchEmoteCacheDuration,
	)

	router.Get("/twitchemotes/set/{setID}", func(w http.ResponseWriter, r *http.Request) {
		setHandler(ctx, twitchemotesCache, w, r)
	})

	return nil
}
