package cache

import "context"

type KeyProvider interface {
	// Returns the name of the cache key generated for the query
	CacheKey(ctx context.Context, query string) string
}

type PrefixKeyProvider struct {
	prefix string
}

func NewPrefixKeyProvider(prefix string) *PrefixKeyProvider {
	return &PrefixKeyProvider{
		prefix: prefix,
	}
}

func (p *PrefixKeyProvider) CacheKey(ctx context.Context, query string) string {
	return p.prefix + ":" + query
}
