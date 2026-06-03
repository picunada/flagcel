package flagcel

import (
	"fmt"
	"sync"

	"github.com/picunada/flagcel/evalcore"
)

type definitionsCache struct {
	mu        sync.RWMutex
	evaluator *evalcore.Evaluator
	etag      string
	lastErr   error
	ready     bool
}

func (c *definitionsCache) snapshot() (*evalcore.Evaluator, string, error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.evaluator, c.etag, c.lastErr, c.ready
}

func (c *definitionsCache) etagValue() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.etag
}

func (c *definitionsCache) markUnchanged(etag string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if etag != "" {
		c.etag = etag
	}
	c.lastErr = nil
	c.ready = true
}

func (c *definitionsCache) store(defs evalcore.Definitions, etag string) error {
	evaluator, err := evalcore.Load(defs)
	if err != nil {
		c.markError(fmt.Errorf("flagcel: compile definitions: %w", err))
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.evaluator = evaluator
	c.etag = etag
	c.lastErr = nil
	c.ready = true
	return nil
}

func (c *definitionsCache) markError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastErr = err
}
