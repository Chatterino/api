package twitch

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/Chatterino/api/pkg/cache"
	"github.com/Chatterino/api/pkg/humanize"
	"github.com/Chatterino/api/pkg/resolver"
	"github.com/nicklaw5/helix"
)

func load(clipSlug string, r *http.Request) (interface{}, time.Duration, error) {
	log.Println("[TwitchClip] GET", clipSlug)

	response, err := helixAPI.GetClips(&helix.ClipsParams{IDs: []string{clipSlug}})
	if err != nil {
		log.Println("[TwitchClip] Error getting clip", clipSlug, ":", err.Error())

		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "An internal error occured while fetching the Twitch clip",
		}, cache.NoSpecialDur, nil
	}

	if len(response.Data.Clips) != 1 {
		return noTwitchClipWithThisIDFound, cache.NoSpecialDur, nil
	}

	var clip = response.Data.Clips[0]

	data := twitchClipsTooltipData{
		Title:        clip.Title,
		AuthorName:   clip.CreatorName,
		ChannelName:  clip.BroadcasterName,
		Duration:     humanize.DurationSeconds(time.Duration(clip.Duration) * time.Second),
		CreationDate: humanize.CreationDateRFC3339(clip.CreatedAt),
		Views:        humanize.Number(uint64(clip.ViewCount)),
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
		Thumbnail: clip.ThumbnailURL,
	}, cache.NoSpecialDur, nil
}
