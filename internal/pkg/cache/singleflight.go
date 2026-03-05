package cache

import (
	"context"
	"sync"
	"time"
)

// Call represents an in-flight or completed singleflight.Do call
type Call struct {
	wg  sync.WaitGroup
	val any
	err error
}

// Group represents a class of work and forms a namespace in which
// units of work can be executed with duplicate suppression.
type Group struct {
	mu sync.Mutex
	m  map[string]*Call
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*Call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(Call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}

// DoWithContext executes the function with context support
func (g *Group) DoWithContext(ctx context.Context, key string, fn func(context.Context) (any, error)) (any, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*Call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()

		// Wait for the call to complete or context to be done
		done := make(chan struct{})
		go func() {
			c.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			return c.val, c.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	c := new(Call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	// Execute the function
	c.val, c.err = fn(ctx)
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}

// Forget tells the singleflight to forget about a key
func (g *Group) Forget(key string) {
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
}

// CacheWithSingleflight wraps a cache get operation with singleflight
type CacheWithSingleflight struct {
	group *Group
	ttl   time.Duration
}

// NewCacheWithSingleflight creates a new cache wrapper with singleflight
func NewCacheWithSingleflight(ttl time.Duration) *CacheWithSingleflight {
	return &CacheWithSingleflight{
		group: &Group{},
		ttl:   ttl,
	}
}

// Get retrieves a value from cache or executes the loader function
func (c *CacheWithSingleflight) Get(ctx context.Context, key string, loader func(context.Context) (any, error)) (any, error) {
	return c.group.DoWithContext(ctx, key, loader)
}
