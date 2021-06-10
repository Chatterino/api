package config

import (
	"fmt"
	"log"
	"reflect"
	//"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	envPrefix = "CHATTERINO_API"
	appName   = "chatterino-api"
)

var (
	Config APIConfig
	v      = viper.New()
)

func init() {
	// Default config
	ref := reflect.ValueOf(defaultConf)

	// TODO: Figure out a better way to set default values - this will not work with nested keys
	for i := 0; i < ref.NumField(); i++ {
		v.SetDefault(ref.Type().Field(i).Tag.Get("mapstructure"), ref.Field(i).Interface())
	}

	// Flags
	pflag.StringP("base_url", "b", "", "Bind address")
	pflag.StringP("bind_address", "l", "", "Base URL (useful if being proxied through something like nginx). Value needs to be full URL up to the application (e.g. https://braize.pajlada.com/chatterino)")
	pflag.Bool("enable_lilliput", true, "enable_lilliput")
	pflag.String("discord_token", "", "discord")
	pflag.String("twitch_client_id", "", "twitch_client_id")
	pflag.String("twitch_client_secret", "", "twitch_client_secret")
	pflag.String("youtube_api_key", "", "youtube_api_key")
	pflag.String("twitter_bearer_token", "", "twitter_bearer_token")
	pflag.String("imgur_client_id", "", "imgur_client_id")
	pflag.String("oembed_facebook_app_id", "", "oembed_facebook_app_id")
	pflag.String("oembed_facebook_app_secret", "", "oembed_facebook_app_secret")
	pflag.String("oembed_providers_path", "", "oembed_providers_path")
	pflag.Parse()
	v.BindPFlags(pflag.CommandLine)

	// File
	v.SetConfigName(appName)
	v.SetConfigType("yaml")
	v.AddConfigPath("/etc/")
	v.AddConfigPath(fmt.Sprintf("$HOME/.config/%s/", appName))
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
	Config = defaultConf
	v.UnmarshalExact(&Config)

	//fmt.Printf("%# v\n", pretty.Formatter(cfg))
}
