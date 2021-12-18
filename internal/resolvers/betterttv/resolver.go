package betterttv

import (
	"errors"
	"html/template"
	"regexp"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/config"
	"github.com/Chatterino/api/pkg/resolver"
)

const thumbnailFormat = "https://cdn.betterttv.net/emote/%s/3x"

var (
	emoteAPIURL = "https://api.betterttv.net/3/emotes/%s"

	errInvalidBTTVEmotePath = errors.New("invalid BetterTTV emote path")

	// BetterTTV hosts we're doing our smart things on
	domains = map[string]struct{}{
		"betterttv.com":     {},
		"www.betterttv.com": {},
	}

	emotePathRegex = regexp.MustCompile(`/emotes/([a-f0-9]+)`)

	emoteCache = cache.New("betterttv_emotes", load, 1*time.Hour)

	templateBTTVEmote = template.Must(template.New("betterttvEmoteTooltip").Parse(templateStringBTTVEmote))
)

func New(cfg config.APIConfig) (resolvers []resolver.CustomURLManager) {
	// Find links matching the BetterTTV direct emote link (e.g. https://betterttv.com/emotes/566ca06065dbbdab32ec054e)
	resolvers = append(resolvers, resolver.CustomURLManager{
		Check: check,

		Run: run,
	})

	return
}
