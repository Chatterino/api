package twitchapiclient

import (
	"errors"

	"github.com/Chatterino/api/pkg/config"
	"github.com/nicklaw5/helix"
)

// New returns a helix.Client that has requested an AppAccessToken and will keep it refreshed every 24h
func New(cfg config.APIConfig) (*helix.Client, error) {
	if cfg.TwitchClientID == "" {
		return nil, errors.New("twitch-client-id is missing, can't make Twitch requests")
	}

	if cfg.TwitchClientSecret == "" {
		return nil, errors.New("twitch-client-secret is missing, can't make Twitch requests")
	}

	apiClient, err := helix.NewClient(&helix.Options{
		ClientID:     cfg.TwitchClientID,
		ClientSecret: cfg.TwitchClientSecret,
	})

	if err != nil {
		return nil, err
	}

	waitForFirstAppAccessToken := make(chan struct{})

	// Initialize methods responsible for refreshing oauth
	go initAppAccessToken(apiClient, waitForFirstAppAccessToken)
	<-waitForFirstAppAccessToken

	return apiClient, nil
}
