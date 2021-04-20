package imgur

import (
	"net/url"

	"github.com/Chatterino/api/pkg/utils"
)

func check(url *url.URL) bool {
	return utils.IsSubdomainOf(url, "imgur.com")
}
