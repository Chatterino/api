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

	response, err := helixAPI.GetClips(&helix.ClipsParams{IDs: []string{clipSlug}})

	if err != nil {
		return noTwitchClipWithThisIDFound, nil, cache.NoSpecialDur
	}

	var clipHelix = response.Data.Clips[0]

	var createdData, _ = time.Parse("2006-01-02T15:04:05Z", clipHelix.CreatedAt)

	data := twitchClipsTooltipData{
		Title:       clipHelix.Title,
		AuthorName:  clipHelix.CreatorName,
		ChannelName: clipHelix.BroadcasterName,
		// Duration: // fmt.Sprintf("%g%s", clip.Duration, "s")
		CreationDate: createdData.Format("02 Jan 2006"),
		Views:        humanize.Number(uint64(clipHelix.ViewCount)),
	}

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
