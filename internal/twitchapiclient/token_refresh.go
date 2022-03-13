package twitchapiclient

import (
	"context"
	"time"

	"github.com/Chatterino/api/internal/logger"
	"github.com/nicklaw5/helix"
)

// initAppAccessToken requests and sets app access token to the provided helix.Client
// and initializes a ticker running every 24 Hours which re-requests and sets app access token
func initAppAccessToken(ctx context.Context, helixAPI *helix.Client, tokenFetched chan struct{}) {
	log := logger.FromContext(ctx)

	response, err := helixAPI.RequestAppAccessToken([]string{})

	if err != nil {
		log.Fatalw("[Helix] Error requesting app access token:",
			"error", err,
		)
	}

	log.Debugw("[Helix] Requested access token",
		"status", response.StatusCode,
		"expiresIn", response.Data.ExpiresIn,
	)
	helixAPI.SetAppAccessToken(response.Data.AccessToken)
	close(tokenFetched)

	// initialize the ticker
	ticker := time.NewTicker(24 * time.Hour)

	for range ticker.C {
		response, err := helixAPI.RequestAppAccessToken([]string{})
		if err != nil {
			log.Errorw("[Helix] Failed to re-request app access token from ticker",
				"error", err,
			)
			continue
		}
		log.Debugw("[Helix] Re-requested access token from ticker",
			"status", response.StatusCode,
			"expiresIn", response.Data.ExpiresIn,
		)

		helixAPI.SetAppAccessToken(response.Data.AccessToken)
	}
}
