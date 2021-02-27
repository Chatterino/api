package wikipedia

import (
	"net/url"
	"strings"

	"github.com/Chatterino/api/pkg/utils"
)

func check(url *url.URL) bool {
	isWikipedia := utils.IsSubdomainOf(url, "wikipedia.org")
	isWikiArticle := strings.HasPrefix(url.Path, "/wiki/")

	return isWikipedia && isWikiArticle
}
