package imgur

import (
	"net/url"
	"strings"
)

func check(url *url.URL) bool {
	// TODO: Make a shared helper function that does basically this.
	// helpers.IsSubdomainOf(url.Host, "imgur.com")
	isImgur := strings.HasSuffix(url.Host, ".imgur.com") || url.Host == "imgur.com"

	if !isImgur {
		return false
	}

	return true
}
