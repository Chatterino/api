package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
)

func FormatThumbnailURL(baseURL string, r *http.Request, urlString string) string {
	if baseURL != "" {
		return fmt.Sprintf("%s/thumbnail/%s", strings.TrimSuffix(baseURL, "/"), url.QueryEscape(urlString))
	}

	forwardedProtocol := r.Header.Get("X-Forwarded-Proto")

	scheme := "https://"

	if forwardedProtocol == "https" {
		scheme = "https://"
	} else if forwardedProtocol == "http" {
		scheme = "http://"
	} else if r.TLS == nil {
		scheme = "http://" // https://github.com/golang/go/issues/28940#issuecomment-441749380
	}

	return fmt.Sprintf("%s%s/thumbnail/%s", scheme, r.Host, url.QueryEscape(urlString))
}

func FormatGeneratedThumbnailURL(baseURL string, r *http.Request, urlString string) string {
	if baseURL != "" {
		return fmt.Sprintf(
			"%s/generated/%s", strings.TrimSuffix(baseURL, "/"), url.QueryEscape(urlString),
		)
	}

	forwardedProtocol := r.Header.Get("X-Forwarded-Proto")

	scheme := "https://"

	if forwardedProtocol == "https" {
		scheme = "https://"
	} else if forwardedProtocol == "http" {
		scheme = "http://"
	} else if r.TLS == nil {
		scheme = "http://" // https://github.com/golang/go/issues/28940#issuecomment-441749380
	}

	return fmt.Sprintf("%s%s/generated/%s", scheme, r.Host, url.QueryEscape(urlString))
}

func UnescapeURLArgument(r *http.Request, key string) (string, error) {
	escapedURL := chi.URLParam(r, key)
	url, err := url.PathUnescape(escapedURL)
	if err != nil {
		return "", err
	}

	return url, nil
}

// IsSubdomainOf checks whether `url` is a subdomain of `parent`
func IsSubdomainOf(url *url.URL, parents ...string) bool {
	// We use Hostname() as that strips possible port numbers (relevant for the suffix check)
	hostname := url.Hostname()

	for _, parent := range parents {
		same := (hostname == parent)
		trueSub := strings.HasSuffix(hostname, "."+parent)

		if same || trueSub {
			return true
		}
	}

	return false
}

// IsDomains checks whether `url`s domain matches any of the given domains exactly (non-case sensitive)
// The `domains` map should only contain fully lowercased domains
func IsDomains(url *url.URL, domains map[string]struct{}) bool {
	host := strings.ToLower(url.Hostname())
	_, ok := domains[host]
	return ok
}

// IsDomain checks whether `url`s domain matches the given domain exactly (non-case sensitive)
// The `domain` string must be fully lowercased
func IsDomain(url *url.URL, domain string) bool {
	host := strings.ToLower(url.Hostname())
	return host == domain
}

func MustParseURL(urlString string) *url.URL {
	u, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}
	return u
}
