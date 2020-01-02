package resource

import (
	"context"
	"sync"
)

type contextKey int // local to the resource module

const (
	ctxKeyCache contextKey = iota
)
const (
	cacheKeyPod = iota
	cacheKeyPodExec
	cacheKeyIngress
)

type cache struct {
	l sync.Mutex
	m map[int]interface{}
}

// getOrSet returns the cached value for the given key. When none exists the passed function is called once to create
// the initial value. When cache is nil no caching happens and the create function is always called.
// Calls are executed thread safe.
func (c *cache) getOrSet(cacheKey int, f func() interface{}) interface{} {
	if c == nil {
		return f()
	}
	c.l.Lock()
	defer c.l.Unlock()
	if p, ok := c.m[cacheKey]; ok {
		return p
	}
	v := f()
	c.m[cacheKey] = v
	return v
}

// WithResourceCache adds a resource cache to the context returned.
func WithResourceCache(ctx context.Context) context.Context {
	c := &cache{m: make(map[int]interface{}, 1)}
	return context.WithValue(ctx, ctxKeyCache, c)
}

// GetResourceCache returns the cache from the context. Result will return nil when none exists.
func GetResourceCache(ctx context.Context) *cache {
	c := ctx.Value(ctxKeyCache)
	if c == nil {
		return nil
	}
	return c.(*cache)
}
