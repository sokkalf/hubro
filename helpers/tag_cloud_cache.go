package helpers

import (
	"html/template"
	"sync"

	"github.com/sokkalf/hubro/index"
)

type tagCloudCache struct {
	mu  sync.RWMutex
	m   map[*index.Index]*template.HTML
}

var globalCache = &tagCloudCache{
	m: make(map[*index.Index]*template.HTML),
}

func (c *tagCloudCache) get(idx *index.Index) (*template.HTML, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	t, ok := c.m[idx]
	return t, ok
}

func (c *tagCloudCache) set(idx *index.Index, t *template.HTML) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[idx] = t
}

func (c *tagCloudCache) reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m = make(map[*index.Index]*template.HTML)
}
