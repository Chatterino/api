package supinic

import (
	"errors"
	"html/template"
	"regexp"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
)

const trackListAPIURL = "https://supinic.com/api/track/detail/%d"

var (
	templateSupinicTrack = template.Must(template.New("trackListEntryTooltip").Parse(templateStringSupinicTrack))

	trackListCache = cache.New("supinic_track_list_tracks", load, 1*time.Hour)

	errInvalidTrackPath = errors.New("invalid track list track path")

	// List of hosts that will be checked for track list paths
	trackListDomains = map[string]struct{}{
		"supinic.com": {},
	}

	trackPathRegex = regexp.MustCompile(`/track/detail/([0-9]+)`)
)

func New(cfg config.APIConfig) (resolvers []resolver.CustomURLManager) {
	// Find links matching the Track list link (e.g. https://supinic.com/track/detail/1883)
	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: check,

		Run: run,
	})

	return
}
