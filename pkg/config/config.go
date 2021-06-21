package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	envPrefix  = "CHATTERINO_API"
	appName    = "chatterino-api"
	configName = "config"
)

// readFromPath reads the config values from the given path (i.e. path/${configName}.yaml) and returns its values as a map.
// This allows us to use mergeConfig cleanly
func readFromPath(path string) (values map[string]interface{}, err error) {
	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType("yaml")
	v.AddConfigPath(path)

	if err = v.ReadInConfig(); err != nil {
		notFoundError := &viper.ConfigFileNotFoundError{}
		if errors.As(err, notFoundError) {
			err = nil
			return
		}
		return
	}

	v.Unmarshal(&values)

	return
}

// mergeConfig uses viper.MergeConfigMap to read config values in the unix
// standard, so you start furthest down with reading the system config file,
// merge those values into the main config map, then read the home directory
// config files, and merge any set values from there, and lastly the config
// file in the cwd and merge those in. If a value is not set in the cwd config
// file, but one is set in the system config file then the system config file
// value will be used
func mergeConfig(v *viper.Viper, configPaths []string) {
	for _, configPath := range configPaths {
		if configMap, err := readFromPath(configPath); err != nil {
			fmt.Printf("Error reading config file from %s.yaml: %s\n", filepath.Join(configPath, configName), err)
			return
		} else {
			v.MergeConfigMap(configMap)
		}
	}
}

func init() {
	// Define command-line flags and default values
	pflag.StringP("base-url", "b", "", "Base URL to which clients will make their requests. Useful if the API is proxied through reverse proxy like nginx. Value needs to contain full URL with protocol scheme, e.g. https://braize.pajlada.com/chatterino")
	pflag.StringP("bind-address", "l", ":1234", "Address to which API will bind and start listening on")
	pflag.Uint64("max-content-length", 5*1024*1024, "Max content size in bytes - requests with body bigger than this value will be skipped")
	pflag.Bool("enable-lilliput", true, "When enabled, will attempt to use lilliput library for building animated thumbnails. Can increase memory usage by a lot")
	pflag.String("discord-token", "", "Discord token")
	pflag.String("twitch-client-id", "", "Twitch client ID")
	pflag.String("twitch-client-secret", "", "Twitch client secret")
	pflag.String("youtube-api-key", "", "YouTube API key")
	pflag.String("twitter-bearer-token", "", "Twitter bearer token")
	pflag.String("imgur-client-id", "", "Imgur client ID")
	pflag.String("oembed-facebook-app-id", "", "oEmbed Facebook app ID")
	pflag.String("oembed-facebook-app-secret", "", "oEmbed Facebook app secret")
	pflag.String("oembed-providers-path", "./data/oembed/providers.json", "Path to a json file containing supported oEmbed resolvers")
	pflag.Parse()
}

func New() (cfg APIConfig) {
	v := viper.New()

	v.BindPFlags(pflag.CommandLine)

	// figure out XDG_DATA_CONFIG to be compliant with the standard
	xdgConfigHome, exists := os.LookupEnv("XDG_CONFIG_HOME")
	if !exists || xdgConfigHome == "" {
		// on Windows, we use appdata since that's the closest equivalent
		if runtime.GOOS == "windows" {
			xdgConfigHome = "$APPDATA"
		} else {
			xdgConfigHome = filepath.Join("$HOME", ".config")
		}
	}

	// config paths to read from, in order of least importance
	var configPaths []string
	if runtime.GOOS != "windows" {
		configPaths = append(configPaths, filepath.Join("/etc", appName))
	}
	configPaths = append(configPaths, filepath.Join(xdgConfigHome, appName))
	configPaths = append(configPaths, ".")

	mergeConfig(v, configPaths)

	// Environment
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()

	v.UnmarshalExact(&cfg)

	//fmt.Printf("%# v\n", cfg) // uncomment for debugging purposes
	return
}
