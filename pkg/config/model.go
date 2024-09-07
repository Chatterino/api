package config

import "time"

type APIConfig struct {
	// Core

	BaseURL                  string `mapstructure:"base-url" json:"base-url"`
	BindAddress              string `mapstructure:"bind-address" json:"bind-address"`
	MaxContentLength         uint64 `mapstructure:"max-content-length" json:"max-content-length"`
	EnableAnimatedThumbnails bool   `mapstructure:"enable-animated-thumbnails" json:"enable-animated-thumbnails"`
	MaxThumbnailSize         uint   `mapstructure:"max-thumbnail-size" json:"max-thumbnail-size"`

	BttvEmoteCacheDuration           time.Duration `mapstructure:"bttv-emote-cache-duration" json:"bttv-emote-cache-duration"`
	ThumbnailCacheDuration           time.Duration `mapstructure:"thumbnail-cache-duration" json:"thumbnail-cache-duration"`
	DefaultLinkCacheDuration         time.Duration `mapstructure:"default-link-cache-duration" json:"default-link-cache-duration"`
	DiscordInviteCacheDuration       time.Duration `mapstructure:"discord-invite-cache-duration" json:"discord-invite-cache-duration"`
	FfzEmoteCacheDuration            time.Duration `mapstructure:"ffz-emote-cache-duration" json:"ffz-emote-cache-duration"`
	ImgurCacheDuration               time.Duration `mapstructure:"imgur-cache-duration" json:"imgur-cache-duration"`
	LivestreamfailsClipCacheDuration time.Duration `mapstructure:"livestreamfails-clip-cache-duration" json:"livestreamfails-clip-cache-duration"`
	OembedCacheDuration              time.Duration `mapstructure:"oembed-cache-duration" json:"oembed-cache-duration"`
	SeventvEmoteCacheDuration        time.Duration `mapstructure:"seventv-emote-cache-duration" json:"seventv-emote-cache-duration"`
	SupinicTrackCacheDuration        time.Duration `mapstructure:"supinic-track-cache-duration" json:"supinic-track-cache-duration"`
	TwitchClipCacheDuration          time.Duration `mapstructure:"twitch-clip-cache-duration" json:"twitch-clip-cache-duration"`
	TwitterTweetCacheDuration        time.Duration `mapstructure:"twitter-tweet-cache-duration" json:"twitter-tweet-cache-duration"`
	TwitterUserCacheDuration         time.Duration `mapstructure:"twitter-user-cache-duration" json:"twitter-user-cache-duration"`
	WikipediaArticleCacheDuration    time.Duration `mapstructure:"wikipedia-article-cache-duration" json:"wikipedia-article-cache-duration"`
	YoutubeChannelCacheDuration      time.Duration `mapstructure:"youtube-channel-cache-duration" json:"youtube-channel-cache-duration"`
	YoutubeVideoCacheDuration        time.Duration `mapstructure:"youtube-video-cache-duration" json:"youtube-video-cache-duration"`
	TwitchUsernameCacheDuration      time.Duration `mapstructure:"twitch-username-cache-duration" json:"twitch-user-cache-duration"`

	LogLevel       string `mapstructure:"log-level" json:"log-level"`
	LogDevelopment bool   `mapstructure:"log-development" json:"log-development"`

	DSN string `mapstructure:"dsn" json:"dsn"`

	EnablePrometheus      bool   `mapstructure:"enable-prometheus" json:"enable-prometheus"`
	PrometheusBindAddress string `mapstructure:"prometheus-bind-address" json:"prometheus-bind-address"`

	// Secrets

	DiscordToken            string `mapstructure:"discord-token" json:"discord-token"`
	TwitchClientID          string `mapstructure:"twitch-client-id" json:"twitch-client-id"`
	TwitchClientSecret      string `mapstructure:"twitch-client-secret" json:"twitch-client-secret"`
	YoutubeApiKey           string `mapstructure:"youtube-api-key" json:"youtube-api-key"`
	TwitterBearerToken      string `mapstructure:"twitter-bearer-token" json:"twitter-bearer-token"`
	ImgurClientID           string `mapstructure:"imgur-client-id" json:"imgur-client-id"`
	OembedFacebookAppID     string `mapstructure:"oembed-facebook-app-id" json:"oembed-facebook-app-id"`
	OembedFacebookAppSecret string `mapstructure:"oembed-facebook-app-secret" json:"oembed-facebook-app-secret"`
	OembedProvidersPath     string `mapstructure:"oembed-providers-path" json:"oembed-providers-path"`
}
