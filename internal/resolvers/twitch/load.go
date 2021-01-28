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

func load(clipSlug string, r *http.Request) (interface{}, error, time.Duration) {
	log.Println("[TwitchClip] GET", clipSlug)
	// clip, _, err := v5API.GetClip(clipSlug)
	// if err != nil {
	// 	return noTwitchClipWithThisIDFound, nil, cache.NoSpecialDur
	// }

	response, err := helixAPI.GetClips(&helix.ClipsParams{IDs: []string{clipSlug}})
	log.Println("[TwitchClip] 2")

	if err != nil {
		return noTwitchClipWithThisIDFound, nil, cache.NoSpecialDur
	}
	log.Println("[TwitchClip] 3")

	var clipHelix = response.Data.Clips[0]
	log.Println("[TwitchClip] 4")

	data := twitchClipsTooltipData{
		Title:       clipHelix.Title,
		ChannelName: clipHelix.BroadcasterName,
		// Duration: // fmt.Sprintf("%g%s", clip.Duration, "s")
		CreationDate: clipHelix.CreatedAt, // clip.CreatedAt.Format("02 Jan 2006")
		Views:        humanize.Number(uint64(clipHelix.ViewCount)),
	}
	log.Println("[TwitchClip] 5")

	var tooltip bytes.Buffer
	if err := twitchClipsTooltip.Execute(&tooltip, data); err != nil {
		return &resolver.Response{
			Status:  http.StatusInternalServerError,
			Message: "twitch clip template error " + resolver.CleanResponse(err.Error()),
		}, nil, cache.NoSpecialDur
	}

	return &resolver.Response{
		Status:    200,
		Tooltip:   url.PathEscape(tooltip.String()),
		Thumbnail: clipHelix.ThumbnailURL,
	}, nil, cache.NoSpecialDur
}
