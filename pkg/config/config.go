package config

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	envPrefix = "CHATTERINO_API"
	appName   = "chatterino-api"
)

var (
	Cfg APIConfig
	v   = viper.New()
)

func init() {
	// Default config
	ref := reflect.ValueOf(defaultConf)

	// TODO: Figure out a better way to set default values - this will not work with nested keys
	for i := 0; i < ref.NumField(); i++ {
		v.SetDefault(ref.Type().Field(i).Tag.Get("mapstructure"), ref.Field(i).Interface())
	}

	// Flags
	pflag.StringP("base_url", "b", "", "Address to which API will bind and start listening")
	pflag.StringP("bind_address", "l", "", "Base URL (useful if being proxied through something like nginx). Value needs to be full URL up to the application (e.g. https://braize.pajlada.com/chatterino)")
	pflag.Uint64("max_content_length", 5*1024*1024, "Max content size (in bytes) - requests with body bigger than this value will be skipped")
	pflag.Bool("enable_lilliput", true, "When enabled, will attempt to use lilliput library for building animated thumbnails. Could increase memory usage by a lot.")
	pflag.String("discord_token", "", "Discord token")
	pflag.String("twitch_client_id", "", "Twitch client ID")
	pflag.String("twitch_client_secret", "", "Twitch client secret")
	pflag.String("youtube_api_key", "", "YouTube API key")
	pflag.String("twitter_bearer_token", "", "Twitter bearer token")
	pflag.String("imgur_client_id", "", "Imgur client ID")
	pflag.String("oembed_facebook_app_id", "", "oEmbed Facebook app ID")
	pflag.String("oembed_facebook_app_secret", "", "oEmbed Facebook app secret")
	pflag.String("oembed_providers_path", "./providers.json", "Path to a json file containing supported oEmbed resolvers")
	pflag.Parse()
	v.BindPFlags(pflag.CommandLine)

	// figure out XDG_DATA_CONFIG to be compliant with the standard
	xdgConfigHome, exists := os.LookupEnv("XDG_CONFIG_HOME")
	if !exists || xdgConfigHome == "" {
		xdgConfigHome = fmt.Sprintf("$HOME/.config/%s/", appName)
	}

	// File
	v.SetConfigName(appName)
	v.SetConfigType("yaml")
	v.AddConfigPath("/etc/")
	v.AddConfigPath(xdgConfigHome)
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found")
		} else {
			log.Println("Fatal error encountered while reading config file")
			panic(err)
		}
	} else {
		v.MergeInConfig()
	}

	// Environment
	//V.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // will be useful once we have nested keys
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()

	// Print config
	Cfg = defaultConf
	v.UnmarshalExact(&Cfg)

	fmt.Printf("%# v\n", Cfg)
}
