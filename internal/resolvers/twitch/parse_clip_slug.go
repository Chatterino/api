package twitch

import "net/url"

func parseClipSlug(url *url.URL) (string, error) {
	matches := clipSlugRegex.FindStringSubmatch(url.Path)

	if len(matches) != 3 {
		return "", errInvalidTwitchClip
	}

	return matches[2], nil
}
