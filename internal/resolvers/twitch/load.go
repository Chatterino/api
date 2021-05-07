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
		return noTwitchClipWithThisIDFound, cache.NoSpecialDur, nil
	}

	var clipHelix = response.Data.Clips[0]

	data := twitchClipsTooltipData{
		Title:        clipHelix.Title,
		AuthorName:   clipHelix.CreatorName,
		ChannelName:  clipHelix.BroadcasterName,
		Duration:     humanize.DurationSeconds(time.Duration(clipHelix.Duration) * time.Second),
		CreationDate: humanize.CreationDateRFC3339(clipHelix.CreatedAt),
		Views:        humanize.Number(uint64(clipHelix.ViewCount)),
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
		Thumbnail: clipHelix.ThumbnailURL,
	}, cache.NoSpecialDur, nil
}
