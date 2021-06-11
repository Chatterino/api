package config

type APIConfig struct {
	// Core

	BaseURL          string `mapstructure:"base_url" json:"base_url"`
	BindAddress      string `mapstructure:"bind_address" json:"bind_address"`
	MaxContentLength uint64 `mapstructure:"max_content_length" json:"max_content_length"`
	EnableLilliput   bool   `mapstructure:"enable_lilliput" json:"enable_lilliput"`

	// Secrets

	DiscordToken            string `mapstructure:"discord_token" json:"discord_token"`
	TwitchClientID          string `mapstructure:"twitch_client_id" json:"twitch_client_id"`
	TwitchClientSecret      string `mapstructure:"twitch_client_secret" json:"twitch_client_secret"`
	YoutubeApiKey           string `mapstructure:"youtube_api_key" json:"youtube_api_key"`
	TwitterBearerToken      string `mapstructure:"twitter_bearer_token" json:"twitter_bearer_token"`
	ImgurClientID           string `mapstructure:"imgur_client_id" json:"imgur_client_id"`
	OembedFacebookAppID     string `mapstructure:"oembed_facebook_app_id" json:"oembed_facebook_app_id"`
	OembedFacebookAppSecret string `mapstructure:"oembed_facebook_app_secret" json:"oembed_facebook_app_secret"`
	OembedProvidersPath     string `mapstructure:"oembed_providers_path" json:"oembed_providers_path"`
}
