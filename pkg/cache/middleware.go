package cache

import (
	"fmt"
	"net/http"
	"time"
)

// MaxAgeHeaders adds the Cache-Control: max-age=$TTL header to every response
func MaxAgeHeaders(ttl time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%.0f", ttl.Seconds()))
			next.ServeHTTP(w, r)
		})
	}
}
