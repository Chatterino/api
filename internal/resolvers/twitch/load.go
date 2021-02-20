package twitch

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
)

func load(clipSlug string, r *http.Request) (interface{}, time.Duration, error) {
	log.Println("[TwitchClip] GET", clipSlug)
	clip, _, err := v5API.GetClip(clipSlug)
	if err != nil {
		return noTwitchClipWithThisIDFound, cache.NoSpecialDur, nil
	}

	data := twitchClipsTooltipData{
		Title:        clip.Title,
		AuthorName:   clip.Curator.DisplayName,
		ChannelName:  clip.Broadcaster.DisplayName,
		Duration:     fmt.Sprintf("%g%s", clip.Duration, "s"),
		CreationDate: clip.CreatedAt.Format("02 Jan 2006"),
		Views:        humanize.Number(uint64(clip.Views)),
	}

	var tooltip bytes.Buffer
	if err := twitchClipsTooltip.Execute(&tooltip, data); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "twitch clip template error " + resolver.CleanResponse(err.Error()),
		}, cache.NoSpecialDur, nil
	}

	return &resolver.Response{
		Status:    200,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: clip.Thumbnails.Medium,
	}, cache.NoSpecialDur, nil
}
