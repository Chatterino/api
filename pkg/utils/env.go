package utils

import "os"

const (
	CHATTERINO_ENV_PREFIX = "CHATTERINO_API_"
)

// LookupEnv is a thin wrapper of os.LookupEnv, except it prefixes the given envSuffix with the standard environment variable prefix (i.e. CHATTERINO_API_)
// Calling LookupEnv("YOUTUBE_API_KEY") would return the value for the environment variable CHATTERINO_API_YOUTUBE_API_KEY
func LookupEnv(envSuffix string) (value string, exists bool) {
	envKey := CHATTERINO_ENV_PREFIX + envSuffix
	return os.LookupEnv(envKey)
}
