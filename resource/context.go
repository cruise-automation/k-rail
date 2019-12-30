package resource

import "context"

type contextKey int // local to the resource module

const (
	ctxKeyCache contextKey = iota
)
const (
	cacheKeyPod = iota
	cacheKeyPodExec
	cacheKeyIngress
)

// WithResourceCache adds a resource cache to the context returned.
func WithResourceCache(ctx context.Context) context.Context {
	c := make(map[int]interface{}, 1)
	return context.WithValue(ctx, ctxKeyCache, c)
}

// GetResourceCache returns the cache from the context. Result will return nil when none exists.
func GetResourceCache(ctx context.Context) map[int]interface{} {
	c := ctx.Value(ctxKeyCache)
	if c == nil {
		return nil
	}
	return c.(map[int]interface{})
}
